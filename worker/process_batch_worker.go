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

	redis "gopkg.in/redis.v5"

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
	Config         *viper.Viper
	Kafka          *extensions.KafkaClient
	Logger         zap.Logger
	MarathonDB     *extensions.PGClient
	RedisClient    *redis.Client
	Zookeeper      *extensions.ZookeeperClient
	SendgridClient *extensions.SendgridClient
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
	redisClient, err := extensions.NewRedis("workers", config, logger)
	checkErr(l, err)

	batchWorker := &ProcessBatchWorker{
		Config:      config,
		Logger:      logger,
		Kafka:       kafka,
		Zookeeper:   zookeeper,
		MarathonDB:  marathonDB,
		RedisClient: redisClient,
	}

	apiKey := config.GetString("sendgrid.key")
	if apiKey != "" {
		batchWorker.SendgridClient = extensions.NewSendgridClient(config, logger, apiKey)
	}
	return batchWorker
}

func (batchWorker *ProcessBatchWorker) incrFailedBatches(jobID uuid.UUID, totalBatches int, appName string) {
	failedJobs, err := batchWorker.RedisClient.Incr(fmt.Sprintf("%s-failedbatches", jobID.String())).Result()
	checkErr(batchWorker.Logger, err)
	ttl, err := batchWorker.RedisClient.TTL(fmt.Sprintf("%s-failedbatches", jobID.String())).Result()
	checkErr(batchWorker.Logger, err)
	if ttl < 0 {
		batchWorker.RedisClient.Expire(fmt.Sprintf("%s-failedbatches", jobID.String()), 7*24*time.Hour)
	}
	if float64(failedJobs)/float64(totalBatches) >= batchWorker.Config.GetFloat64("workers.processBatch.maxBatchFailure") {
		job := model.Job{}
		_, err := batchWorker.MarathonDB.DB.Model(&job).Set("status = 'circuitbreak'").Where("id = ?", jobID).Returning("*").Update()
		checkErr(batchWorker.Logger, err)
		if batchWorker.SendgridClient != nil {
			var expireAt int64
			if ttl > 0 {
				expireAt = time.Now().Add(ttl).UnixNano()
			} else {
				expireAt = time.Now().Add(7 * 24 * time.Hour).UnixNano()
			}
			SendCircuitBreakJobEmail(batchWorker.SendgridClient, &job, appName, expireAt)
		}
	}
}

func (batchWorker *ProcessBatchWorker) sendToKafka(service, topic string, msg, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, deviceToken string, expiresAt int64) error {
	pushExpiry := expiresAt / 1000000000 // convert from nanoseconds to seconds
	switch service {
	//TODO por pushmetadata
	case "apns":
		_, _, err := batchWorker.Kafka.SendAPNSPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry)
		if err != nil {
			return err
		}
	case "gcm":
		_, _, err := batchWorker.Kafka.SendGCMPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry)
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

func (batchWorker *ProcessBatchWorker) getJobTemplatesByLocale(appID uuid.UUID, templateName string) (map[string]model.Template, error) {
	templateByLocale := make(map[string]model.Template)
	var templates []model.Template
	err := batchWorker.MarathonDB.DB.Model(&templates).Where("app_id = ? AND name = ?", appID, templateName).Select()
	if err != nil {
		return nil, err
	}
	for _, tpl := range templates {
		templateByLocale[tpl.Locale] = tpl
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
	if job.TotalBatches != 0 && job.CompletedBatches >= job.TotalBatches && job.CompletedAt == 0 {
		job.CompletedAt = time.Now().UnixNano()
		_, err = batchWorker.MarathonDB.DB.Model(&job).Column("completed_at").Update()
	}
	return err
}

func (batchWorker *ProcessBatchWorker) moveJobToPausedQueue(jobID uuid.UUID, message *workers.Msg) {
	_, err := batchWorker.RedisClient.RPush(fmt.Sprintf("%s-pausedjobs", jobID.String()), message.ToJson()).Result()
	checkErr(batchWorker.Logger, err)
	ttl, err := batchWorker.RedisClient.TTL(fmt.Sprintf("%s-pausedjobs", jobID.String())).Result()
	checkErr(batchWorker.Logger, err)
	if ttl < 0 {
		batchWorker.RedisClient.Expire(fmt.Sprintf("%s-pausedjobs", jobID.String()), 7*24*time.Hour)
	}
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	batchErrorCounter := 0
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

	switch job.Status {
	case "circuitbreak":
		log.I(l, "circuit break process_batch_worker")
		batchWorker.moveJobToPausedQueue(job.ID, message)
		return
	case "paused":
		log.I(l, "paused process_batch_worker")
		batchWorker.moveJobToPausedQueue(job.ID, message)
		return
	case "stopped":
		log.I(l, "stopped process_batch_worker")
		return
	default:
		log.D(l, "valid process_batch_worker")
	}

	templatesByLocale, err := batchWorker.getJobTemplatesByLocale(job.AppID, job.TemplateName)
	if err != nil {
		batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
	}
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
		var template model.Template
		if val, ok := templatesByLocale[user.Locale]; ok {
			template = val
		} else if val, ok := templatesByLocale["en"]; ok {
			template = val
		} else {
			batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
			checkErr(l, fmt.Errorf("there is no template for the given locale or 'en'"))
		}

		msgStr, msgErr := BuildMessageFromTemplate(template, job.Context)
		if msgErr != nil {
			batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		}
		checkErr(l, msgErr)
		var msg map[string]interface{}
		err = json.Unmarshal([]byte(msgStr), &msg)
		if err != nil {
			batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		}
		checkErr(l, err)
		pushMetadata := map[string]interface{}{
			"userId":       user.UserID,
			"templateName": job.TemplateName,
			"jobId":        job.ID.String(),
			"pushType":     "massive",
		}
		err = batchWorker.sendToKafka(job.Service, topic, msg, job.Metadata, pushMetadata, user.Token, job.ExpiresAt)
		if err != nil {
			batchErrorCounter = batchErrorCounter + 1
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
	log.D(l, "Sent push to aguia for batch users.")
	err = batchWorker.updateJobBatchesInfo(parsed.JobID)
	checkErr(l, err)
	log.D(l, "Updated job batches info successfully.")
	err = batchWorker.updateJobUsersInfo(parsed.JobID, len(parsed.Users)-batchErrorCounter)
	checkErr(l, err)
	log.D(l, "Updated job users info successfully.")
	if float64(batchErrorCounter)/float64(len(parsed.Users)) > batchWorker.Config.GetFloat64("workers.processBatch.maxUserFailureInBatch") {
		batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		checkErr(l, fmt.Errorf("failed to send message to several users, considering batch as failed"))
	}
	log.I(l, "finished process_batch_worker")
}
