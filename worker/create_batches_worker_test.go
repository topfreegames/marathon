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

	goworkers2 "github.com/digitalocean/go-workers2"
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
		zap.DebugLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())
	createBatchesWorker := worker.NewCreateBatchesWorker(w)
	createCSVSplitWorker := worker.NewCSVSplitWorker(w)
	//processBatchWorker := worker.NewProcessBatchWorker(w)
	//processBatchWorkerConcurrency := w.Config.GetInt("workers.processBatch.concurrency")

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
user_id
idisnotanuuidanditssizecanvaryforeachuser
gamesendsanuseridthatisnotanuuid
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
		fakeData6 := []byte(`userIds
stange-token`)

		// Ids that do not exist
		fakeData7 := []byte(`userIds
d6333e62-2778-463c-b7d6-4d99aab04fb8
438d72f7-3e2f-4439-be9d-ee48b9ba2e76`)
		fakeData8 := []byte(`userIds
9e558649-9c23-469d-a11c-59b05813e3d5
57be9009-e616-42c6-9cfe-505508ede2d0
a8e8d2d5-f178-4d90-9b31-683ad3aae920
5c3033c0-24ad-487a-a80d-68432464c8de`)
		// Ids are not uuid
		fakeData9 := []byte(`userIds
useridisnotanuuid
gamesendsanuseridthatisnotanuuid
idisnotanuuidanditssizecanvaryforeachuser
user_id
`)

		fakeS3.PutObject("test/jobs/obj1.csv", &fakeData1)
		fakeS3.PutObject("test/jobs/obj2.csv", &fakeData2)
		fakeS3.PutObject("test/jobs/obj3.csv", &fakeData3)
		fakeS3.PutObject("test/jobs/obj4.csv", &fakeData4)
		fakeS3.PutObject("test/jobs/obj5.csv", &fakeData5)
		fakeS3.PutObject("test/jobs/obj6.csv", &fakeData6)
		fakeS3.PutObject("test/jobs/obj7.csv", &fakeData7)
		fakeS3.PutObject("test/jobs/obj8.csv", &fakeData8)
		fakeS3.PutObject("test/jobs/obj9.csv", &fakeData9)
		app = CreateTestApp(w.MarathonDB)
		defaults := map[string]interface{}{
			"user_name":   "Someone",
			"object_name": "village",
		}
		body := map[string]interface{}{
			"alert": "{{user_name}} just liked your {{object_name}}!",
		}
		template = CreateTestTemplate(w.MarathonDB, app.ID, map[string]interface{}{
			"defaults": defaults,
			"body":     body,
			"locale":   "en",
		})
		context = map[string]interface{}{
			"user_name": "Everyone",
		}
		CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
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
			msg, err := goworkers2.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesWorker.Process(msg) }).Should(Panic())
		})

		It("should panic if csvPath is invalid", func() {
			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
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
			msg, err := goworkers2.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesWorker.Process(msg) }).Should(Panic())
		})

		It("should not panic if csvPath and jobID are valid", func() {
			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj2.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			createBatchesWorker.Process(msg)
		})

		It("should not panic if csv was only one ID", func() {
			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj6.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			createBatchesWorker.Process(msg)
		})

		It("should work if CSV is from Excel/Windows", func() {
			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj4.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			createCSVSplitWorker.Process(msg)

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			createBatchesWorker.Process(msg)
		})

		It("should do nothing if job status is stopped", func() {
			//
			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			_, err = w.MarathonDB.Model(&model.Job{}).Set("status = 'stopped'").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(0))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(10))
		})

		It("should create batches with the right number of tokens if a controlGroup is specified", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context":      context,
				"filters":      map[string]interface{}{},
				"csvPath":      "test/jobs/obj5.csv",
				"controlGroup": 0.2,
			})
			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context":      context,
				"filters":      map[string]interface{}{},
				"csvPath":      "test/jobs/obj5.csv",
				"controlGroup": 0.4,
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(10))
		})

		It("should skip batches if startsAt is past and pastTimeStrategy is skip", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context":          context,
				"filters":          map[string]interface{}{},
				"csvPath":          "test/jobs/obj1.csv",
				"localized":        true,
				"startsAt":         time.Now().UTC().Add(-12 * time.Hour).UnixNano(),
				"pastTimeStrategy": "skip",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			res, err := w.RedisClient.ZCard("schedule").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(0))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if a filter has multiple values separated bt comma", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{
					"locale": "pt,en",
				},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(10))
		})

		It("should create batches with the right tokens and tz and send to process_batches_worker if service is gcm", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
				"service": "gcm",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(9))
		})

		It("should not panic if job is a reexecution", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
				"service": "gcm",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
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
			Expect(len(wMessage1.Users)).To(BeEquivalentTo(9))
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
		})

		It("should use job DBPageSize if specified", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			w.MarathonDB.Model(j).Set("db_page_size = ?", 500).Returning("*").Update()

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			err = w.MarathonDB.Model(j).Column("job.*", "App").Where("job.id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(j.DBPageSize).To(Equal(500))
		})

		It("should increment job totalBatches when no previous totalBatches", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalBatches).To(BeEquivalentTo(1))
		})

		It("should increment job totalBatches when previous totalBatches", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			_, err := w.MarathonDB.Model(j).Set("total_batches = 4").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalBatches).To(BeEquivalentTo(5))
		})

		It("should update totalTokens and totalUsers correctly", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalUsers).To(BeEquivalentTo(10))
			Expect(job.TotalTokens).To(BeEquivalentTo(10))
		})

		It("should increment job totalTokens when no previous totalTokens", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj1.csv",
			})
			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(jobData)
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(jobData)
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(10))
		})

		It("should set totalTokens and totalUsers correctly", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj3.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(4))
			Expect(job.TotalUsers).To(BeEquivalentTo(4))
		})

		It("should increment job totalTokens when previous totalTokens", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj3.csv",
			})
			_, err := w.MarathonDB.Model(j).Set("total_tokens = 4").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())

			_, err = w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			job := &model.Job{}
			err = w.MarathonDB.Model(job).Where("id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(job.TotalTokens).To(BeEquivalentTo(8))
		})

		It("should not panic and change job status to stopped if bad csv", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj2.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			updatedJob := &model.Job{
				ID: j.ID,
			}
			err = w.MarathonDB.Model(updatedJob).Column("job.*").Where("job.id = ?", updatedJob.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedJob.Status).To(Equal("stopped"))
		})

		It("should finish if no id is known and add error message", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj7.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())

			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(0))

			err = w.MarathonDB.Model(j).Column("job.*", "App").Where("job.id = ?", j.ID).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(j.CompletedAt).ToNot(BeNil())
		})

		It("should not ignore first line if is not the first part of the file", func() {

			j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
				"context": context,
				"filters": map[string]interface{}{},
				"csvPath": "test/jobs/obj8.csv",
			})

			_, err := w.CreateCSVSplitJob(j)
			Expect(err).NotTo(HaveOccurred())

			jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err := goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			// it's the size of the header + first 2 id
			sizeLimit := 81
			totalParts := 2
			Expect(err).NotTo(HaveOccurred())
			createCSVSplitWorker.Workers.Config.Set("workers.csvSplitWorker.csvSizeLimitMB", float64(sizeLimit)/1024/1024)
			Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

			for i := 0; i < totalParts; i++ {
				jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
				Expect(err).NotTo(HaveOccurred())
				msg, err = goworkers2.NewMsg(string(jobData))
				Expect(err).NotTo(HaveOccurred())
				Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
			}

			res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			// two processed batches and 1 batch from split ids process
			Expect(res).To(BeEquivalentTo(3))

			userIds := map[string]string{}
			for i := 0; i < int(res); i++ {
				job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
				Expect(err).NotTo(HaveOccurred())
				j1 := map[string]interface{}{}
				err = json.Unmarshal([]byte(job1), &j1)
				Expect(err).NotTo(HaveOccurred())
				Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
				wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))

				for _, user := range wMessage1.Users {
					userIds[user.UserID] = user.UserID
				}

				Expect(err).NotTo(HaveOccurred())
			}
			Expect(len(userIds)).To(BeEquivalentTo(4))

			Expect(userIds["9e558649-9c23-469d-a11c-59b05813e3d5"]).To(BeEquivalentTo("9e558649-9c23-469d-a11c-59b05813e3d5"))
			Expect(userIds["57be9009-e616-42c6-9cfe-505508ede2d0"]).To(BeEquivalentTo("57be9009-e616-42c6-9cfe-505508ede2d0"))
			Expect(userIds["a8e8d2d5-f178-4d90-9b31-683ad3aae920"]).To(BeEquivalentTo("a8e8d2d5-f178-4d90-9b31-683ad3aae920"))
			Expect(userIds["5c3033c0-24ad-487a-a80d-68432464c8de"]).To(BeEquivalentTo("5c3033c0-24ad-487a-a80d-68432464c8de"))

		})
	})

	It("should process all ids even when split a file in middle of a line", func() {

		j := CreateTestJob(w.MarathonDB, app.ID, template.Name, map[string]interface{}{
			"context": context,
			"filters": map[string]interface{}{},
			"csvPath": "test/jobs/obj9.csv",
		})

		_, err := w.CreateCSVSplitJob(j)
		Expect(err).NotTo(HaveOccurred())

		jobData, err := w.RedisClient.LPop("queue:csv_split_worker").Result()
		Expect(err).NotTo(HaveOccurred())
		msg, err := goworkers2.NewMsg(string(jobData))
		Expect(err).NotTo(HaveOccurred())
		// it's the size of the header + part of the first id
		sizeLimit := 40
		totalParts := 3
		Expect(err).NotTo(HaveOccurred())
		createCSVSplitWorker.Workers.Config.Set("workers.csvSplitWorker.csvSizeLimitMB", float64(sizeLimit)/1024/1024)
		Expect(func() { createCSVSplitWorker.Process(msg) }).ShouldNot(Panic())

		for i := 0; i < totalParts; i++ {
			jobData, err = w.RedisClient.LPop("queue:create_batches_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			msg, err = goworkers2.NewMsg(string(jobData))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesWorker.Process(msg) }).ShouldNot(Panic())
		}

		res, err := w.RedisClient.LLen("queue:process_batch_worker").Result()
		Expect(err).NotTo(HaveOccurred())
		// two processed batches and 1 batch from split ids process
		Expect(res).To(BeEquivalentTo(3))

		userIds := map[string]string{}
		for i := 0; i < int(res); i++ {
			job1, err := w.RedisClient.LPop("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			j1 := map[string]interface{}{}
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("process_batch_worker"))
			wMessage1, err := worker.ParseProcessBatchWorkerMessageArray(j1["args"].([]interface{}))
			for _, user := range wMessage1.Users {
				userIds[user.UserID] = user.UserID
			}
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(userIds["user_id"]).To(BeEquivalentTo("user_id"))
		Expect(userIds["useridisnotanuuid"]).To(BeEquivalentTo("useridisnotanuuid"))
		Expect(userIds["gamesendsanuseridthatisnotanuuid"]).To(BeEquivalentTo("gamesendsanuseridthatisnotanuuid"))
		Expect(userIds["idisnotanuuidanditssizecanvaryforeachuser"]).To(BeEquivalentTo("idisnotanuuidanditssizecanvaryforeachuser"))
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
