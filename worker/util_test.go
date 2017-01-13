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

	workers "github.com/jrallison/go-workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	"github.com/topfreegames/marathon/worker"
)

var _ = Describe("Worker Util", func() {
	var template model.Template
	var users []worker.User
	var usersObj []interface{}
	var jobID string
	var appName string
	BeforeEach(func() {
		template = model.Template{
			Body: map[string]interface{}{
				"alert": "{{user_name}} just liked your {{object_name}}!",
			},
			Defaults: map[string]interface{}{
				"user_name":   "Someone",
				"object_name": "village",
			},
		}

		users = make([]worker.User, 2)
		usersObj = make([]interface{}, 2)
		for index, _ := range users {
			id := uuid.NewV4().String()
			token := strings.Replace(uuid.NewV4().String(), "-", "", -1)
			users[index] = worker.User{
				UserID: id,
				Token:  token,
				Locale: "en",
			}
			usersObj[index] = map[string]interface{}{
				"user_id": id,
				"token":   token,
				"locale":  "en",
			}
		}

		appName = strings.Split(uuid.NewV4().String(), "-")[0]
		jobID = uuid.NewV4().String()
	})

	Describe("Check if userID is valid", func() {
		It("should reject userID with double quote", func() {
			valid := worker.IsUserIDValid("+.\"Paco\".+")
			Expect(valid).To(BeFalse())
		})
		It("should reject userID with single quote", func() {
			valid := worker.IsUserIDValid("+.'Paco'.+")
			Expect(valid).To(BeFalse())
		})
		It("should reject userID with comma", func() {
			valid := worker.IsUserIDValid("+,Paco,+")
			Expect(valid).To(BeFalse())
		})
		It("should accept userID with no invalid characters", func() {
			valid := worker.IsUserIDValid("b64398fd8f460376d9536a39e48009ef39d218bd3")
			Expect(valid).To(BeTrue())
		})
	})

	Describe("Build message from template", func() {
		It("should make correct substitutions using defaults", func() {
			context := map[string]interface{}{}
			msgString, err := worker.BuildMessageFromTemplate(template, context)
			Expect(err).NotTo(HaveOccurred())
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
			msgString, err := worker.BuildMessageFromTemplate(template, context)
			Expect(err).NotTo(HaveOccurred())
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
			msgString, err := worker.BuildMessageFromTemplate(template, context)
			Expect(err).NotTo(HaveOccurred())
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
			messageObj := []interface{}{
				jobID,
				appName,
				usersObj,
			}
			msgB, err := json.Marshal(map[string][]interface{}{
				"args": messageObj,
			})
			Expect(err).NotTo(HaveOccurred())

			message, err := workers.NewMsg(string(msgB))
			Expect(err).NotTo(HaveOccurred())
			arr, err := message.Args().Array()
			Expect(err).NotTo(HaveOccurred())

			parsed, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.JobID.String()).To(Equal(jobID))
			Expect(parsed.AppName).To(Equal(appName))
			Expect(len(parsed.Users)).To(Equal(len(users)))

			for idx, user := range users {
				Expect(parsed.Users[idx]).To(Equal(user))
			}
		})

		It("should fail if array has less than 3 elements", func() {
			arr := []interface{}{jobID, appName}
			_, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(worker.InvalidMessageArray))
		})

		It("should fail if array has more than 3 elements", func() {
			arr := []interface{}{jobID, appName, usersObj, usersObj}
			_, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(worker.InvalidMessageArray))
		})

		It("should fail if jobID is not uuid", func() {
			arr := []interface{}{"some-string", appName, usersObj}
			_, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("uuid: UUID string too short"))
		})

		It("should fail if users is not array", func() {
			arr := []interface{}{jobID, appName, "some-string"}
			Expect(func() { worker.ParseProcessBatchWorkerMessageArray(arr) }).Should(Panic())
		})

		It("should fail if users is an empty array", func() {
			emptyUsers := []interface{}{}
			arr := []interface{}{jobID, appName, emptyUsers}
			_, err := worker.ParseProcessBatchWorkerMessageArray(arr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("there must be at least one user"))
		})
	})

	Describe("Get Clause From Filters", func() {
		It("should return empty string if filters is empty", func() {
			filters := map[string]interface{}{}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(Equal(""))
		})

		It("should succeed with one simple filter", func() {
			filters := map[string]interface{}{
				"region": "US",
			}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(Equal("\"region\"='US'"))
		})

		It("should succeedd with one comma separated filter", func() {
			filters := map[string]interface{}{
				"region": "US,CA",
			}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(Equal("(\"region\"='US' OR \"region\"='CA')"))
		})

		It("should succeed with one negative simple filter", func() {
			filters := map[string]interface{}{
				"NOTregion": "US",
			}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(Equal("\"region\"!='US'"))
		})

		It("should succeed with one negative comma separated filter", func() {
			filters := map[string]interface{}{
				"NOTregion": "US,CA",
			}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(Equal("(\"region\"!='US' AND \"region\"!='CA')"))
		})

		It("should succeed with multiple filters", func() {
			filters := map[string]interface{}{
				"NOTregion": "US,CA",
				"locale":    "en,fr",
			}
			where := worker.GetWhereClauseFromFilters(filters)
			Expect(where).To(ContainSubstring("(\"locale\"='en' OR \"locale\"='fr')"))
			Expect(where).To(ContainSubstring("(\"region\"!='US' AND \"region\"!='CA')"))
			Expect(where).To(ContainSubstring(") AND ("))
		})
	})
})
