/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package feedback

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

var feedbackCacheMutex sync.Mutex

// APNS string representation
const APNS = "apns"

// GCM string representation
const GCM = "gcm"

// Handler is a feedback handler
type Handler struct {
	Config            *viper.Viper
	pendingMessagesWG *sync.WaitGroup
	FeedbackCache     map[string]map[string]int
	FlushInterval     time.Duration
	MarathonDB        *extensions.PGClient
	Logger            zap.Logger
	run               bool
}

// Message is a struct that will decode a apns or gcm feedback message
type Message struct {
	From             string                 `json:"from"`
	MessageID        string                 `json:"message_id"`
	MessageType      string                 `json:"message_type"`
	Error            string                 `json:"error"`
	ErrorDescription string                 `json:"error_description"`
	DeviceToken      string                 `json:"DeviceToken"`
	ID               string                 `json:"id"`
	Err              map[string]interface{} `json:"Err"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewHandler creates a new instance of feedback.Handler
func NewHandler(config *viper.Viper, logger zap.Logger, pendingMessagesWG *sync.WaitGroup, DBOrNil ...*extensions.PGClient) (*Handler, error) {
	h := &Handler{
		Config:            config,
		Logger:            logger,
		pendingMessagesWG: pendingMessagesWG,
		FeedbackCache:     map[string]map[string]int{},
	}
	if len(DBOrNil) > 0 {
		h.configure(DBOrNil[0])
		return h, nil
	}
	err := h.configure()
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handler) loadConfigurationDefaults() {
	h.Config.SetDefault("feedbackListener.flushInterval", 5000)
}

func (h *Handler) configure(DBOrNil ...*extensions.PGClient) error {
	h.loadConfigurationDefaults()
	interval := h.Config.GetInt("feedbackListener.flushInterval")
	h.FlushInterval = time.Duration(interval) * time.Millisecond
	if len(DBOrNil) > 0 {
		h.MarathonDB = DBOrNil[0]
		return nil
	}
	marathonDB, err := extensions.NewPGClient("db", h.Config, h.Logger)
	if err != nil {
		return err
	}
	h.MarathonDB = marathonDB
	return nil
}

func (h *Handler) feedbackService(msg *Message) string {
	if len(msg.MessageID) > 0 {
		return GCM
	}
	return APNS
}

func (h *Handler) handleSuccessMessage(jobID string) {
	feedbackCacheMutex.Lock()
	if _, ok := h.FeedbackCache[jobID]; ok {
		h.FeedbackCache[jobID]["ack"]++
	} else {
		h.FeedbackCache[jobID] = map[string]int{
			"ack": 1,
		}
	}
	feedbackCacheMutex.Unlock()
}

func (h *Handler) handleErrorMessage(jobID string, err string) {
	feedbackCacheMutex.Lock()
	if _, ok := h.FeedbackCache[jobID]; ok {
		h.FeedbackCache[jobID][err]++
	} else {
		h.FeedbackCache[jobID] = map[string]int{
			err: 1,
		}
	}
	feedbackCacheMutex.Unlock()
}

func (h *Handler) handleMessage(msg []byte) {
	defer func() {
		if h.pendingMessagesWG != nil {
			h.pendingMessagesWG.Done()
		}
	}()
	l := h.Logger.With(
		zap.String("method", "feedback.handler.handleMessage"),
	)

	var message Message
	err := json.Unmarshal(msg, &message)
	service := h.feedbackService(&message)
	if err != nil {
		//TODO add sentry
		l.Error("error handling message", zap.Error(err))
	}
	//TODO too many logging, zap has leaks
	log.D(l, "new message", func(cm log.CM) {
		cm.Write(
			zap.String("service", service),
			zap.Object("message", message),
		)
	})

	if message.Metadata == nil || len(message.Metadata) == 0 {
		return
	}

	if message.Metadata["jobId"] == nil || len(message.Metadata["jobId"].(string)) == 0 {
		return
	}

	if len(message.Error) == 0 && (message.Err == nil || len(message.Err) == 0) {
		h.handleSuccessMessage(message.Metadata["jobId"].(string))
	} else {
		if service == APNS {
			h.handleErrorMessage(message.Metadata["jobId"].(string), message.Err["Key"].(string))
		} else if service == GCM {
			h.handleErrorMessage(message.Metadata["jobId"].(string), message.Error)
		}
	}

}

func (h *Handler) generatePGIncrJSON(jobID string, values map[string]int) string {
	q := "UPDATE jobs SET feedbacks = feedbacks || CONCAT('{"
	m := []string{}
	for k, v := range values {
		m = append(m, fmt.Sprintf(`"%s":',COALESCE(feedbacks->>'%s','0')::int + %d,'`, k, k, v))
	}
	joinedModifiers := strings.Join(m, ",")
	endQ := fmt.Sprintf(`}')::jsonb WHERE id = '%s';`, jobID)
	return fmt.Sprintf("%s%s%s", q, joinedModifiers, endQ)
}

func (h *Handler) flushFeedbacks() {
	ticker := time.NewTicker(h.FlushInterval)
	for range ticker.C {
		numFeedbacks := len(h.FeedbackCache)
		if numFeedbacks > 0 {
			h.Logger.Info("flushing feedbacks", zap.Int("feedbacks", numFeedbacks))
		} else {
			h.Logger.Debug("no feedbacks to flush")
		}
		feedbackCacheMutex.Lock()
		for k, v := range h.FeedbackCache {
			query := h.generatePGIncrJSON(k, v)
			results, err := h.MarathonDB.DB.ExecOne(query)
			if err != nil {
				h.Logger.Error("error updating feedbacks table", zap.Error(err))
			} else {
				h.Logger.Debug("successfully updated rows", zap.Int("rows affected", results.RowsAffected()))
			}
			delete(h.FeedbackCache, k)
		}
		feedbackCacheMutex.Unlock()
	}
}

// HandleMessages get messages from msgChan
func (h *Handler) HandleMessages(msgChan *chan []byte) {
	h.run = true
	go h.flushFeedbacks()
	for h.run == true {
		select {
		case message := <-*msgChan:
			h.handleMessage(message)
		}
	}
}
