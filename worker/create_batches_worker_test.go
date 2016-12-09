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
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package worker_test

import (
	"encoding/json"
	"strings"

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("CreateBatches Worker", func() {
	var logger zap.Logger
	var config *viper.Viper
	var createBatchesWorker *worker.CreateBatchesWorker
	var job *model.Job
	var app *model.App
	var template *model.Template
	var context map[string]interface{}
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()),
			zap.FatalLevel,
		)
		config = GetConf()
		w := worker.NewWorker(false, logger, GetConfPath())
		createBatchesWorker = worker.NewCreateBatchesWorker(config, logger, w)
		createBatchesWorker.S3Client = &FakeS3{}
		app = CreateTestApp(createBatchesWorker.MarathonDB.DB)
		defaults := map[string]interface{}{
			"user_name":   "Someone",
			"object_name": "village",
		}
		body := map[string]interface{}{
			"alert": "{{user_name}} just liked your {{object_name}}!",
		}
		template = CreateTestTemplate(createBatchesWorker.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": defaults,
			"body":     body,
			"locale":   "en",
		})
		context = map[string]interface{}{
			"user_name": "Everyone",
		}
		job = CreateTestJob(createBatchesWorker.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
			"context": context,
		})
		users := make([]worker.User, 2)
		for index := range users {
			id := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			users[index] = worker.User{
				UserID: id,
				Token:  token,
				Locale: "en",
			}
		}
		createBatchesWorker.RedisClient.FlushAll()
	})

	Describe("Process", func() {
		It("should panic if jobID is invalid", func() {
			m := map[string]interface{}{
				"jid":  2,
				"args": []string{"8df23db3-b02e-40a0-82b6-4993876c5fc8"},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesWorker.Process(msg) }).Should(Panic())
		})

		It("should panic if csvPath is invalid", func() {
			j := CreateTestJob(createBatchesWorker.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "algum",
			})
			m := map[string]interface{}{
				"jid":  2,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(func() { createBatchesWorker.Process(msg) }).Should(Panic())
		})

		It("should not panic if csvPath and jobID are valid", func() {
			j := CreateTestJob(createBatchesWorker.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "obj2.csv",
			})
			m := map[string]interface{}{
				"jid":  2,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker", func() {
			a := CreateTestApp(createBatchesWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(createBatchesWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "obj1.csv",
			})
			m := map[string]interface{}{
				"jid":  2,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := createBatchesWorker.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(len((j1["args"].([]interface{}))[2].([]interface{})) + len((j2["args"].([]interface{}))[2].([]interface{}))).To(BeEquivalentTo(10))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if service is gcm", func() {
			a := CreateTestApp(createBatchesWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(createBatchesWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "obj1.csv",
				"service": "gcm",
			})
			m := map[string]interface{}{
				"jid":  2,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := createBatchesWorker.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(len((j1["args"].([]interface{}))[2].([]interface{})) + len((j2["args"].([]interface{}))[2].([]interface{}))).To(BeEquivalentTo(10))
		})
	})

	It("should not panic if job is a reexecution", func() {
		a := CreateTestApp(createBatchesWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
		j := CreateTestJob(createBatchesWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
			"context": context,
			"filters": map[string]interface{}{},
			"csvPath": "obj1.csv",
			"service": "gcm",
		})
		m := map[string]interface{}{
			"jid":  2,
			"args": []string{j.ID.String()},
		}
		smsg, err := json.Marshal(m)
		Expect(err).NotTo(HaveOccurred())
		msg, err := workers.NewMsg(string(smsg))
		Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
		res, err := createBatchesWorker.RedisClient.LLen("queue:process_batch_worker").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeEquivalentTo(2))
		job1, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
		Expect(err).NotTo(HaveOccurred())
		job2, err := createBatchesWorker.RedisClient.LPop("queue:process_batch_worker").Result()
		Expect(err).NotTo(HaveOccurred())
		j1 := map[string]interface{}{}
		j2 := map[string]interface{}{}
		err = json.Unmarshal([]byte(job1), &j1)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal([]byte(job2), &j2)
		Expect(err).NotTo(HaveOccurred())
		Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
		Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
		Expect(len((j1["args"].([]interface{}))[2].([]interface{})) + len((j2["args"].([]interface{}))[2].([]interface{}))).To(BeEquivalentTo(10))
		Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
	})
})
