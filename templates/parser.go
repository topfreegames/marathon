package templates

import (
	"encoding/json"
	"fmt"

	"github.com/uber-go/zap"

	"git.topfreegames.com/topfreegames/marathon/messages"
)

type parseError struct {
	Message string
}

func (e parseError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// Parser is the function responsible to process the messages received from
// kafka, it listens to the channel from kafka and writes to the channel
// listened by TemplateFetcher.
// Multiple Parser instances should be able to run in parallel.
func Parser(l zap.Logger, requireToken bool, inChan <-chan string, outChan chan<- *messages.InputMessage, doneChan <-chan struct{}) {
	l = l.With(zap.Bool("requireToken", requireToken))
	l.Info("Starting parser")
	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		case msg := <-inChan:
			req, err := Parse(l, msg, requireToken)
			if err != nil {
				l.Error("Error parsing message", zap.Error(err))
				continue
			}
			l.Debug("Parsed message", zap.String("input", msg))
			outChan <- req
		}
	}
}

// Parse parses the message string received, it expects to be in JSON format,
// as defined by the RequestMessage struct
func Parse(l zap.Logger, msg string, requireToken bool) (*messages.InputMessage, error) {
	l = l.With(zap.String("msg", msg), zap.Bool("requireToken", requireToken))
	msgObj := messages.NewInputMessage()
	err := json.Unmarshal([]byte(msg), msgObj)
	if err != nil {
		e := parseError{fmt.Sprintf("Error parsing JSON: %+v, the message was: %s", err, msg)}
		return msgObj, e
	}

	e := parseError{}
	// All fields should be set, except by either Template & Params or Message
	if msgObj.App == "" || msgObj.Service == "" || (requireToken && msgObj.Token == "") {
		l.Error(
			"One of the mandatory fields is missing",
			zap.String("app", msgObj.App),
			zap.String("token", msgObj.Token),
			zap.String("service", msgObj.Service),
		)
		e = parseError{fmt.Sprintf(
			"One of the mandatory fields is missing app=%s, token=%s, service=%s", msgObj.App, msgObj.Token, msgObj.Service),
		}
	}
	// Either Template & Params should be defined or Message should be defined
	// Not both at the same time
	if msgObj.Template != "" && msgObj.Params == nil {
		l.Error(
			"Template defined, but not Params",
			zap.String("template", msgObj.Template),
			zap.Object("params", msgObj.Params),
		)
		e = parseError{"Template defined, but not Params"}
	}
	if msgObj.Template != "" && (msgObj.Message != nil && len(msgObj.Message) > 0) {
		l.Error(
			"Both Template and Message defined",
			zap.String("template", msgObj.Template),
			zap.Object("message", msgObj.Message),
		)
		e = parseError{"Both Template and Message defined"}
	}
	if msgObj.Template == "" && (msgObj.Message == nil || len(msgObj.Message) == 0) {
		l.Error(
			"Either Template or Message should be defined",
			zap.String("template", msgObj.Template),
			zap.Object("message", msgObj.Message),
		)
		e = parseError{"Either Template or Message should be defined"}
	}

	if msgObj.PushExpiry < 0 {
		errStr := "PushExpiry should be above 0"
		l.Error(errStr, zap.Int64("pushexpiry", msgObj.PushExpiry))
		e = parseError{errStr}
	}

	if e.Message != "" {
		return nil, e
	}

	l.Debug("Decoded message", zap.String("msg", msg))
	return msgObj, nil
}
