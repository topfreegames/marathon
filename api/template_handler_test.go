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
	"strings"

	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/api"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("App Handler", func() {
	var logger zap.Logger
	var faultyDb *gorm.DB
	var app *api.Application
	var existingApp *model.App
	var baseRoute string
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		app = GetDefaultTestApp(logger)
		faultyDb = GetFaultyTestDB(app)

		var dbApp model.App
		app.DB.Delete(&dbApp)
		var dbTemplate model.Template
		app.DB.Delete(&dbTemplate)
		existingApp = CreateTestApp(app.DB)
		baseRoute = fmt.Sprintf("/apps/%s/templates", existingApp.ID)
	})

	Describe("Get /apps/:id/templates", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list of templates if there are no templates", func() {
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				response := []model.Template{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of templates", func() {
				testTemplates := CreateTestTemplates(app.DB, existingApp.ID, 10)
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				response := []model.Template{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, template := range response {
					Expect(template.ID).ToNot(BeNil())
					Expect(template.AppID).To(Equal(existingApp.ID))
					Expect(template.Name).To(Equal(testTemplates[idx].Name))
					Expect(template.Locale).To(Equal(testTemplates[idx].Locale))
					Expect(template.CompiledBody).To(Equal(testTemplates[idx].CompiledBody))
					Expect(template.CreatedBy).To(Equal(testTemplates[idx].CreatedBy))
					Expect(template.CreatedAt).ToNot(BeNil())
					Expect(template.UpdatedAt).ToNot(BeNil())

					var tempBody map[string]interface{}
					err = json.Unmarshal([]byte(template.Body), &tempBody)
					Expect(err).NotTo(HaveOccurred())
					var existBody map[string]interface{}
					err = json.Unmarshal([]byte(testTemplates[idx].Body), &existBody)
					Expect(err).NotTo(HaveOccurred())
					for key, _ := range existBody {
						Expect(tempBody[key]).To(Equal(existBody[key]))
					}

					var tempDefaults map[string]interface{}
					err = json.Unmarshal([]byte(template.Defaults), &tempDefaults)
					Expect(err).NotTo(HaveOccurred())
					var existDefaults map[string]interface{}
					err = json.Unmarshal([]byte(testTemplates[idx].Defaults), &existDefaults)
					Expect(err).NotTo(HaveOccurred())
					for key, _ := range existDefaults {
						Expect(tempDefaults[key]).To(Equal(existDefaults[key]))
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
				app.DB = faultyDb
				status, _ := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
			})

			It("should return 422 if app id is not UUID", func() {
				status, body := Get(app, "/apps/not-uuid/templates", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})
		})
	})

	Describe("Post /apps/:id/templates", func() {
		Describe("Sucesfully", func() {
			It("should return 201 and the created template", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var template model.Template
				err := json.Unmarshal([]byte(body), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template.ID).ToNot(BeNil())
				Expect(template.AppID).To(Equal(existingApp.ID))
				Expect(template.Name).To(Equal(payload["name"]))
				Expect(template.Locale).To(Equal(payload["locale"]))
				// TODO: will this exist? when will it be created?
				// Expect(template.CompiledBody).To(Equal(Equal(payload["compiledBody"])))
				Expect(template.CreatedBy).To(Equal("success@test.com"))
				Expect(template.CreatedAt).ToNot(BeNil())
				Expect(template.UpdatedAt).ToNot(BeNil())

				var tempBody map[string]interface{}
				err = json.Unmarshal([]byte(template.Body), &tempBody)
				Expect(err).NotTo(HaveOccurred())
				var plBody map[string]interface{}
				err = json.Unmarshal([]byte(payload["body"].(string)), &tempBody)
				Expect(err).NotTo(HaveOccurred())
				for key, _ := range plBody {
					Expect(tempBody[key]).To(Equal(plBody[key]))
				}

				var tempDefaults map[string]interface{}
				err = json.Unmarshal([]byte(template.Defaults), &tempDefaults)
				Expect(err).NotTo(HaveOccurred())
				var plDefaults map[string]interface{}
				err = json.Unmarshal([]byte(payload["defaults"].(string)), &tempDefaults)
				Expect(err).NotTo(HaveOccurred())
				for key, _ := range plDefaults {
					Expect(tempDefaults[key]).To(Equal(plDefaults[key]))
				}

				var dbTemplate model.Template
				err = app.DB.Where(&model.Template{ID: template.ID}).First(&dbTemplate).Error
				Expect(err).NotTo(HaveOccurred())
				Expect(dbTemplate.ID).ToNot(BeNil())
				Expect(dbTemplate.AppID).To(Equal(existingApp.ID))
				Expect(dbTemplate.Name).To(Equal(payload["name"]))
				Expect(dbTemplate.Locale).To(Equal(payload["locale"]))
				// TODO: will this exist? when will it be created?
				// Expect(dbTemplate.CompiledBody).To(Equal(Equal(payload["compiledBody"])))
				Expect(dbTemplate.CreatedBy).To(Equal("success@test.com"))
				Expect(dbTemplate.CreatedAt).ToNot(BeNil())
				Expect(dbTemplate.UpdatedAt).ToNot(BeNil())

				err = json.Unmarshal([]byte(dbTemplate.Body), &tempBody)
				Expect(err).NotTo(HaveOccurred())
				for key, _ := range plBody {
					Expect(tempBody[key]).To(Equal(plBody[key]))
				}

				err = json.Unmarshal([]byte(dbTemplate.Defaults), &tempDefaults)
				Expect(err).NotTo(HaveOccurred())
				for key, _ := range plDefaults {
					Expect(tempDefaults[key]).To(Equal(plDefaults[key]))
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, baseRoute, "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				app.DB = faultyDb
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, baseRoute, string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
			})

			It("should return 422 if app id is not UUID", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps/not-uuid/templates", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: UUID string too short"))
			})

			It("should return 409 if template with same appId, name and locale already exists", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload(map[string]interface{}{
					"name":   existingTemplate.Name,
					"locale": existingTemplate.Locale,
				})
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("pq: duplicate key value violates unique constraint \"name_locale_app\""))
			})

			It("should return 422 if app with given id does not exist", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/templates", uuid.NewV4().String()), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("pq: insert or update on table \"templates\" violates foreign key constraint \"templates_app_id_apps_id_foreign\""))
			})

			It("should return 422 if missing name", func() {
				payload := GetTemplatePayload()
				delete(payload, "name")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if missing locale", func() {
				payload := GetTemplatePayload()
				delete(payload, "locale")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid locale"))
			})

			It("should return 422 if missing defaults", func() {
				payload := GetTemplatePayload()
				delete(payload, "defaults")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid defaults"))
			})

			It("should return 422 if missing body", func() {
				payload := GetTemplatePayload()
				delete(payload, "body")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid body"))
			})

			It("should return 422 if invalid auth header", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "not-a-valid-email")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid createdBy"))
			})

			It("should return 422 if invalid name", func() {
				payload := GetTemplatePayload()
				payload["name"] = strings.Repeat("a", 256)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if invalid locale", func() {
				payload := GetTemplatePayload()
				payload["locale"] = strings.Repeat("a", 11)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid locale"))
			})

			It("should return 422 if invalid defaults", func() {
				payload := GetTemplatePayload()
				payload["defaults"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid defaults"))
			})

			It("should return 422 if invalid body", func() {
				payload := GetTemplatePayload()
				payload["body"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid body"))
			})
		})
	})
})
