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

	workers "github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameProcessBatchWorker = "process_batch_worker"

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Logger  zap.Logger
	Workers *Worker
}

// NewProcessBatchWorker gets a new ProcessBatchWorker
func NewProcessBatchWorker(workers *Worker) *ProcessBatchWorker {
	b := &ProcessBatchWorker{
		Logger:  workers.Logger.With(zap.String("worker", "ProcessBatchWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Configured ProcessBatchWorker successfully")
	return b
}

func (b *ProcessBatchWorker) incrFailedBatches(jobID uuid.UUID, totalBatches int, appName string) {
	failedJobs, err := b.Workers.RedisClient.Incr(fmt.Sprintf("%s-failedbatches", jobID.String())).Result()
	checkErr(b.Logger, err)
	ttl, err := b.Workers.RedisClient.TTL(fmt.Sprintf("%s-failedbatches", jobID.String())).Result()
	checkErr(b.Logger, err)
	if ttl < 0 {
		b.Workers.RedisClient.Expire(fmt.Sprintf("%s-failedbatches", jobID.String()), 7*24*time.Hour)
	}
	if float64(failedJobs)/float64(totalBatches) >= b.Workers.Config.GetFloat64("workers.processBatch.maxBatchFailure") {
		job := model.Job{}
		_, err := b.Workers.MarathonDB.Model(&job).Set("status = 'circuitbreak'").Where("id = ?", jobID).Returning("*").Update()
		checkErr(b.Logger, err)
		changedStatus, err := b.Workers.RedisClient.SetNX(fmt.Sprintf("%s-circuitbreak", jobID.String()), 1, 1*time.Minute).Result()
		checkErr(b.Logger, err)
		if changedStatus && b.Workers.SendgridClient != nil {
			var expireAt int64
			if ttl > 0 {
				expireAt = time.Now().Add(ttl).UnixNano()
			} else {
				expireAt = time.Now().Add(7 * 24 * time.Hour).UnixNano()
			}
			email.SendCircuitBreakJobEmail(b.Workers.SendgridClient, &job, appName, expireAt)
		}
	}
}

func (b *ProcessBatchWorker) sendToKafka(service, topic string, msg, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, deviceToken string, expiresAt int64, templateName string) error {
	pushExpiry := expiresAt / 1000000000 // convert from nanoseconds to seconds
	switch service {
	case "apns":
		err := b.Workers.Kafka.SendAPNSPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry, templateName)
		if err != nil {
			return err
		}
	case "gcm":
		err := b.Workers.Kafka.SendGCMPush(topic, deviceToken, msg, messageMetadata, pushMetadata, pushExpiry, templateName)
		if err != nil {
			return err
		}
	default:
		panic("service should be in ['apns', 'gcm']")
	}
	return nil
}

func (b *ProcessBatchWorker) updateJobUsersInfo(jobID uuid.UUID, numUsers int) error {
	job := model.Job{}
	_, err := b.Workers.MarathonDB.Model(&job).Set("completed_tokens = completed_tokens + ?", numUsers).Where("id = ?", jobID).Update()
	return err
}

func (b *ProcessBatchWorker) updateJobBatchesInfo(jobID uuid.UUID) error {
	job := model.Job{}
	_, err := b.Workers.MarathonDB.Model(&job).Set("completed_batches = completed_batches + 1").Where("id = ?", jobID).Returning("*").Update()
	if err != nil {
		return err
	}
	if job.TotalBatches != 0 && job.CompletedBatches == 1 && job.CompletedAt == 0 {
		job.TagRunning(b.Workers.MarathonDB, "process_batche_worker", "starting")
	}
	if job.TotalBatches != 0 && job.CompletedBatches >= job.TotalBatches && job.CompletedAt == 0 {
		l := b.Logger.With(
			zap.String("source", "processBatchWorker"),
			zap.String("operation", "updateJobBatchesInfo"),
			zap.Int("totalBatches", job.TotalBatches),
			zap.Int("completedBatches", job.CompletedBatches),
		)

		log.I(l, "Finished all batches")
		job.TagSuccess(b.Workers.MarathonDB, "process_batche_worker", "Finished all batches")
		job.CompletedAt = time.Now().UnixNano()
		_, err = b.Workers.MarathonDB.Model(&job).Column("completed_at").Update()
		if err != nil {
			return err
		}
		at := time.Now().Add(b.Workers.Config.GetDuration("workers.processBatch.intervalToSendCompletedJob")).UnixNano()
		_, err = b.Workers.ScheduleJobCompletedJob(jobID.String(), at)
	}
	return err
}

func (b *ProcessBatchWorker) moveJobToPausedQueue(jobID uuid.UUID, message *workers.Msg) {
	_, err := b.Workers.RedisClient.RPush(fmt.Sprintf("%s-pausedjobs", jobID.String()), message.ToJson()).Result()
	checkErr(b.Logger, err)
	ttl, err := b.Workers.RedisClient.TTL(fmt.Sprintf("%s-pausedjobs", jobID.String())).Result()
	checkErr(b.Logger, err)
	if ttl < 0 {
		b.Workers.RedisClient.Expire(fmt.Sprintf("%s-pausedjobs", jobID.String()), 7*24*time.Hour)
	}
}

func (b *ProcessBatchWorker) checkErrWithReEnqueue(parsed *BatchWorkerMessage, l zap.Logger, err error) {
	if err != nil {
		at := time.Now().Add(time.Duration(rand.Intn(100)) * time.Second).UnixNano()
		b.Workers.ScheduleProcessBatchJob(
			parsed.JobID.String(),
			parsed.AppName,
			&parsed.Users,
			at,
		)
	}
	checkErr(l, err)
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (b *ProcessBatchWorker) Process(message *workers.Msg) {
	batchErrorCounter := 0
	l := b.Logger.With(
		zap.String("source", "processBatchWorker"),
		zap.String("operation", "process"),
		zap.String("worker", nameProcessBatchWorker),
	)
	log.I(l, "starting")
	arr, err := message.Args().Array()
	checkErr(l, err)
	parsed, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(l, err)
	log.D(l, "Parsed message info successfully.")

	job, err := b.Workers.GetJob(parsed.JobID)
	b.checkErrWithReEnqueue(parsed, l, err)

	log.D(l, "Retrieved job successfully.")
	b.Workers.Statsd.Incr("starting_process_batch_worker", job.Labels(), 1)

	if job.ExpiresAt > 0 && job.ExpiresAt < time.Now().UnixNano() {
		log.I(l, "expired")
		return
	}

	switch job.Status {
	case "circuitbreak":
		log.I(l, "circuit break")
		b.moveJobToPausedQueue(job.ID, message)
		return
	case "paused":
		log.I(l, "paused")
		b.moveJobToPausedQueue(job.ID, message)
		return
	case "stopped":
		log.I(l, "stopped")
		return
	default:
		log.D(l, "valid")
	}

	templatesByNameAndLocale, err := job.GetJobTemplatesByNameAndLocale(b.Workers.MarathonDB)
	if err != nil {
		b.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
	}
	b.checkErrWithReEnqueue(parsed, l, err)
	log.D(l, "Retrieved templatesByNameAndLocale successfully.", func(cm log.CM) {
		cm.Write(zap.Object("templatesByNameAndLocale", templatesByNameAndLocale))
	})

	topicTemplate := b.Workers.Config.GetString("workers.topicTemplate")
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
			b.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
			checkErr(l, fmt.Errorf("there is no template for the given locale or 'en'"))
		}

		msgStr, msgErr := BuildMessageFromTemplate(template, job.Context)
		if msgErr != nil {
			b.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		}
		checkErr(l, msgErr)
		var msg map[string]interface{}
		err = json.Unmarshal([]byte(msgStr), &msg)
		if err != nil {
			b.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		}
		checkErr(l, err)
		pushMetadata := map[string]interface{}{
			"userId":       user.UserID,
			"pushTime":     time.Now().Unix(),
			"templateName": templateName,
			"jobId":        job.ID.String(),
			"pushType":     "massive",
			"muid":         uuid.NewV4().String(),
		}

		dryRun := false
		if val, ok := job.Metadata["dryRun"]; ok {
			if dryRun, ok = val.(bool); ok {
				pushMetadata["dryRun"] = dryRun
			}
		}

		err = b.sendToKafka(job.Service, topic, msg, job.Metadata, pushMetadata, user.Token, job.ExpiresAt, templateName)
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
	err = b.updateJobBatchesInfo(parsed.JobID)
	checkErr(l, err)
	log.D(l, "Updated job batches info successfully.")
	err = b.updateJobUsersInfo(parsed.JobID, len(parsed.Users)-batchErrorCounter)
	checkErr(l, err)
	log.D(l, "Updated job users info successfully.")
	if float64(batchErrorCounter)/float64(len(parsed.Users)) > b.Workers.Config.GetFloat64("workers.processBatch.maxUserFailureInBatch") {
		b.incrFailedBatches(job.ID, job.TotalBatches, parsed.AppName)
		checkErr(l, fmt.Errorf("failed to send message to several users, considering batch as failed"))
	}
	log.I(l, "finished")
}
