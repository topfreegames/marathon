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
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
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
	"time"

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("CreateBatches Worker", func() {
	var app *model.App
	var template *model.Template
	var context map[string]interface{}

	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()),
		zap.FatalLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())
	createBatchesWorker := worker.NewCreateBatchesWorker(w)
	createCSVSplitWorker := worker.NewCSVSplitWorker(w)

	BeforeEach(func() {
		fakeS3 := NewFakeS3(w.Config)
		w.S3Client = fakeS3
		fakeData1 := []byte(`userids
9e558649-9c23-469d-a11c-59b05813e3d5
57be9009-e616-42c6-9cfe-505508ede2d0
a8e8d2d5-f178-4d90-9b31-683ad3aae920
5c3033c0-24ad-487a-a80d-68432464c8de
4223171e-c665-4612-9edd-485f229240bf
2df5bb01-15d1-4569-bc56-49fa0a33c4c3
67b872de-8ae4-4763-aef8-7c87a7f928a7
3f8732a1-8642-4f22-8d77-a9688dd6a5ae
21854bbf-ea7e-43e3-8f79-9ab2c121b941
843a61f8-45b3-44f9-9ab7-8becb2765653`)
		fakeData2 := []byte(`userids`)
		fakeData3 := []byte(`userids
e78431ca-69a8-4326-af1f-48f817a4a669
ee4455fe-8ff6-4878-8d7c-aec096bd68b4`)
		fakeData4 := []byte(`userids
b00b2bf9-9999-4be9-bdbd-cf0dbbd82cb2
6ce8a64f-c888-48c4-a040-f24ca7a71714`)
		fakeData5 := []byte(`userids
7ae62ce6-94fb-4636-9484-05bae4398505
9e3dfdf8-5991-4609-82ba-258ed2a78504
f57a0010-1318-4997-9a92-dcfb8ca0f24a
6be7b349-6034-4f99-847c-dab3ee4576d0
830a4cbf-c95f-40de-ab20-fef493899944
7ed725ce-e516-4386-bc6a-0b16bbbac678
6ec06ad1-0416-4e0a-9c2c-0b4381976091
a04087d6-4d95-4d99-901f-a1ff8578a2bf
5146be6c-ffda-401c-8721-3c43c7370872
dc2be5c1-2b6d-47d6-9a45-c188fd96d124`)
		fakeS3.PutObject("test/jobs/obj1.csv", &fakeData1)
		fakeS3.PutObject("test/jobs/obj2.csv", &fakeData2)
		fakeS3.PutObject("test/jobs/obj3.csv", &fakeData3)
		fakeS3.PutObject("test/jobs/obj4.csv", &fakeData4)
		fakeS3.PutObject("test/jobs/obj5.csv", &fakeData5)
		app = CreateTestApp(w.MarathonDB.DB)
		defaults := map[string]interface{}{
			"user_name":   "Someone",
			"object_name": "village",
		}
		body := map[string]interface{}{
			"alert": "{{user_name}} just liked your {{object_name}}!",
		}
		template = CreateTestTemplate(w.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": defaults,
			"body":     body,
			"locale":   "en",
		})
		context = map[string]interface{}{
			"user_name": "Everyone",
		}
		CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
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
		w.RedisClient.FlushAll()
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
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
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
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesWorker.Process(msg) }).Should(Panic())
		})

		It("should not panic if csvPath and jobID are valid", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj2.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			createBatchesWorker.Process(msg)
		})

		It("should work if CSV is from Excel/Windows", func() {
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj4.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			createCSVSplitWorker.Process(msg)

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			createBatchesWorker.Process(msg)
		})

		It("should do nothing if job status is stopped", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			_, err = w.MarathonDB.DB.Model(&model.Job{}).Set("status = 'stopped'").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(0))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			wMessage2, err := worker.ParseProcessBatchWorkerMessageArray(j2["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users) + len(wMessage2.Users)).To(BeEquivalentTo(10))
		})

		It("should create batches with the right number of tokens if a controlGroup is specified", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context":      context,
				"filters":      map[string]interface{}{},
				"csvPath":      "test/jobs/obj5.csv",
				"controlGroup": 0.2,
			})
			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(1))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(8))
		})

		It("should create batches with the right number of tokens if a controlGroup is specified", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context":      context,
				"filters":      map[string]interface{}{},
				"csvPath":      "test/jobs/obj5.csv",
				"controlGroup": 0.4,
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(1))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(6))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if numPushes < dbPageSize", func() {
			w.DBPageSize = 500
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			wMessage2, err := worker.ParseProcessBatchWorkerMessageArray(j2["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users) + len(wMessage2.Users)).To(BeEquivalentTo(10))
		})

		It("should skip batches if startsAt is past and pastTimeStrategy is skip", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context":          context,
				"filters":          map[string]interface{}{},
				"csvPath":          "test/jobs/obj1.csv",
				"localized":        true,
				"startsAt":         time.Now().UTC().Add(-12 * time.Hour).UnixNano(),
				"pastTimeStrategy": "skip",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			res, err := w.RedisClient.ZCard("schedule").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(0))
		})

		It("should delay batches to next day if startsAt is past and pastTimeStrategy is nextDay", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context":          context,
				"filters":          map[string]interface{}{},
				"csvPath":          "test/jobs/obj1.csv",
				"localized":        true,
				"startsAt":         time.Now().UTC().Add(time.Duration(-6) * time.Hour).UnixNano(),
				"pastTimeStrategy": "nextDay",
			})
			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.ZCard("schedule").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			var data workers.EnqueueData
			jobs, err := w.RedisClient.ZRange("schedule", 0, 2).Result()
			bytes, err := RedisReplyToBytes(jobs[0], err)
			Expect(err).NotTo(HaveOccurred())
			json.Unmarshal(bytes, &data)
			pushTime := time.Unix(0, int64(data.At*workers.NanoSecondPrecision))
			Expect(pushTime.After(time.Now())).To(Equal(true))
		})

		It("should schedule process_batches_worker if push is localized and starts in future", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context":   context,
				"filters":   map[string]interface{}{},
				"csvPath":   "test/jobs/obj1.csv",
				"localized": true,
				"startsAt":  time.Now().UTC().Add(12 * time.Hour).UnixNano(),
			})
			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.ZCard("schedule").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			var data workers.EnqueueData
			jobs, err := w.RedisClient.ZRange("schedule", 0, 2).Result()
			bytes, err := RedisReplyToBytes(jobs[0], err)
			Expect(err).NotTo(HaveOccurred())
			json.Unmarshal(bytes, &data)
			pushTime := time.Unix(0, int64(data.At*workers.NanoSecondPrecision))
			Expect(pushTime.After(time.Now())).To(Equal(true))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if a filter has multiple values separated bt comma", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{
					"locale": "pt,en",
				},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			wMessage2, err := worker.ParseProcessBatchWorkerMessageArray(j2["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users) + len(wMessage2.Users)).To(BeEquivalentTo(10))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if service is gcm", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
				"service": "gcm",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			wMessage2, err := worker.ParseProcessBatchWorkerMessageArray(j2["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users) + len(wMessage2.Users)).To(BeEquivalentTo(9))
		})

		It("should not panic if job is a reexecution", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
				"service": "gcm",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(2))
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			job2, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			j2 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal([]byte(job2), &j2)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			Expect(j2["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			wMessage2, err := worker.ParseProcessBatchWorkerMessageArray(j2["args"].([]interface{}))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(wMessage1.Users) + len(wMessage2.Users)).To(BeEquivalentTo(9))
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
		})

		It("should use job DBPageSize if specified", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			w.MarathonDB.DB.Model(j).Set("db_page_size = ?", 500).Returning("*").Update()

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			err = w.MarathonDB.DB.Model(j).Column("job.*", "App").Where("job.id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(j.DBPageSize).To(Equal(500))
		})

		It("should increment job totalBatches when no previous totalBatches", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalBatches).To(BeEquivalentTo(2))
		})

		It("should increment job totalBatches when previous totalBatches", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			_, err := w.MarathonDB.DB.Model(j).Set("total_batches = 4").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalBatches).To(BeEquivalentTo(6))
		})

		It("should update totalTokens and totalUsers correctly", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalUsers).To(BeEquivalentTo(10))
			Expect(job.TotalTokens).To(BeEquivalentTo(10))
		})

		It("should increment job totalTokens when no previous totalTokens", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(10))
		})

		It("should set totalTokens and totalUsers correctly", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj3.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(4))
			Expect(job.TotalUsers).To(BeEquivalentTo(2))
		})

		It("should increment job totalTokens when previous totalTokens", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj3.csv",
			})
			_, err := w.MarathonDB.DB.Model(j).Set("total_tokens = 4").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.DB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(8))
		})

		It("should not panic and change job status to stopped if bad csv", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj2.csv",
			})

			_, err := w.CSVSplitJob(j.ID.String())
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = workers.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			updatedJob := &model.Job{
				ID: j.ID,
			}
			err = w.MarathonDB.DB.Model(updatedJob).Column("job.*").Where("job.id = ?", updatedJob.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedJob.Status).To(Equal("stopped"))
		})
	})

	// Describe("Read CSV from S3", func() {
	// 	It("should return correct array from Unix csv data", func() {
	// 		res := w.ReadCSVFromS3("test/jobs/obj3.csv")
	// 		Expect(res).To(HaveLen(2))
	// 	})

	// 	It("should return correct array from DOS csv data", func() {
	// 		res := w.ReadCSVFromS3("test/jobs/obj4.csv")
	// 		Expect(res).To(HaveLen(2))
	// 	})
	// })
})
