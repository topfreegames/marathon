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

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

var _ = Describe("ProcessBatch Worker", func() {
	var logger zap.Logger
	var config *viper.Viper
	var resumeJobWorker *worker.ResumeJobWorker
	var app *model.App
	var template *model.Template
	var job *model.Job
	var users []worker.User
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		config = GetConf()
		w := worker.NewWorker(false, logger, GetConfPath())
		resumeJobWorker = worker.NewResumeJobWorker(config, logger, w)
		resumeJobWorker.RedisClient.FlushAll()

		app = CreateTestApp(resumeJobWorker.MarathonDB.DB)
		appName := strings.Split(app.BundleID, ".")[2]
		template = CreateTestTemplate(resumeJobWorker.MarathonDB.DB, app.ID)
		job = CreateTestJob(resumeJobWorker.MarathonDB.DB, app.ID, template.Name)
		users = make([]worker.User, 10)
		for index, _ := range users {
			id := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			user := worker.User{
				UserID: id,
				Token:  token,
				Locale: "en",
			}
			users[index] = user
			messageObj := []interface{}{
				job.ID,
				appName,
				[]worker.User{user},
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())
			_, err = resumeJobWorker.RedisClient.RPush(fmt.Sprintf("%s-pausedjobs", job.ID.String()), message.ToJson()).Result()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Process", func() {
		It("should remove jobs from the paused jobs list until it is empty", func() {
			messageObj := []interface{}{
				job.ID,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())
			resumeJobWorker.Process(message)

			res, err := resumeJobWorker.RedisClient.LLen("queue:process_batch_worker").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeEquivalentTo(len(users)))

			remainingJobsLen, err := resumeJobWorker.RedisClient.LLen(fmt.Sprintf("%s-pausedjobs", job.ID.String())).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(remainingJobsLen).To(Equal(int64(0)))
		})
	})
})
