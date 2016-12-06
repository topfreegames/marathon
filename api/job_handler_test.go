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

var _ = Describe("App Handler", func() {
	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	app := GetDefaultTestApp(logger)
	faultyDb := GetFaultyTestDB(app)
	var existingApp *model.App
	var existingTemplate *model.Template
	var baseRoute string
	BeforeEach(func() {
		var dbApp model.App
		app.DB.Delete(&dbApp)
		var dbTemplate model.Template
		app.DB.Delete(&dbTemplate)
		existingApp = CreateTestApp(app.DB)
		existingTemplate = CreateTestTemplate(app.DB, existingApp.ID)
		baseRoute = fmt.Sprintf("/apps/%s/templates/%s/jobs", existingApp.ID, existingTemplate.ID)
	})

	Describe("Get /apps/:id/templates/:tid/jobs", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list if there are no jobs", func() {
				status, body := Get(app, baseRoute, "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				response := []model.Job{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of jobs", func() {
				testJobs := CreateTestJobs(app.DB, existingApp.ID, existingTemplate.ID, 10)
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				response := []model.Job{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, job := range response {
					Expect(job.ID).ToNot(BeNil())
					Expect(job.AppID).To(Equal(existingApp.ID))
					Expect(job.TemplateID).To(Equal(existingTemplate.ID))
					Expect(job.TotalBatches).To(Equal(testJobs[idx].TotalBatches))
					Expect(job.CompletedBatches).To(Equal(testJobs[idx].CompletedBatches))
					Expect(job.ExpiresAt.Unix()).To(Equal(testJobs[idx].ExpiresAt.Unix()))
					Expect(job.CsvURL).To(Equal(testJobs[idx].CsvURL))
					Expect(job.Service).To(Equal(testJobs[idx].Service))
					Expect(job.CreatedBy).To(Equal(testJobs[idx].CreatedBy))
					Expect(job.CreatedAt).ToNot(BeNil())
					Expect(job.UpdatedAt).ToNot(BeNil())

					var tempFilters map[string]interface{}
					err = json.Unmarshal([]byte(job.Filters), &tempFilters)
					Expect(err).NotTo(HaveOccurred())
					var existFilters map[string]interface{}
					err = json.Unmarshal([]byte(testJobs[idx].Filters), &existFilters)
					Expect(err).NotTo(HaveOccurred())
					for key := range existFilters {
						Expect(tempFilters[key]).To(Equal(existFilters[key]))
					}

					var tempContext map[string]interface{}
					err = json.Unmarshal([]byte(job.Context), &tempContext)
					Expect(err).NotTo(HaveOccurred())
					var existContext map[string]interface{}
					err = json.Unmarshal([]byte(testJobs[idx].Context), &existContext)
					Expect(err).NotTo(HaveOccurred())
					for key := range existContext {
						Expect(tempContext[key]).To(Equal(existContext[key]))
					}
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
				status, body := Get(app, fmt.Sprintf("/apps/not-uuid/templates/%s", existingTemplate.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if template id is not UUID", func() {
				status, body := Get(app, fmt.Sprintf("/apps/%s/templates/not-uuid", existingApp.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})

	Describe("Post /apps/:id/templates/:tid/jobs", func() {
		Describe("Sucesfully", func() {
			It("should return 201 and the created template with filters", func() {
				payload := GetJobPayload()
				delete(payload, "csvUrl")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job model.Job
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job.ID).ToNot(BeNil())
				Expect(job.AppID).To(Equal(existingApp.ID))
				Expect(job.TemplateID).To(Equal(existingTemplate.ID))
				Expect(job.TotalBatches).To(Equal(0))
				Expect(job.CompletedBatches).To(Equal(0))
				Expect(job.ExpiresAt).To(Equal(payload["expiresAt"]))
				Expect(job.CsvURL).To(Equal(""))
				Expect(job.Service).To(Equal(payload["service"]))
				Expect(job.CreatedBy).To(Equal("success@test.com"))
				Expect(job.CreatedAt).ToNot(BeNil())
				Expect(job.UpdatedAt).ToNot(BeNil())

				var tempFilters map[string]interface{}
				err = json.Unmarshal([]byte(job.Filters), &tempFilters)
				Expect(err).NotTo(HaveOccurred())
				var plFilters map[string]interface{}
				err = json.Unmarshal([]byte(payload["filters"].(string)), &plFilters)
				Expect(err).NotTo(HaveOccurred())
				for key := range plFilters {
					Expect(tempFilters[key]).To(Equal(plFilters[key]))
				}

				var tempContext map[string]interface{}
				err = json.Unmarshal([]byte(job.Context), &tempContext)
				Expect(err).NotTo(HaveOccurred())
				var plContext map[string]interface{}
				err = json.Unmarshal([]byte(payload["context"].(string)), &plContext)
				Expect(err).NotTo(HaveOccurred())
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}

				dbJob := &model.Job{
					ID: job.ID,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateID).To(Equal(existingTemplate.ID))
				Expect(dbJob.TotalBatches).To(Equal(0))
				Expect(dbJob.CompletedBatches).To(Equal(0))
				Expect(dbJob.ExpiresAt.Unix()).To(Equal(payload["expiresAt"].(time.Time).Unix()))
				Expect(dbJob.CsvURL).To(Equal(""))
				Expect(dbJob.Service).To(Equal(payload["service"]))
				Expect(dbJob.CreatedBy).To(Equal("success@test.com"))
				Expect(dbJob.CreatedAt).ToNot(BeNil())
				Expect(dbJob.UpdatedAt).ToNot(BeNil())

				err = json.Unmarshal([]byte(dbJob.Filters), &tempFilters)
				Expect(err).NotTo(HaveOccurred())
				for key := range plFilters {
					Expect(tempFilters[key]).To(Equal(plFilters[key]))
				}

				err = json.Unmarshal([]byte(dbJob.Context), &tempContext)
				Expect(err).NotTo(HaveOccurred())
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}
			})

			It("should return 201 and the created template with csvUrl", func() {
				payload := GetJobPayload()
				delete(payload, "filters")
				payload["csvUrl"] = "s3.aws.com/my-link"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job model.Job
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job.ID).ToNot(BeNil())
				Expect(job.AppID).To(Equal(existingApp.ID))
				Expect(job.TemplateID).To(Equal(existingTemplate.ID))
				Expect(job.TotalBatches).To(Equal(0))
				Expect(job.CompletedBatches).To(Equal(0))
				Expect(job.ExpiresAt).To(Equal(payload["expiresAt"]))
				Expect(job.CsvURL).To(Equal(payload["csvUrl"]))
				Expect(job.Filters).To(Equal("{}"))
				Expect(job.Service).To(Equal(payload["service"]))
				Expect(job.CreatedBy).To(Equal("success@test.com"))
				Expect(job.CreatedAt).ToNot(BeNil())
				Expect(job.UpdatedAt).ToNot(BeNil())

				var tempContext map[string]interface{}
				err = json.Unmarshal([]byte(job.Context), &tempContext)
				Expect(err).NotTo(HaveOccurred())
				var plContext map[string]interface{}
				err = json.Unmarshal([]byte(payload["context"].(string)), &plContext)
				Expect(err).NotTo(HaveOccurred())
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}

				dbJob := &model.Job{
					ID: job.ID,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateID).To(Equal(existingTemplate.ID))
				Expect(job.TotalBatches).To(Equal(0))
				Expect(dbJob.CompletedBatches).To(Equal(0))
				Expect(dbJob.ExpiresAt.Unix()).To(Equal(payload["expiresAt"].(time.Time).Unix()))
				Expect(dbJob.CsvURL).To(Equal(payload["csvUrl"]))
				Expect(dbJob.Filters).To(Equal("{}"))
				Expect(dbJob.Service).To(Equal(payload["service"]))
				Expect(dbJob.CreatedBy).To(Equal("success@test.com"))
				Expect(dbJob.CreatedAt).ToNot(BeNil())
				Expect(dbJob.UpdatedAt).ToNot(BeNil())

				err = json.Unmarshal([]byte(dbJob.Context), &tempContext)
				Expect(err).NotTo(HaveOccurred())
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}
			})

			It("should return 201 and the created template without expiresAt", func() {
				payload := GetJobPayload()
				delete(payload, "expiresAt")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var job model.Job
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())

				Expect(job.ID).ToNot(BeNil())
				Expect(job.AppID).To(Equal(existingApp.ID))
				Expect(job.TemplateID).To(Equal(existingTemplate.ID))
				Expect(int(job.ExpiresAt.Unix())).To(Equal(-62135596800))

				dbJob := &model.Job{
					ID: job.ID,
				}
				err = app.DB.Select(&dbJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbJob.ID).ToNot(BeNil())
				Expect(dbJob.AppID).To(Equal(existingApp.ID))
				Expect(dbJob.TemplateID).To(Equal(existingTemplate.ID))
				Expect(int(dbJob.ExpiresAt.Unix())).To(Equal(-62135596800))
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
				status, body := Post(app, fmt.Sprintf("/apps/not-uuid/templates/%s/jobs", existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if template id is not UUID", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/templates/not-uuid/jobs", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if app with given id does not exist", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/templates/%s/jobs", uuid.NewV4().String(), existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("pq: insert or update on table \"jobs\" violates foreign key constraint \"jobs_app_id_apps_id_foreign\""))
			})

			It("should return 422 if template with given id does not exist", func() {
				payload := GetJobPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/templates/%s/jobs", existingApp.ID, uuid.NewV4().String()), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("pq: insert or update on table \"jobs\" violates foreign key constraint \"jobs_template_id_templates_id_foreign\""))
			})

			It("should return 422 if missing csvUrl and filters", func() {
				payload := GetJobPayload()
				delete(payload, "csvUrl")
				delete(payload, "filters")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid filters or csvUrl must exist"))
			})

			It("should return 422 if both csvUrl and filters are provided", func() {
				payload := GetJobPayload()
				payload["csvUrl"] = "s3.aws.com/my-link"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid filters or csvUrl must exist, not both"))
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
				Expect(response["reason"]).To(Equal("invalid context"))
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
				Expect(response["reason"]).To(Equal("invalid filters"))
			})

			It("should return 422 if invalid csvUrl", func() {
				payload := GetJobPayload()
				payload["csvUrl"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid csvUrl"))
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
				Expect(response["reason"]).To(ContainSubstring("parsing time"))
			})
		})
	})

	Describe("Get /apps/:id/templates/:tid/jobs/:jid", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and the requested template", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.ID)
				status, body := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingJob.ID), "success@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var job model.Job
				err := json.Unmarshal([]byte(body), &job)
				Expect(err).NotTo(HaveOccurred())
				Expect(job.ID).ToNot(BeNil())
				Expect(job.AppID).To(Equal(existingApp.ID))
				Expect(job.TemplateID).To(Equal(existingTemplate.ID))
				Expect(job.TotalBatches).To(Equal(existingJob.TotalBatches))
				Expect(job.CompletedBatches).To(Equal(existingJob.CompletedBatches))
				Expect(job.ExpiresAt.Unix()).To(Equal(existingJob.ExpiresAt.Unix()))
				Expect(job.CsvURL).To(Equal(existingJob.CsvURL))
				Expect(job.Service).To(Equal(existingJob.Service))
				Expect(job.CreatedBy).To(Equal(existingJob.CreatedBy))
				Expect(job.CreatedAt.Unix()).To(Equal(existingJob.CreatedAt.Unix()))
				Expect(job.UpdatedAt.Unix()).To(Equal(existingJob.UpdatedAt.Unix()))

				var tempFilters map[string]interface{}
				err = json.Unmarshal([]byte(job.Filters), &tempFilters)
				Expect(err).NotTo(HaveOccurred())
				var plFilters map[string]interface{}
				err = json.Unmarshal([]byte(existingJob.Filters), &plFilters)
				Expect(err).NotTo(HaveOccurred())
				for key := range plFilters {
					Expect(tempFilters[key]).To(Equal(plFilters[key]))
				}

				var tempContext map[string]interface{}
				err = json.Unmarshal([]byte(job.Context), &tempContext)
				Expect(err).NotTo(HaveOccurred())
				var plContext map[string]interface{}
				err = json.Unmarshal([]byte(existingJob.Context), &plContext)
				Expect(err).NotTo(HaveOccurred())
				for key := range plContext {
					Expect(tempContext[key]).To(Equal(plContext[key]))
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.ID)
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingJob.ID), "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.ID)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingJob.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the job does not exist", func() {
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, uuid.NewV4().String()), "test@test.com")
				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.ID)
				status, body := Get(app, fmt.Sprintf("/apps/not-uuid/templates/%s/jobs/%s", existingTemplate.ID, existingJob.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if template id is not UUID", func() {
				existingJob := CreateTestJob(app.DB, existingApp.ID, existingTemplate.ID)
				status, body := Get(app, fmt.Sprintf("/apps/%s/templates/not-uuid/jobs/%s", existingApp.ID, existingJob.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 422 if job id is not UUID", func() {
				status, body := Get(app, fmt.Sprintf("%s/not-uuid", baseRoute), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})
})
