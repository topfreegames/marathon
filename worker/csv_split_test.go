/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
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
	"math/rand"

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("CSVSplit Worker", func() {
	var app *model.App
	var template *model.Template

	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()),
		zap.FatalLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())
	createCSVSplitWorker := worker.NewCSVSplitWorker(w)

	BeforeEach(func() {
		w.S3Client = NewFakeS3(w.Config)
		app = CreateTestApp(w.MarathonDB.DB)
		template = CreateTestTemplate(w.MarathonDB.DB, app.ID)
		w.RedisClient.FlushAll()
	})

	Describe("Process", func() {
		It("should create jobs for the next worker", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
				"csvPath": "test/test.csv",
			})

			randomData := make([]byte, 47000000)
			rand.Read(randomData)
			copy(randomData[:], "userIds\n")

			_, err := w.S3Client.PutObject("test/test.csv", &randomData)
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			message, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			createCSVSplitWorker.Process(message)

			size, err := w.RedisClient.LLen("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(2)))

			// Check first info
			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			message, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			var msg worker.BatchPart
			data := message.Args().ToJson()
			err = json.Unmarshal([]byte(data), &msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(msg.TotalSize).To(Equal(47000000))
			Expect(msg.TotalParts).To(Equal(2))
			Expect(msg.Size).To(Equal(31457280))
			Expect(msg.Part).To(Equal(0))
			Expect(msg.Job.ID).To(Equal(j.ID))

			// Check secound info
			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			message, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			data = message.Args().ToJson()
			err = json.Unmarshal([]byte(data), &msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(msg.TotalSize).To(Equal(47000000))
			Expect(msg.TotalParts).To(Equal(2))
			Expect(msg.Size).To(Equal(47000000 - 31457280))
			Expect(msg.Part).To(Equal(1))
			Expect(msg.Job.ID).To(Equal(j.ID))
		})

		It("should panic if is incorrect file content", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
				"csv_path": "test/test.csv",
			})

			randomData := make([]byte, 10)
			rand.Read(randomData)

			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3
			_, err := w.S3Client.PutObject("test/test.csv", &randomData)
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).Should(Panic())
		})

		It("should panic if is incorrect file path", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
				"csv_path": "test.csv",
			})

			randomData := make([]byte, 10)
			rand.Read(randomData)

			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).Should(Panic())
		})

		It("should panic job id don`t exist", func() {

			randomData := make([]byte, 10)
			rand.Read(randomData)

			_, err := w.CSVSplitJob(uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())
			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).Should(Panic())
		})

		It("should do nothing if job status is stopped", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			randomData := make([]byte, 47000000)
			rand.Read(randomData)
			copy(randomData[:], "userIds\n")
			_, err := w.S3Client.PutObject("test/test.csv", &randomData)
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			_, err = w.MarathonDB.DB.Model(&model.Job{}).Set("status = 'stopped'").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			size, err := w.RedisClient.LLen("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(0)))
		})
	})
})
