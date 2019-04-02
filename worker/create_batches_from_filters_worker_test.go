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
	"fmt"
	"strings"

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("CreateBatchesFromFilters Worker", func() {
	var app *model.App
	var template *model.Template

	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()),
		zap.FatalLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())
	createBatchesFromFiltersWorker := worker.NewCreateBatchesFromFiltersWorker(w)

	BeforeEach(func() {
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
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).To(Panic())
		})

		It("should do nothing if job status is stopped", func() {
			jobOptions := map[string]interface{}{
				"filters": map[string]interface{}{},
			}
			j := CreateTestJob(w.MarathonDB.DB, app.ID, template.Name, jobOptions)
			_, err := w.MarathonDB.DB.Model(&model.Job{}).Set("status = 'stopped'").Where("id = ?", j.ID).Update()
			Expect(err).NotTo(HaveOccurred())
			m := map[string]interface{}{
				"jid":  3,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
			key := fmt.Sprintf("%s/job-%s.csv", w.Config.GetString("s3.folder"), j.ID)
			fakeS3 := NewFakeS3(w.Config)
			_, err = fakeS3.GetObject(key)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("NoSuchKey: The specified key does not exist. status code: 404"))
		})

		//TODO fazer wg wait do createBatchesWoker (que reporta numero)
		It("should not panic if job.ID is valid and filters are not empty", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(
				w.MarathonDB.DB,
				a.ID,
				template.Name,
				map[string]interface{}{"filters": map[string]interface{}{"locale": "en"}},
			)
			w.S3Client = NewFakeS3(w.Config)
			m := map[string]interface{}{
				"jid":  4,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
		})

		It("should not panic if job.ID is valid and filters are not empty", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			w.S3Client = NewFakeS3(w.Config)
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
					"tz":     "-0500",
				},
			})
			m := map[string]interface{}{
				"jid":  5,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
		})

		It("should panic if no users match the filters", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "n",
				},
			})
			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3
			m := map[string]interface{}{
				"jid":  6,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).To(Panic())
		})

		It("should generate the correct job with the right number of users", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
			})
			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3
			m := map[string]interface{}{
				"jid":  6,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:db_to_csv_worker").Result()

			Expect(res).To(BeEquivalentTo(1))
			j1 := map[string]interface{}{}
			job1, err := w.RedisClient.LPop("queue:db_to_csv_worker").Result()
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("db_to_csv_worker"))
		})

		It("should update job's csvPath correctly", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "PT",
					"tz":     "-0300",
				},
				"service": "gcm",
			})
			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3
			m := map[string]interface{}{
				"jid":  6,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
			key := fmt.Sprintf("%s/job-%s.csv", w.Config.GetString("s3.folder"), j.ID)
			dbJob := &model.Job{
				ID: j.ID,
			}
			err = w.MarathonDB.DB.Model(&dbJob).Column("csv_path").Where("id = ?", j.ID.String()).Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(dbJob.CSVPath).To(Equal(key))
		})

		It("should enqueue a db_to_csv with the right job", func() {
			a := CreateTestApp(w.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(w.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "PT",
					"tz":     "-0300",
				},
				"service": "gcm",
			})
			fakeS3 := NewFakeS3(w.Config)
			w.S3Client = fakeS3
			m := map[string]interface{}{
				"jid":  6,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
			res, err := w.RedisClient.LLen("queue:db_to_csv_worker").Result()
			Expect(res).To(BeEquivalentTo(1))
			j1 := map[string]interface{}{}
			job1, err := w.RedisClient.LPop("queue:db_to_csv_worker").Result()
			err = json.Unmarshal([]byte(job1), &j1)
			Expect(err).NotTo(HaveOccurred())
			Expect(j1["queue"].(string)).To(Equal("db_to_csv_worker"))
			Expect(j1["args"].(map[string]interface{})["Job"].(map[string]interface{})["id"]).To(Equal(j.ID.String()))

		})
	})

})
