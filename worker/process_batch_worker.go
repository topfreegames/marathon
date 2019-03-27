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
	"math/rand"
	"strings"
	"time"

	pg "gopkg.in/pg.v5"
	redis "gopkg.in/redis.v5"

	workers "github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameProcessBatchWorker = "process_batch_worker"

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Config         *viper.Viper
	Kafka          interfaces.PushProducer
	Logger         zap.Logger
	MarathonDB     *extensions.PGClient
	RedisClient    *redis.Client
	SendgridClient *extensions.SendgridClient
	Workers        *Worker
}

// NewProcessBatchWorker gets a new ProcessBatchWorker
func NewProcessBatchWorker(config *viper.Viper, logger zap.Logger, kafkaClient interfaces.PushProducer, workers *Worker) *ProcessBatchWorker {
	l := logger.With(
		zap.String("source", "processBatchWorker"),
		zap.String("operation", "NewProcessBatchWorker"),
	)
	//Wait 10s at max for a connection
	k := kafkaClient
	if k == nil {
		var kafka *extensions.KafkaProducer
		var err error
		kafka, err = extensions.NewKafkaProducer(config, logger, workers.Statsd)
		checkErr(l, err)
		k = kafka
	}
	marathonDB, err := extensions.NewPGClient("db", config, logger)
	checkErr(l, err)
	redisClient, err := extensions.NewRedis("workers", config, logger)
	checkErr(l, err)

	batchWorker := &ProcessBatchWorker{
		Config:      config,
		Logger:      logger,
		Kafka:       k,
		MarathonDB:  marathonDB,
		RedisClient: redisClient,
		Workers:     workers,
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
		changedStatus, err := batchWorker.RedisClient.SetNX(fmt.Sprintf("%s-circuitbreak", jobID.String()), 1, 1*time.Minute).Result()
		checkErr(batchWorker.Logger, err)
		if changedStatus && batchWorker.SendgridClient != nil {
			var expireAt int64
			if ttl > 0 {
				expireAt = time.Now().Add(ttl).UnixNano()
			} else {
				expireAt = time.Now().Add(7 * 24 * time.Hour).UnixNano()
			}
			email.SendCircuitBreakJobEmail(batchWorker.SendgridClient, &job, appName, expireAt)
		}
	}
}

func (batchWorker *ProcessBatchWorker) sendToKafka(service, topic string, msg, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, deviceToken string, expiresAt int64, templateName string) error {
	pushExpiry := expiresAt / 1000000000 // convert from nanoseconds to seconds
	switch service {
	case "apns":
		err := batchWorker.Kafka.SendAPNSPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry, templateName)
		if err != nil {
			return err
		}
	case "gcm":
		err := batchWorker.Kafka.SendGCMPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry, templateName)
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

func (batchWorker *ProcessBatchWorker) getJobTemplatesByNameAndLocale(appID uuid.UUID, templateName string) (map[string]map[string]model.Template, error) {
	var templates []model.Template
	var err error
	if len(strings.Split(templateName, ",")) > 1 {
		err = batchWorker.MarathonDB.DB.Model(&templates).Where(
			"app_id = ? AND name IN (?)",
			appID,
			pg.In(strings.Split(templateName, ",")),
		).Select()
	} else {
		err = batchWorker.MarathonDB.DB.Model(&templates).Where(
			"app_id = ? AND name = ?",
			appID,
			templateName,
		).Select()
	}
	if err != nil {
		return nil, err
	}
	templateByLocale := make(map[string]map[string]model.Template)
	for _, tpl := range templates {
		if templateByLocale[tpl.Name] != nil {
			templateByLocale[tpl.Name][tpl.Locale] = tpl
		} else {
			templateByLocale[tpl.Name] = map[string]model.Template{
				tpl.Locale: tpl,
			}
		}
	}

	if len(templateByLocale) == 0 {
		return nil, fmt.Errorf("No templates were found with name %s", templateName)
	}
	return templateByLocale, nil
}

func (batchWorker *ProcessBatchWorker) updateJobUsersInfo(jobID uuid.UUID, numUsers int) error {
	job := model.Job{}
	_, err := batchWorker.MarathonDB.DB.Model(&job).Set("completed_tokens = completed_tokens + ?", numUsers).Where("id = ?", jobID).Update()
	return err
}

func (batchWorker *ProcessBatchWorker) updateJobBatchesInfo(jobID uuid.UUID) error {
	job := model.Job{}
	_, err := batchWorker.MarathonDB.DB.Model(&job).Set("completed_batches = completed_batches + 1").Where("id = ?", jobID).Returning("*").Update()
	if err != nil {
		return err
	}
	if job.TotalBatches != 0 && job.CompletedBatches == 1 && job.CompletedAt == 0 {
		job.TagRunning(batchWorker.MarathonDB, "process_batche_worker", "starting")
	}
	if job.TotalBatches != 0 && job.CompletedBatches >= job.TotalBatches && job.CompletedAt == 0 {
		l := batchWorker.Logger.With(
			zap.String("source", "processBatchWorker"),
			zap.String("operation", "updateJobBatchesInfo"),
			zap.Int("totalBatches", job.TotalBatches),
			zap.Int("completedBatches", job.CompletedBatches),
		)

		log.I(l, "Finished all batches")
		job.TagSuccess(batchWorker.MarathonDB, "process_batche_worker", "Finished all batches")
		job.CompletedAt = time.Now().UnixNano()
		_, err = batchWorker.MarathonDB.DB.Model(&job).Column("completed_at").Update()
		if err != nil {
			return err
		}
		at := time.Now().Add(batchWorker.Config.GetDuration("workers.processBatch.intervalToSendCompletedJob")).UnixNano()
		_, err = batchWorker.Workers.ScheduleJobCompletedJob(jobID.String(), at)
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

func (batchWorker *ProcessBatchWorker) checkErrWithReEnqueue(parsed *BatchWorkerMessage, l zap.Logger, err error) {
	if err != nil {
		at := time.Now().Add(time.Duration(rand.Intn(100)) * time.Second).UnixNano()
		batchWorker.Workers.ScheduleProcessBatchJob(
			parsed.JobID.String(),
			parsed.AppName,
			&parsed.Users,
			at,
		)
	}
	checkErr(l, err)
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	batchErrorCounter := 0
	l := batchWorker.Logger.With(
		zap.String("source", "processBatchWorker"),
		zap.String("operation", "process"),
		zap.String("worker", nameProcessBatchWorker),
	)
	log.I(l, "starting")
	batchWorker.Workers.Statsd.Incr("starting_process_batch_worker", []string{}, 1)
	arr, err := message.Args().Array()
	checkErr(l, err)
	parsed, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(l, err)
	log.D(l, "Parsed message info successfully.")

	job, err := batchWorker.getJob(parsed.JobID)
	batchWorker.checkErrWithReEnqueue(parsed, l, err)
	log.D(l, "Retrieved job successfully.")

	if job.ExpiresAt > 0 && job.ExpiresAt < time.Now().UnixNano() {
		log.I(l, "expired")
		return
	}

	switch job.Status {
	case "circuitbreak":
		log.I(l, "circuit break")
		batchWorker.moveJobToPausedQueue(job.ID, message)
		return
	case "paused":
		log.I(l, "paused")
		batchWorker.moveJobToPausedQueue(job.ID, message)
		return
	case "stopped":
		log.I(l, "stopped")
		return
	default:
		log.D(l, "valid")
	}

	templatesByNameAndLocale, err := batchWorker.getJobTemplatesByNameAndLocale(job.AppID, job.TemplateName)
	if err != nil {
		batchWorker.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
	}
	batchWorker.checkErrWithReEnqueue(parsed, l, err)
	log.D(l, "Retrieved templatesByNameAndLocale successfully.", func(cm log.CM) {
		cm.Write(zap.Object("templatesByNameAndLocale", templatesByNameAndLocale))
	})

	topicTemplate := batchWorker.Config.GetString("workers.topicTemplate")
	topic := BuildTopicName(parsed.AppName, job.Service, topicTemplate)
	log.D(l, "Built topic name successfully.", func(cm log.CM) {
		cm.Write(zap.String("topic", topic))
	})
	for _, user := range parsed.Users {
		templateName := job.TemplateName
		templateNames := strings.Split(job.TemplateName, ",")

		if templateNames != nil && len(templateNames) > 1 {
			templateName = RandomElementFromSlice(templateNames)
			log.D(l, "selected template", func(cm log.CM) {
				cm.Write(zap.Object("name", templateName))
			})
		}

		templatesByLocale := templatesByNameAndLocale[templateName]
		var template model.Template
		if val, ok := templatesByLocale[strings.ToLower(user.Locale)]; ok {
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
			"userId": user.UserID,
			// "fiu":          user.Fiu,
			// "adid":         user.Adid,
			"pushTime": time.Now().Unix(),
			// "vendorId":     user.VendorID,
			"templateName": templateName,
			"jobId":        job.ID.String(),
			"pushType":     "massive",
			"muid":         uuid.NewV4().String(),
		}
		// if user.CreatedAt.Unix() > 0 {
		// 	pushMetadata["tokenCreatedAt"] = user.CreatedAt.Unix()
		// }

		dryRun := false
		if val, ok := job.Metadata["dryRun"]; ok {
			if dryRun, ok = val.(bool); ok {
				pushMetadata["dryRun"] = dryRun
			}
		}

		err = batchWorker.sendToKafka(job.Service, topic, msg, job.Metadata, pushMetadata, user.Token, job.ExpiresAt, templateName)
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
	log.D(l, "Sent push to pusher for batch users.")
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
	log.I(l, "finished")
}
