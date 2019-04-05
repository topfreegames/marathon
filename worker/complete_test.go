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
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var letterRunes = []rune("-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var _ = Describe("Complete Test Worker", func() {
	var app *model.App
	var template *model.Template
	var producer *FakeKafkaProducer
	config := GetConf()

	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()),
		zap.FatalLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())

	createBatchesFromFiltersWorker := worker.NewCreateBatchesFromFiltersWorker(w)
	createDbToCsvBatchesWorker := worker.NewDbToCsvWorker(w)

	createCSVSplitWorker := worker.NewCSVSplitWorker(w)
	createBatchesWorker := worker.NewCreateBatchesWorker(w)
	processBatchWorker := worker.NewProcessBatchWorker(w)
	jobCompleteWorker := worker.NewJobCompletedWorker(w)

	rand.Seed(42)

	runAllSteps := func(jobID string) {
		m := map[string]interface{}{
			"jid":  1,
			"args": []string{jobID},
		}
		smsg, err := json.Marshal(m)
		Expect(err).NotTo(HaveOccurred())
		msg, err := workers.NewMsg(string(smsg))
		Expect(err).NotTo(HaveOccurred())
		createBatchesFromFiltersWorker.Process(msg)

		dataSlice, err := w.RedisClient.LRange("queue:db_to_csv_worker", 0, -1).Result()
		Expect(err).NotTo(HaveOccurred())
		for _, data := range dataSlice {
			msg, err = workers.NewMsg(data)
			Expect(err).NotTo(HaveOccurred())
			createDbToCsvBatchesWorker.Process(msg)
		}

		dataSlice, err = w.RedisClient.LRange("queue:csv_split_worker", 0, -1).Result()
		Expect(err).NotTo(HaveOccurred())
		for _, data := range dataSlice {
			msg, err = workers.NewMsg(data)
			Expect(err).NotTo(HaveOccurred())
			createCSVSplitWorker.Process(msg)
		}

		dataSlice, err = w.RedisClient.LRange("queue:create_batches_worker", 0, -1).Result()
		Expect(err).NotTo(HaveOccurred())
		for _, data := range dataSlice {
			msg, err = workers.NewMsg(data)
			Expect(err).NotTo(HaveOccurred())
			createBatchesWorker.Process(msg)
		}

		dataSlice, err = w.RedisClient.LRange("queue:process_batch_worker", 0, -1).Result()
		Expect(err).NotTo(HaveOccurred())
		for _, data := range dataSlice {
			msg, err = workers.NewMsg(data)
			Expect(err).NotTo(HaveOccurred())
			processBatchWorker.Process(msg)
		}

		dataSlice, err = w.RedisClient.ZRange("schedule", 0, -1).Result()
		Expect(err).NotTo(HaveOccurred())
		for _, data := range dataSlice {
			msg, err = workers.NewMsg(data)
			Expect(err).NotTo(HaveOccurred())
			jobCompleteWorker.Process(msg)
		}
	}

	BeforeEach(func() {
		app = CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "myapp"})
		template = CreateTestTemplate(w.MarathonDB.DB, app.ID, map[string]interface{}{
			"locale": "en",
		})

		w.PushDB.DB.Query(nil, `
				DROP TABLE myapp_apns;
			`)
		w.PushDB.DB.Query(nil, `
				CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
			`)
		w.PushDB.DB.Query(nil, `
				CREATE TABLE IF NOT EXISTS "myapp_apns" (
				  "id" uuid DEFAULT uuid_generate_v4(),
    			  "seq_id" integer UNIQUE NOT NULL,
				  "user_id" text NOT NULL,
				  "token" text NOT NULL,
				  "region" text NOT NULL,
				  "locale" text NOT NULL,
				  "tz" text NOT NULL,
				  PRIMARY KEY ("id")
				);
			`)

		w.RedisClient.FlushAll()
		w.S3Client = NewFakeS3(w.Config)
		producer = NewFakeKafkaProducer()
		w.Kafka = producer
	})

	Describe("Process", func() {
		It("create 10000 queries test", func() {
			_, err := w.PushDB.DB.Query(nil, `
				INSERT INTO myapp_apns (seq_id, user_id, token, locale, region, tz)
				SELECT
					generate_series(1, 10000) AS seq_id,
					encode(gen_random_bytes(floor(random() * 60 + 1)::int), 'hex') as user_id,
					encode(gen_random_bytes(60), 'hex') AS token,
					'en' as locale,
					'us' as region,
					'+0000' as tz;
			`)
			Expect(err).NotTo(HaveOccurred())

			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
			})
			runAllSteps(j.ID.String())

			Expect(len(producer.APNSMessages)).To(Equal(10000))
		})

		It("create 1000 queries with the same user_id", func() {
			_, err := w.PushDB.DB.Query(nil, `
				INSERT INTO myapp_apns (seq_id, user_id, token, locale, region, tz)
				SELECT
					generate_series(1, 1000) AS seq_id,
					'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' as user_id,
					encode(gen_random_bytes(60), 'hex') AS token,
					'en' as locale,
					'us' as region,
					'+0000' as tz;
			`)
			Expect(err).NotTo(HaveOccurred())

			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
			})

			runAllSteps(j.ID.String())

			Expect(len(producer.APNSMessages)).To(Equal(1000))
		})

		It("should put control group in s3 and also update job with controlGroupCSVPath", func() {
			_, err := w.PushDB.DB.Query(nil, `
				INSERT INTO myapp_apns (seq_id, user_id, token, locale, region, tz)
				VALUES
				(1, '1', '1', 'en', 'us', '+0000'),
				(10, '10', '10', 'en', 'us', '+0000'),
				(2, '2', '2', 'en', 'us', '+0000'),
				(20, '20', '20', 'en', 'us', '+0000'),
				(3, '3', '3', 'en', 'us', '+0000'),
				(30, '30', '30', 'en', 'us', '+0000'),
				(4, '4', '4', 'en', 'us', '+0000'),
				(40, '40', '40', 'en', 'us', '+0000'),
				(5, '5', '5', 'en', 'us', '+0000'),
				(50, '50', '50', 'en', 'us', '+0000');
			`)
			Expect(err).NotTo(HaveOccurred())

			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
				"csvPath":      "test/jobs/obj5.csv",
				"controlGroup": 0.4,
			})

			runAllSteps(j.ID.String())

			bucket := config.GetString("s3.bucket")
			key := fmt.Sprintf("%s/%s/job-%s.csv", bucket, w.Config.GetString("s3.controlGroupFolder"), j.ID.String())
			controlGroupCSV, err := w.S3Client.GetObject(key)
			Expect(err).NotTo(HaveOccurred())
			lines := ReadLinesFromIOReader(bytes.NewReader(controlGroupCSV))
			Expect(len(lines)).To(Equal(5)) //5 -> header + 4 control group userIds

			dbJob := &model.Job{
				ID: j.ID,
			}
			err = w.MarathonDB.DB.Model(&dbJob).Column("control_group_csv_path").Where("id = ?", j.ID.String()).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.ControlGroupCSVPath).To(Equal(key))
		})
	})

})
