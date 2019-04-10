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

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("JobCompleted Worker", func() {
	var jobCompletedWorker *worker.JobCompletedWorker
	var app *model.App
	var template *model.Template
	var job *model.Job

	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	w := worker.NewWorker(logger, GetConfPath())
	fakeS3 := NewFakeS3(w.Config)
	w.S3Client = fakeS3

	BeforeEach(func() {
		jobCompletedWorker = worker.NewJobCompletedWorker(w)

		app = CreateTestApp(w.MarathonDB.DB)
		template = CreateTestTemplate(w.MarathonDB.DB, app.ID)
		job = CreateTestJob(w.MarathonDB.DB, app.ID, template.Name)
	})

	Describe("Process", func() {
		It("should process when job is retrieved correctly", func() {
			messageObj := []interface{}{job.ID.String()}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() {
				jobCompletedWorker.Process(message)
			}).ShouldNot(Panic())
		})

		It("should not process when job is not found in db", func() {
			_, err := w.MarathonDB.DB.Exec("DELETE FROM jobs;")
			Expect(err).NotTo(HaveOccurred())
			messageObj := []interface{}{job.ID.String()}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())

			Expect(func() { jobCompletedWorker.Process(message) }).Should(Panic())
		})
	})
})
