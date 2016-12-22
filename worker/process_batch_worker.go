/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
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

package worker

import (
	"encoding/json"
	"fmt"
	"time"

	workers "github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Config     *viper.Viper
	Kafka      *extensions.KafkaClient
	Logger     zap.Logger
	MarathonDB *extensions.PGClient
	Zookeeper  *extensions.ZookeeperClient
}

// NewProcessBatchWorker gets a new ProcessBatchWorker
func NewProcessBatchWorker(config *viper.Viper, logger zap.Logger) *ProcessBatchWorker {
	l := logger.With(
		zap.String("source", "processBatchWorker"),
		zap.String("operation", "NewProcessBatchWorker"),
	)
	zookeeper, err := extensions.NewZookeeperClient(config, logger)
	checkErr(l, err)
	//Wait 10s at max for a connection
	zookeeper.WaitForConnection(10)
	kafka, err := extensions.NewKafkaClient(zookeeper, config, logger)
	checkErr(l, err)
	marathonDB, err := extensions.NewPGClient("db", config, logger)
	checkErr(l, err)
	batchWorker := &ProcessBatchWorker{
		Config:     config,
		Logger:     logger,
		Kafka:      kafka,
		Zookeeper:  zookeeper,
		MarathonDB: marathonDB,
	}
	return batchWorker
}

func (batchWorker *ProcessBatchWorker) sendToKafka(service, topic string, msg, metadata map[string]interface{}, deviceToken string, expiresAt int64) error {
	pushExpiry := expiresAt / 1000000000 // convert from nanoseconds to seconds
	switch service {
	case "apns":
		_, _, err := batchWorker.Kafka.SendAPNSPush(topic, deviceToken, msg, metadata, pushExpiry)
		if err != nil {
			return err
		}
	case "gcm":
		_, _, err := batchWorker.Kafka.SendGCMPush(topic, deviceToken, msg, metadata, pushExpiry)
		if err != nil {
			return err
		}
	default:
		panic("service should be in ['apns', 'gcm']")
	}
	return nil
}

func (batchWorker *ProcessBatchWorker) getJob(jobID uuid.UUID) (*model.Job, error) {
	job := model.Job{
		ID: jobID,
	}
	err := batchWorker.MarathonDB.DB.Select(&job)
	return &job, err
}

func (batchWorker *ProcessBatchWorker) getJobTemplatesByLocale(appID uuid.UUID, templateName string) (map[string]*model.Template, error) {
	templateByLocale := make(map[string]*model.Template)
	var templates []model.Template
	err := batchWorker.MarathonDB.DB.Model(&templates).Where("app_id = ? AND name = ?", appID, templateName).Select()
	if err != nil {
		return nil, err
	}
	for _, tpl := range templates {
		templateByLocale[tpl.Locale] = &tpl
	}

	return templateByLocale, nil
}

func (batchWorker *ProcessBatchWorker) updateJobUsersInfo(jobID uuid.UUID, numUsers int) error {
	job := model.Job{}
	_, err := batchWorker.MarathonDB.DB.Model(&job).Set("completed_users = completed_users + ?", numUsers).Where("id = ?", jobID).Returning("*").Update()
	return err
}

func (batchWorker *ProcessBatchWorker) updateJobBatchesInfo(jobID uuid.UUID) error {
	job := model.Job{}
	_, err := batchWorker.MarathonDB.DB.Model(&job).Set("completed_batches = completed_batches + 1").Where("id = ?", jobID).Returning("*").Update()
	if err != nil {
		return err
	}
	if job.CompletedBatches >= job.TotalBatches && job.CompletedAt == 0 {
		job.CompletedAt = time.Now().UnixNano()
		_, err = batchWorker.MarathonDB.DB.Model(&job).Column("completed_at").Update()
	}
	return err
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	l := batchWorker.Logger.With(
		zap.String("source", "processBatchWorker"),
		zap.String("operation", "process"),
	)
	log.I(l, "starting process_batch_worker")
	arr, err := message.Args().Array()
	checkErr(l, err)
	parsed, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(l, err)
	log.D(l, "Parsed message info successfully.")

	job, err := batchWorker.getJob(parsed.JobID)
	checkErr(l, err)
	log.D(l, "Retrieved job successfully.")

	if job.ExpiresAt > 0 && job.ExpiresAt < time.Now().UnixNano() {
		log.I(l, "expired process_batch_worker")
		return
	}

	templatesByLocale, err := batchWorker.getJobTemplatesByLocale(job.AppID, job.TemplateName)
	checkErr(l, err)
	log.D(l, "Retrieved templatesByLocale successfully.", func(cm log.CM) {
		cm.Write(zap.Object("templatesByLocale", templatesByLocale))
	})

	topicTemplate := batchWorker.Config.GetString("workers.topicTemplate")
	topic := BuildTopicName(parsed.AppName, job.Service, topicTemplate)
	log.D(l, "Built topic name successfully.", func(cm log.CM) {
		cm.Write(zap.String("topic", topic))
	})
	for _, user := range parsed.Users {
		var template *model.Template
		if val, ok := templatesByLocale[user.Locale]; ok {
			template = val
		} else {
			template = templatesByLocale["en"]
		}

		if template == nil {
			checkErr(l, fmt.Errorf("there is no template for the given locale or 'en'"))
		}

		msgStr, msgErr := BuildMessageFromTemplate(template, job.Context)
		checkErr(l, msgErr)
		var msg map[string]interface{}
		err = json.Unmarshal([]byte(msgStr), &msg)
		checkErr(l, err)
		err = batchWorker.sendToKafka(job.Service, topic, msg, job.Metadata, user.Token, job.ExpiresAt)
		if err != nil {
			log.E(l, "Failed to send message to Kafka.", func(cm log.CM) {
				cm.Write(
					zap.String("service", job.Service),
					zap.String("topic", topic),
					zap.Object("msg", msg),
					zap.Object("metadata", job.Metadata),
					zap.Object("token", user.Token),
					zap.Object("expiresAt", job.ExpiresAt),
					zap.Error(err),
				)
			})
		}
	}
	log.D(l, "Sent push to aguia for all batch users.")

	err = batchWorker.updateJobBatchesInfo(parsed.JobID)
	log.D(l, "Updated job batches info successfully.")
	err = batchWorker.updateJobUsersInfo(parsed.JobID, len(parsed.Users))
	log.D(l, "Updated job users info successfully.")
	checkErr(l, err)
	log.I(l, "finished process_batch_worker")
}
