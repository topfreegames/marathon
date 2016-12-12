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
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Job Handler", func() {
	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	app := GetDefaultTestApp(logger)
	faultyDb := GetFaultyTestDB(app)
	var existingApp *model.App
	var existingTemplate *model.Template
	var baseRoute string
	var baseRouteWithoutTemplate string
	BeforeEach(func() {
		app.DB.Exec("DELETE FROM apps;")
		app.DB.Exec("DELETE FROM templates;")
		existingApp = CreateTestApp(app.DB)
		existingTemplate = CreateTestTemplate(app.DB, existingApp.ID)
		baseRoute = fmt.Sprintf("/apps/%s/jobs?template=%s", existingApp.ID, existingTemplate.Name)
		baseRouteWithoutTemplate = fmt.Sprintf("/apps/%s/jobs", existingApp.ID)
	})

	Describe("Get /apps/:id/jobs?template=:templateName", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list if there are no jobs", func() {
				status, body := Get(app, baseRoute, "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of jobs with template", func() {
				testJobs := CreateTestJobs(app.DB, existingApp.ID, existingTemplate.Name, 10)
				anotherTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				CreateTestJobs(app.DB, existingApp.ID, anotherTemplate.Name, 10)
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, job := range response {
					Expect(job["id"]).ToNot(BeNil())
					Expect(job["appId"]).To(Equal(existingApp.ID.String()))
					Expect(job["templateName"]).To(Equal(existingTemplate.Name))
					Expect(job["totalBatches"]).To(Equal(float64(testJobs[idx].TotalBatches)))
					Expect(job["completedBatches"]).To(Equal(float64(testJobs[idx].CompletedBatches)))
					Expect(job["expiresAt"]).To(Equal(float64(testJobs[idx].ExpiresAt)))
					Expect(job["startsAt"]).To(Equal(float64(testJobs[idx].StartsAt)))
					Expect(job["csvPath"]).To(Equal(testJobs[idx].CSVPath))
					Expect(job["service"]).To(Equal(testJobs[idx].Service))
					Expect(job["createdBy"]).To(Equal(testJobs[idx].CreatedBy))
					Expect(job["createdAt"]).ToNot(BeNil())
					Expect(job["createdAt"]).ToNot(Equal(0))
					Expect(job["updatedAt"]).ToNot(BeNil())
					Expect(job["updatedAt"]).ToNot(Equal(0))

					tempFilters := job["filters"].(map[string]interface{})
					existFilters := testJobs[idx].Filters
					for key := range existFilters {
						Expect(tempFilters[key]).To(Equal(existFilters[key]))
					}

					tempContext := job["context"].(map[string]interface{})
					existContext := testJobs[idx].Context
					for key := range existContext {
						Expect(tempContext[key]).To(Equal(existContext[key]))
					}

					tempMetadata := job["metadata"].(map[string]interface{})
					existMetadata := testJobs[idx].Metadata
					for key := range existMetadata {
						Expect(tempMetadata[key]).To(Equal(existMetadata[key]))
					}
				}
			})

			It("should return 200 and a list of jobs without template", func() {
				CreateTestJobs(app.DB, existingApp.ID, existingTemplate.Name, 10)
				anotherTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				CreateTestJobs(app.DB, existingApp.ID, anotherTemplate.Name, 10)
				status, body := Get(app, baseRouteWithoutTemplate, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(20))

				for _, job := range response {
					Expect(job["id"]).ToNot(BeNil())
					Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Get(app, baseRoute, "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				status, body := Get(app, fmt.Sprintf("/apps/not-uuid/jobs?template=%s", existingTemplate.Name), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})

	Describe("Post /apps/:id/jobs?template=:templateName", func() {
		Describe("Sucesfully", func() {
			It("should return 201 and the created job with filters", func() {
				payload := GetJobPayload()
				delete(payload, "csvPath")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["totalBatches"]).To(BeEquivalentTo(0))
				Expect(job["completedBatches"]).To(BeEquivalentTo(0))
				Expect(job["expiresAt"]).To(BeNumerically("==", payload["expiresAt"]))
				Expect(job["startsAt"]).To(BeNumerically("==", payload["startsAt"]))
				Expect(job["csvPath"]).To(Equal(""))
				Expect(job["service"]).To(Equal(payload["service"]))
				Expect(job["createdBy"]).To(Equal("success@test.com"))
				Expect(job["createdAt"]).ToNot(BeNil())
				Expect(job["createdAt"]).ToNot(Equal(0))
				Expect(job["updatedAt"]).ToNot(BeNil())
				Expect(job["updatedAt"]).ToNot(Equal(0))

				tempFilters := job["filters"].(map[string]interface{})
				plFilters := payload["filters"].(map[string]interface{})
				for key := range plFilters {
					Expect(tempFilters[key]).To(Equal(plFilters[key]))
				}

				tempContext := job["context"].(map[string]interface{})
				plContext := payload["context"].(map[string]interface{})
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}

				tempMetadata := job["metadata"].(map[string]interface{})
				plMetadata := payload["metadata"].(map[string]interface{})
				for key := range plMetadata {
					Expect(tempMetadata[key]).To(Equal(plMetadata[key]))
				}

				id, err := uuid.FromString(job["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbJob := &model.Job{
					ID: id,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateName).To(Equal(existingTemplate.Name))
				Expect(dbJob.TotalBatches).To(Equal(0))
				Expect(dbJob.CompletedBatches).To(Equal(0))
				Expect(dbJob.ExpiresAt).To(BeEquivalentTo(payload["expiresAt"]))
				Expect(dbJob.StartsAt).To(BeEquivalentTo(payload["startsAt"]))
				Expect(dbJob.CSVPath).To(Equal(""))
				Expect(dbJob.Service).To(Equal(payload["service"]))
				Expect(dbJob.CreatedBy).To(Equal("success@test.com"))
				Expect(dbJob.CreatedAt).ToNot(BeNil())
				Expect(dbJob.UpdatedAt).ToNot(BeNil())

				for key := range plFilters {
					Expect(dbJob.Filters[key]).To(Equal(plFilters[key]))
				}

				for key := range plContext {
					Expect(dbJob.Context[key]).To(Equal(plContext[key]))
				}

				for key := range plMetadata {
					Expect(dbJob.Metadata[key]).To(Equal(plMetadata[key]))
				}
			})

			It("should return 201 and the created job with csvPath", func() {
				payload := GetJobPayload()
				delete(payload, "filters")
				payload["csvPath"] = "s3.aws.com/my-link"
				payload["service"] = "gcm"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["totalBatches"]).To(BeEquivalentTo(0))
				Expect(job["completedBatches"]).To(BeEquivalentTo(0))
				Expect(job["expiresAt"]).To(BeNumerically("==", payload["expiresAt"]))
				Expect(job["startsAt"]).To(BeNumerically("==", payload["startsAt"]))
				Expect(job["csvPath"]).To(Equal(payload["csvPath"]))
				Expect(job["filters"]).To(Equal(map[string]interface{}{}))
				Expect(job["service"]).To(Equal(payload["service"]))
				Expect(job["createdBy"]).To(Equal("success@test.com"))
				Expect(job["createdAt"]).ToNot(BeNil())
				Expect(job["createdAt"]).ToNot(Equal(0))
				Expect(job["updatedAt"]).ToNot(BeNil())
				Expect(job["updatedAt"]).ToNot(Equal(0))

				tempContext := job["context"].(map[string]interface{})
				plContext := payload["context"].(map[string]interface{})
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}

				tempMetadata := job["metadata"].(map[string]interface{})
				plMetadata := payload["metadata"].(map[string]interface{})
				for key := range plMetadata {
					Expect(tempMetadata[key]).To(Equal(plMetadata[key]))
				}

				id, err := uuid.FromString(job["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbJob := &model.Job{
					ID: id,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateName).To(Equal(existingTemplate.Name))
				Expect(dbJob.TotalBatches).To(Equal(0))
				Expect(dbJob.CompletedBatches).To(Equal(0))
				Expect(dbJob.ExpiresAt).To(Equal(payload["expiresAt"]))
				Expect(dbJob.StartsAt).To(Equal(payload["startsAt"]))
				Expect(dbJob.CSVPath).To(Equal(payload["csvPath"]))
				Expect(dbJob.Filters).To(Equal(map[string]interface{}{}))
				Expect(dbJob.Service).To(Equal(payload["service"]))
				Expect(dbJob.CreatedBy).To(Equal("success@test.com"))
				Expect(dbJob.CreatedAt).ToNot(BeNil())
				Expect(dbJob.CreatedAt).ToNot(Equal(0))
				Expect(dbJob.UpdatedAt).ToNot(BeNil())
				Expect(dbJob.UpdatedAt).ToNot(Equal(0))

				for key := range plContext {
					Expect(dbJob.Context[key]).To(Equal(plContext[key]))
				}

				for key := range plMetadata {
					Expect(dbJob.Metadata[key]).To(Equal(plMetadata[key]))
				}
			})

			It("should return 201 and the created job without expiresAt", func() {
				payload := GetJobPayload()
				delete(payload, "expiresAt")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["expiresAt"]).To(BeEquivalentTo(0))

				id, err := uuid.FromString(job["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbJob := &model.Job{
					ID: id,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateName).To(Equal(existingTemplate.Name))
				Expect(int(dbJob.ExpiresAt)).To(BeEquivalentTo(0))
			})

			It("should return 201 and the created job without startsAt", func() {
				payload := GetJobPayload()
				delete(payload, "startsAt")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["startsAt"]).To(BeEquivalentTo(0))

				id, err := uuid.FromString(job["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbJob := &model.Job{
					ID: id,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateName).To(Equal(existingTemplate.Name))
				Expect(int(dbJob.StartsAt)).To(BeEquivalentTo(0))
			})

			It("should return 201 and the created job without metadata", func() {
				payload := GetJobPayload()
				delete(payload, "metadata")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["metadata"]).To(Equal(map[string]interface{}{}))

				id, err := uuid.FromString(job["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbJob := &model.Job{
					ID: id,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateName).To(Equal(existingTemplate.Name))
				Expect(dbJob.Metadata).To(Equal(map[string]interface{}{}))
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, baseRoute, "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, baseRoute, string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/not-uuid/jobs?template=%s", existingTemplate.Name), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if app with given id does not exist", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/jobs?template=%s", uuid.NewV4().String(), existingTemplate.Name), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("no rows in result set"))
			})

			It("should return 422 if template with given name does not exist", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/jobs?template=%s", existingApp.ID, uuid.NewV4().String()), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("no rows in result set"))
			})

			It("should return 422 if template is not specified", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/jobs", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("template name must be specified"))
			})

			It("should return 422 if both csvPath and filters are provided", func() {
				payload := GetJobPayload()
				payload["csvPath"] = "s3.aws.com/my-link"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid filters or csvPath must exist, not both"))
			})

			It("should return 422 if missing context", func() {
				payload := GetJobPayload()
				delete(payload, "context")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid context"))
			})

			It("should return 422 if missing service", func() {
				payload := GetJobPayload()
				delete(payload, "service")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid service"))
			})

			It("should return 422 if invalid auth header", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "not-a-valid-email")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid createdBy"))
			})

			It("should return 422 if invalid context", func() {
				payload := GetJobPayload()
				payload["context"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go value"))
			})

			It("should return 422 if invalid service", func() {
				payload := GetJobPayload()
				payload["service"] = "blabla"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid service"))
			})

			It("should return 422 if invalid filters", func() {
				payload := GetJobPayload()
				payload["filters"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go value"))
			})

			It("should return 422 if invalid expiresAt", func() {
				payload := GetJobPayload()
				payload["expiresAt"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go value"))
			})

			It("should return 422 if invalid startsAt", func() {
				payload := GetJobPayload()
				payload["startsAt"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go value"))
			})

			It("should return 422 if past startsAt", func() {
				payload := GetJobPayload()
				payload["startsAt"] = time.Now().Add(-1 * time.Hour).UnixNano()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid startsAt"))
			})
		})
	})

	Describe("Get /apps/:id/jobs/:jid", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and the requested job", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.Name)
				status, body := Get(app, fmt.Sprintf("%s/%s", baseRouteWithoutTemplate, existingJob.ID), "success@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var job map[string]interface{}
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())
				Expect(job["id"]).ToNot(BeNil())
				Expect(job["appId"]).To(Equal(existingApp.ID.String()))
				Expect(job["templateName"]).To(Equal(existingTemplate.Name))
				Expect(job["totalBatches"]).To(Equal(float64(existingJob.TotalBatches)))
				Expect(job["completedBatches"]).To(Equal(float64(existingJob.CompletedBatches)))
				Expect(job["expiresAt"]).To(Equal(float64(existingJob.ExpiresAt)))
				Expect(job["startsAt"]).To(Equal(float64(existingJob.StartsAt)))
				Expect(job["csvPath"]).To(Equal(existingJob.CSVPath))
				Expect(job["service"]).To(Equal(existingJob.Service))
				Expect(job["createdBy"]).To(Equal(existingJob.CreatedBy))
				Expect(job["createdAt"]).To(Equal(float64(existingJob.CreatedAt)))
				Expect(job["updatedAt"]).To(Equal(float64(existingJob.UpdatedAt)))

				tempFilters := job["filters"].(map[string]interface{})
				plFilters := existingJob.Filters
				for key := range plFilters {
					Expect(tempFilters[key]).To(Equal(plFilters[key]))
				}

				tempContext := job["context"].(map[string]interface{})
				plContext := existingJob.Context
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}

				tempMetadata := job["metadata"].(map[string]interface{})
				plMetadata := existingJob.Metadata
				for key := range plMetadata {
					Expect(tempMetadata[key]).To(Equal(plMetadata[key]))
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.Name)
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRouteWithoutTemplate, existingJob.ID), "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.Name)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRouteWithoutTemplate, existingJob.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the job does not exist", func() {
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRouteWithoutTemplate, uuid.NewV4().String()), "test@test.com")
				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.Name)
				status, body := Get(app, fmt.Sprintf("/apps/not-uuid/jobs/%s", existingJob.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if job id is not UUID", func() {
				status, body := Get(app, fmt.Sprintf("%s/not-uuid", baseRouteWithoutTemplate), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})
})
