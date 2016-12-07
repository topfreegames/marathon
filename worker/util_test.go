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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	"github.com/topfreegames/marathon/worker"
)

var _ = Describe("Worker Util", func() {
	var err error
	var template *model.Template
	var templateStr string
	var context map[string]interface{}
	var contextStr string
	var users []model.User
	var usersStr string
	var jobID string
	BeforeEach(func() {
		template = &model.Template{
			Body: map[string]string{
				"alert": "{{user_name}} just liked your {{object_name}}!",
			},
			Defaults: map[string]string{
				"user_name":   "Someone",
				"object_name": "village",
			},
		}

		tmp, err := json.Marshal(template)
		Expect(err).NotTo(HaveOccurred())
		templateStr = string(tmp)

		context = map[string]interface{}{
			"user_name":   "Camila",
			"object_name": "building",
		}
		tmp, err = json.Marshal(context)
		Expect(err).NotTo(HaveOccurred())
		contextStr = string(tmp)

		users = make([]model.User, 2)
		for index, _ := range users {
			users[index] = model.User{
				ID:    uuid.NewV4(),
				Token: strings.Replace(uuid.NewV4().String(), "-", "", -1),
			}
		}
		tmp, err = json.Marshal(users)
		Expect(err).NotTo(HaveOccurred())
		usersStr = string(tmp)

		jobID = uuid.NewV4().String()
	})

	Describe("Build message from template", func() {
		It("should make correct substitutions using defaults", func() {
			context := map[string]interface{}{}
			msgString := worker.BuildMessageFromTemplate(template, context)
			var msg map[string]interface{}
			err = json.Unmarshal([]byte(msgString), &msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(msg["alert"]).To(Equal("Someone just liked your village!"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{user_name}}"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{object_name}}"))
		})

		It("should make correct substitutions using context", func() {
			context := map[string]interface{}{
				"user_name":   "Camila",
				"object_name": "building",
			}
			msgString := worker.BuildMessageFromTemplate(template, context)
			var msg map[string]interface{}
			err = json.Unmarshal([]byte(msgString), &msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(msg["alert"]).To(Equal("Camila just liked your building!"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{user_name}}"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{object_name}}"))
		})

		It("should make correct substitutions mixing defaults and context", func() {
			context := map[string]interface{}{
				"user_name": "Camila",
			}
			msgString := worker.BuildMessageFromTemplate(template, context)
			var msg map[string]interface{}
			err = json.Unmarshal([]byte(msgString), &msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(msg["alert"]).To(Equal("Camila just liked your village!"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{user_name}}"))
			Expect(msg["alert"]).NotTo(ContainSubstring("{{object_name}}"))
		})
	})

	Describe("Parse ProcessBatchWorker message array", func() {
		It("should succeed if all params are correct", func() {
			arr := []interface{}{jobID, templateStr, contextStr, usersStr}
			parsedJobID, parsedTemplate, parsedContext, parsedUsers, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedJobID.String()).To(Equal(jobID))
			Expect(parsedTemplate.Body).To(Equal(template.Body))
			Expect(parsedTemplate.Defaults).To(Equal(template.Defaults))
			Expect(parsedContext).To(Equal(context))
			Expect(len(parsedUsers)).To(Equal(len(users)))

			for idx, user := range users {
				Expect(parsedUsers[idx]).To(Equal(user))
			}
		})

		It("should fail if array has less than 4 elements", func() {
			arr := []interface{}{jobID, templateStr, contextStr}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("array must be of the form [jobId, template, context, users]"))
		})

		It("should fail if array has more than 4 elements", func() {
			arr := []interface{}{jobID, templateStr, contextStr, usersStr, usersStr}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("array must be of the form [jobId, template, context, users]"))
		})

		It("should fail if jobID is not uuid", func() {
			arr := []interface{}{"some-string", templateStr, contextStr, usersStr}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("uuid: UUID string too short"))
		})

		It("should fail if template is not json", func() {
			arr := []interface{}{jobID, "some-string", contextStr, usersStr}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})

		It("should fail if context is not json", func() {
			arr := []interface{}{jobID, templateStr, "some-string", usersStr}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})

		It("should fail if users is not array", func() {
			arr := []interface{}{jobID, templateStr, contextStr, "some-string"}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})

		It("should fail if users is an empty array", func() {
			arr := []interface{}{jobID, templateStr, contextStr, "[]"}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("there must be at least one user"))
		})

		It("should fail if user has bad ID", func() {
			arr := []interface{}{jobID, templateStr, contextStr, "[{\"id\": \"whatever\", \"token\": \"whatever\"}]"}
			_, _, _, _, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("uuid: UUID string too short"))
		})
	})
})
