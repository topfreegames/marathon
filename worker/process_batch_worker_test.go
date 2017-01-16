/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permifsion is hereby granted, free of charge, to any person obtaining a copy of
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

package worker_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/messages"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

func getNextMessageFrom(kafkaBrokers []string, topic string, partition int32, offset int64) (*sarama.ConsumerMessage, error) {
	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	Expect(err).NotTo(HaveOccurred())
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	Expect(err).NotTo(HaveOccurred())
	defer partitionConsumer.Close()

	msg := <-partitionConsumer.Messages()
	return msg, nil
}

var _ = Describe("ProcessBatch Worker", func() {
	var logger zap.Logger
	var config *viper.Viper
	var processBatchWorker *worker.ProcessBatchWorker
	var app *model.App
	var template *model.Template
	var job *model.Job
	var gcmJob *model.Job
	var users []worker.User
	var mockKafkaProducer *FakeKafkaProducer

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		config = GetConf()
		mockKafkaProducer = NewFakeKafkaProducer()
		processBatchWorker = worker.NewProcessBatchWorker(config, logger, mockKafkaProducer)

		app = CreateTestApp(processBatchWorker.MarathonDB.DB)
		defaults := map[string]interface{}{
			"user_name":   "Someone",
			"object_name": "village",
		}
		body := map[string]interface{}{
			"alert": "{{user_name}} just liked your {{object_name}}!",
		}
		template = CreateTestTemplate(processBatchWorker.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": defaults,
			"body":     body,
			"locale":   "en",
			"name":     "village-like",
		})
		CreateTestTemplate(processBatchWorker.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": map[string]interface{}{
				"user_name":   "Alguém",
				"object_name": "vila",
			},
			"body": map[string]interface{}{
				"alert": "{{user_name}} curtiram sua {{object_name}}!",
			},
			"locale": "pt",
			"name":   "village-like",
		})
		CreateTestTemplate(processBatchWorker.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": map[string]interface{}{
				"user_name":   "Quelqu'un",
				"object_name": "ville",
			},
			"body": map[string]interface{}{
				"alert": "{{user_name}} a aimé ta {{object_name}}!",
			},
			"locale": "fr",
			"name":   "village-like",
		})
		context := map[string]interface{}{
			"user_name": "Everyone",
		}
		job = CreateTestJob(processBatchWorker.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
			"context": context,
		})
		gcmJob = CreateTestJob(processBatchWorker.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
			"context": context,
			"service": "gcm",
		})
		Expect(job.CompletedAt).To(Equal(int64(0)))
		users = make([]worker.User, 2)
		for index := range users {
			id := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			users[index] = worker.User{
				UserID: id,
				Token:  token,
				Locale: "en",
			}
		}
	})

	Describe("Process", func() {
		It("should process when service is gcm and increment job completed batches", func() {
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				gcmJob.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			for idx := range users {
				m := mockKafkaProducer.GCMMessages[idx]
				var gcmMessage messages.GCMMessage
				err = json.Unmarshal([]byte(m), &gcmMessage)
				Expect(err).NotTo(HaveOccurred())

				Expect(gcmMessage.To).To(Equal(users[idx].Token))
				Expect(gcmMessage.TimeToLive).To(BeEquivalentTo(gcmJob.ExpiresAt / 1000000000))
				Expect(gcmMessage.Data["alert"]).To(Equal("Everyone just liked your village!"))
				Expect(gcmMessage.Data["m"].(map[string]interface{})["meta"]).To(Equal(gcmJob.Metadata["meta"]))
			}
		})

		It("should process when service is apns and increment job completed batches", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("service = apns").Where("id = ?", job.ID).Update()
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			for idx := range users {
				m := mockKafkaProducer.APNSMessages[idx]
				var apnsMessage messages.APNSMessage
				err = json.Unmarshal([]byte(m), &apnsMessage)
				Expect(err).NotTo(HaveOccurred())
				Expect(apnsMessage.DeviceToken).To(Equal(users[idx].Token))
				Expect(apnsMessage.PushExpiry).To(BeEquivalentTo(job.ExpiresAt / 1000000000))
				Expect(apnsMessage.Payload.Aps["alert"]).To(Equal("Everyone just liked your village!"))
				Expect(apnsMessage.Payload.M["meta"]).To(Equal(job.Metadata["meta"]))
				idx++
			}
		})

		It("should set job completedAt if last batch", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("completed_batches = 0").Set("total_batches = 1").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(1))
			Expect(dbJob.CompletedAt).To(BeNumerically("~", time.Now().UnixNano(), 50000000))
		})

		It("should not set job completedAt if not last batch", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("completed_batches = 0").Set("total_batches = 2").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(1))
			Expect(dbJob.CompletedAt).To(Equal(int64(0)))
		})

		It("should not set job completedAt total_batches is null", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("completed_batches = 0").Set("total_batches = null").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(1))
			Expect(dbJob.CompletedAt).To(Equal(int64(0)))
		})

		It("should increment job completed users", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("service = gcm").Where("id = ?", job.ID).Update()
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedUsers).To(Equal(len(users)))
		})

		It("should not process batch if job is expired", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("completed_batches = 0").Set("expires_at = ?", time.Now().UnixNano()-50000).Where("id = ?", job.ID).Update()
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			Expect(err).NotTo(HaveOccurred())

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.CompletedUsers).To(Equal(0))
		})

		It("should not process batch if job is stopped", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("completed_batches = 0").Set("status = 'stopped'").Where("id = ?", job.ID).Update()
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.CompletedUsers).To(Equal(0))
		})

		It("should process the message using the correct template", func() {
			users = make([]worker.User, 2)
			for index := range users {
				id := uuid.NewV4().String()
				token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
				users[index] = worker.User{
					UserID: id,
					Token:  token,
					Locale: "pt",
				}
			}
			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			for idx := range users {
				m := mockKafkaProducer.APNSMessages[idx]

				var apnsMessage messages.APNSMessage
				err = json.Unmarshal([]byte(m), &apnsMessage)
				Expect(err).NotTo(HaveOccurred())
				Expect(apnsMessage.DeviceToken).To(Equal(users[idx].Token))
				Expect(apnsMessage.Payload.Aps["alert"]).To(Equal("Everyone curtiram sua vila!"))
				idx++
			}
		})

		It("should process the message and put the right pushMetadata on it if apns push", func() {
			userID := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			user := worker.User{
				UserID: userID,
				Token:  token,
				Locale: "pt",
			}
			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				job.ID,
				appName,
				&[]worker.User{user},
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			expectedPushMetadata := map[string]interface{}{
				"jobId":        job.ID.String(),
				"userId":       userID,
				"templateName": job.TemplateName,
				"pushType":     "massive",
			}

			m := mockKafkaProducer.APNSMessages[0]
			var apnsMessage messages.APNSMessage
			err = json.Unmarshal([]byte(m), &apnsMessage)

			Expect(err).NotTo(HaveOccurred())
			Expect(apnsMessage.Metadata).To(BeEquivalentTo(expectedPushMetadata))
		})

		It("should process the message and put the right pushMetadata on it if gcm push", func() {
			userID := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			user := worker.User{
				UserID: userID,
				Token:  token,
				Locale: "pt",
			}
			appName := strings.Split(app.BundleID, ".")[2]
			messageObj := []interface{}{
				gcmJob.ID,
				appName,
				&[]worker.User{user},
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			expectedPushMetadata := map[string]interface{}{
				"jobId":        gcmJob.ID.String(),
				"userId":       userID,
				"templateName": job.TemplateName,
				"pushType":     "massive",
			}

			m := mockKafkaProducer.GCMMessages[0]
			var gcmMessage messages.GCMMessage
			err = json.Unmarshal([]byte(m), &gcmMessage)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcmMessage.Metadata).To(BeEquivalentTo(expectedPushMetadata))
		})

		It("should increment failedJobs", func() {
			// unexistent template
			processBatchWorker.MarathonDB.DB.Exec("DELETE FROM templates;")
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("total_batches = 100").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())
			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { processBatchWorker.Process(message) }).Should(Panic())

			failedJobs, err := processBatchWorker.RedisClient.Get(fmt.Sprintf("%s-failedbatches", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(failedJobs).To(Equal("1"))
			ttl, err := processBatchWorker.RedisClient.TTL(fmt.Sprintf("%s-failedbatches", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically("~", 7*24*time.Hour, 10))

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.Status).To(Equal(""))
		})

		It("should increment failedJobs and mark job status as circuitbreak", func() {
			// unexistent template
			processBatchWorker.MarathonDB.DB.Exec("DELETE FROM templates;")
			err := processBatchWorker.RedisClient.Set(fmt.Sprintf("%s-failedbatches", job.ID.String()), 4, time.Hour).Err()
			Expect(err).NotTo(HaveOccurred())
			_, err = processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("total_batches = 100").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { processBatchWorker.Process(message) }).Should(Panic())

			failedJobs, err := processBatchWorker.RedisClient.Get(fmt.Sprintf("%s-failedbatches", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(failedJobs).To(Equal("5"))
			ttl, err := processBatchWorker.RedisClient.TTL(fmt.Sprintf("%s-failedbatches", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically("~", time.Hour, 10))

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.Status).To(Equal("circuitbreak"))
		})

		//// TODO: found out how to test this
		XIt("should increment failedJobs if push no sent to users", func() {
		})

		It("should not process job and add it to paused jobs list if job is paused", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("status = 'paused'").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.CompletedUsers).To(Equal(0))

			pausedMsg, err := processBatchWorker.RedisClient.LPop(fmt.Sprintf("%s-pausedjobs", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pausedMsg).To(Equal(message.ToJson()))
		})

		It("should not process job and add it to paused jobs list if job is circuitbreak", func() {
			_, err := processBatchWorker.MarathonDB.DB.Model(&model.Job{}).Set("status = 'circuitbreak'").Where("id = ?", job.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			appName := strings.Split(app.BundleID, ".")[2]

			messageObj := []interface{}{
				job.ID,
				appName,
				users,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			processBatchWorker.Process(message)

			dbJob := model.Job{
				ID: job.ID,
			}
			err = processBatchWorker.MarathonDB.DB.Select(&dbJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CompletedBatches).To(Equal(0))
			Expect(dbJob.CompletedUsers).To(Equal(0))

			pausedMsg, err := processBatchWorker.RedisClient.LPop(fmt.Sprintf("%s-pausedjobs", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pausedMsg).To(Equal(message.ToJson()))
		})
	})
})
