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

var _ = Describe("CreateBatchesFromFilters Worker", func() {
	var logger zap.Logger
	var config *viper.Viper
	var createBatchesFromFiltersWorker *worker.CreateBatchesFromFiltersWorker
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
		createBatchesFromFiltersWorker = worker.NewCreateBatchesFromFiltersWorker(config, logger, w)
		app = CreateTestApp(createBatchesFromFiltersWorker.MarathonDB.DB)
		defaults := map[string]interface{}{
			"user_name":   "Someone",
			"object_name": "village",
		}
		body := map[string]interface{}{
			"alert": "{{user_name}} just liked your {{object_name}}!",
		}
		template = CreateTestTemplate(createBatchesFromFiltersWorker.MarathonDB.DB, app.ID, map[string]interface{}{
			"defaults": defaults,
			"body":     body,
			"locale":   "en",
		})
		context = map[string]interface{}{
			"user_name": "Everyone",
		}
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
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).Should(Panic())
		})

		It("should panic if filters has 0 len", func() {
			jobOptions := map[string]interface{}{
				"filters": map[string]interface{}{},
			}
			j := CreateTestJob(createBatchesFromFiltersWorker.MarathonDB.DB, app.ID, template.Name, jobOptions)
			m := map[string]interface{}{
				"jid":  3,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).Should(Panic())
		})

		It("should not panic if job.ID is valid and filters are not empty", func() {
			a := CreateTestApp(createBatchesFromFiltersWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(createBatchesFromFiltersWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{})
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
			a := CreateTestApp(createBatchesFromFiltersWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(createBatchesFromFiltersWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
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

		It("should generate the right number of batches", func() {
			a := CreateTestApp(createBatchesFromFiltersWorker.MarathonDB.DB, map[string]interface{}{"name": "testapp"})
			j := CreateTestJob(createBatchesFromFiltersWorker.MarathonDB.DB, a.ID, template.Name, map[string]interface{}{
				"filters": map[string]interface{}{
					"locale": "en",
				},
			})
			m := map[string]interface{}{
				"jid":  6,
				"args": []string{j.ID.String()},
			}
			smsg, err := json.Marshal(m)
			Expect(err).NotTo(HaveOccurred())
			msg, err := workers.NewMsg(string(smsg))
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { createBatchesFromFiltersWorker.Process(msg) }).ShouldNot(Panic())
		})
	})
})
