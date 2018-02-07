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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/api"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("Template Handler", func() {
	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	app := GetDefaultTestApp(logger)
	faultyDb := GetFaultyTestDB(app)
	var existingApp *model.App
	var baseRoute string
	var multipleRoute string
	BeforeEach(func() {
		app.DB.Exec("DELETE FROM apps;")
		app.DB.Delete("DELETE FROM templates;")
		app.DB.Exec("DELETE FROM users;")
		CreateTestUser(app.DB, map[string]interface{}{"email": "test@test.com", "isAdmin": true})
		CreateTestUser(app.DB, map[string]interface{}{"email": "success@test.com", "isAdmin": true})
		existingApp = CreateTestApp(app.DB)
		baseRoute = fmt.Sprintf("/apps/%s/templates", existingApp.ID)
		multipleRoute = fmt.Sprintf("%s?multiple=true", baseRoute)
	})

	Describe("Get /apps/:id/templates", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list of templates if there are no templates", func() {
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of templates", func() {
				testTemplates := CreateTestTemplates(app.DB, existingApp.ID, 10)
				status, body := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, template := range response {
					Expect(template["id"]).ToNot(BeNil())
					Expect(template["appId"]).To(Equal(existingApp.ID.String()))
					Expect(template["name"]).To(Equal(testTemplates[idx].Name))
					Expect(template["locale"]).To(Equal(testTemplates[idx].Locale))
					Expect(template["createdBy"]).To(Equal(testTemplates[idx].CreatedBy))
					Expect(template["createdAt"]).ToNot(BeNil())
					Expect(template["createdAt"]).ToNot(Equal(0))
					Expect(template["updatedAt"]).ToNot(BeNil())
					Expect(template["updatedAt"]).ToNot(Equal(0))

					tempBody := template["body"].(map[string]interface{})
					existBody := testTemplates[idx].Body
					for key := range existBody {
						Expect(tempBody[key]).To(Equal(existBody[key]))
					}

					tempDefaults := template["defaults"].(map[string]interface{})
					existDefaults := testTemplates[idx].Defaults
					for key := range existDefaults {
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
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, baseRoute, "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				status, body := Get(app, "/apps/not-uuid/templates", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})
		})
	})

	Describe("Post /apps/:id/templates", func() {
		Describe("Successfully", func() {
			It("should return 201 and the created template", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, baseRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var template map[string]interface{}
				err := json.Unmarshal([]byte(body), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template["id"]).ToNot(BeNil())
				Expect(template["appId"]).To(Equal(existingApp.ID.String()))
				Expect(template["name"]).To(Equal(payload["name"]))
				Expect(template["locale"]).To(Equal(payload["locale"]))
				Expect(template["createdBy"]).To(Equal("success@test.com"))
				Expect(template["createdAt"]).ToNot(BeNil())
				Expect(template["createdAt"]).ToNot(Equal(0))
				Expect(template["updatedAt"]).ToNot(BeNil())
				Expect(template["updatedAt"]).ToNot(Equal(0))

				tempBody := template["body"].(map[string]interface{})
				plBody := payload["body"].(map[string]interface{})
				for key := range plBody {
					Expect(tempBody[key]).To(Equal(plBody[key]))
				}

				tempDefaults := template["defaults"].(map[string]interface{})
				plDefaults := payload["defaults"].(map[string]interface{})
				for key := range plDefaults {
					Expect(tempDefaults[key]).To(Equal(plDefaults[key]))
				}

				id, err := uuid.FromString(template["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbTemplate := &model.Template{
					ID: id,
				}
				err = app.DB.Select(&dbTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbTemplate.ID).ToNot(BeNil())
				Expect(dbTemplate.AppID).To(Equal(existingApp.ID))
				Expect(dbTemplate.Name).To(Equal(payload["name"]))
				Expect(dbTemplate.Locale).To(Equal(payload["locale"]))
				Expect(dbTemplate.CreatedBy).To(Equal("success@test.com"))
				Expect(dbTemplate.CreatedAt).ToNot(BeNil())
				Expect(dbTemplate.UpdatedAt).ToNot(BeNil())

				for key := range plBody {
					Expect(dbTemplate.Body[key]).To(Equal(plBody[key]))
				}

				for key := range plDefaults {
					Expect(dbTemplate.Defaults[key]).To(Equal(plDefaults[key]))
				}
			})

			It("should return 201 and the created templates when with flag multiple", func() {
				payload := GetTemplatePayloads(3)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, multipleRoute, string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var templates []map[string]interface{}
				err := json.Unmarshal([]byte(body), &templates)
				Expect(err).NotTo(HaveOccurred())

				Expect(templates[0]["id"]).ToNot(BeNil())
				Expect(templates[0]["appId"]).To(Equal(existingApp.ID.String()))
				Expect(templates[0]["name"]).To(Equal(payload[0]["name"]))
				Expect(templates[0]["locale"]).To(Equal(payload[0]["locale"]))
				Expect(templates[0]["createdBy"]).To(Equal("success@test.com"))
				Expect(templates[0]["createdAt"]).ToNot(BeNil())
				Expect(templates[0]["createdAt"]).ToNot(Equal(0))
				Expect(templates[0]["updatedAt"]).ToNot(BeNil())
				Expect(templates[0]["updatedAt"]).ToNot(Equal(0))

				Expect(templates[1]["id"]).ToNot(BeNil())
				Expect(templates[1]["appId"]).To(Equal(existingApp.ID.String()))
				Expect(templates[1]["name"]).To(Equal(payload[1]["name"]))
				Expect(templates[1]["locale"]).To(Equal(payload[1]["locale"]))
				Expect(templates[1]["createdBy"]).To(Equal("success@test.com"))
				Expect(templates[1]["createdAt"]).ToNot(BeNil())
				Expect(templates[1]["createdAt"]).ToNot(Equal(0))
				Expect(templates[1]["updatedAt"]).ToNot(BeNil())
				Expect(templates[1]["updatedAt"]).ToNot(Equal(0))

				Expect(templates[2]["id"]).ToNot(BeNil())
				Expect(templates[2]["appId"]).To(Equal(existingApp.ID.String()))
				Expect(templates[2]["name"]).To(Equal(payload[2]["name"]))
				Expect(templates[2]["locale"]).To(Equal(payload[2]["locale"]))
				Expect(templates[2]["createdBy"]).To(Equal("success@test.com"))
				Expect(templates[2]["createdAt"]).ToNot(BeNil())
				Expect(templates[2]["createdAt"]).ToNot(Equal(0))
				Expect(templates[2]["updatedAt"]).ToNot(BeNil())
				Expect(templates[2]["updatedAt"]).ToNot(Equal(0))

				id, err := uuid.FromString(templates[0]["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbTemplate := &model.Template{
					ID: id,
				}
				err = app.DB.Select(&dbTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbTemplate.ID).ToNot(BeNil())
				Expect(dbTemplate.AppID).To(Equal(existingApp.ID))
				Expect(dbTemplate.Name).To(Equal(payload[0]["name"]))
				Expect(dbTemplate.Locale).To(Equal(payload[0]["locale"]))
				Expect(dbTemplate.CreatedBy).To(Equal("success@test.com"))
				Expect(dbTemplate.CreatedAt).ToNot(BeNil())
				Expect(dbTemplate.UpdatedAt).ToNot(BeNil())
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, baseRoute, "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, baseRoute, string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps/not-uuid/templates", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
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
				Expect(response["reason"]).To(ContainSubstring("violates unique constraint \"name_locale_app\""))
			})

			It("should return 409 if template with same appId, name and locale already exists with multiple templates", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := []map[string]interface{}{GetTemplatePayload(map[string]interface{}{
					"name":   existingTemplate.Name,
					"locale": existingTemplate.Locale,
				})}
				payload = append(payload, GetTemplatePayloads(2)...)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, multipleRoute, string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("violates unique constraint \"name_locale_app\""))
			})

			It("should return 422 if app with given id does not exist", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, fmt.Sprintf("/apps/%s/templates", uuid.NewV4().String()), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("violates foreign key constraint \"templates_app_id_apps_id_foreign\""))
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

			It("should return 422 if invalid name with multiple templates", func() {
				payload := GetTemplatePayloads(3)
				payload[0]["name"] = strings.Repeat("a", 256)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, multipleRoute, string(pl), "test@test.com")
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
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go struct"))
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
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go struct"))
			})
		})
	})

	Describe("Get /apps/:id/templates/:tid", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and the requested template", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, body := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "success@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var template map[string]interface{}
				err := json.Unmarshal([]byte(body), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template["id"]).ToNot(BeNil())
				Expect(template["appId"]).To(Equal(existingApp.ID.String()))
				Expect(template["name"]).To(Equal(existingTemplate.Name))
				Expect(template["locale"]).To(Equal(existingTemplate.Locale))
				Expect(template["createdBy"]).To(Equal(existingTemplate.CreatedBy))
				Expect(template["createdAt"]).ToNot(BeNil())
				Expect(template["createdAt"]).ToNot(Equal(0))
				Expect(template["updatedAt"]).ToNot(BeNil())
				Expect(template["updatedAt"]).ToNot(Equal(0))

				tempBody := template["body"].(map[string]interface{})
				for key := range existingTemplate.Body {
					Expect(tempBody[key]).To(Equal(existingTemplate.Body[key]))
				}

				tempDefaults := template["defaults"].(map[string]interface{})
				for key := range existingTemplate.Defaults {
					Expect(tempDefaults[key]).To(Equal(existingTemplate.Defaults[key]))
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the template does not exist", func() {
				status, _ := Get(app, fmt.Sprintf("%s/%s", baseRoute, uuid.NewV4().String()), "test@test.com")
				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, body := Get(app, fmt.Sprintf("/apps/not-uuid/templates/%s", existingTemplate.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})

			It("should return 422 if template id is not UUID", func() {
				status, body := Get(app, fmt.Sprintf("%s/not-uuid", baseRoute), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})
		})
	})

	Describe("Put /apps/:id/templates/:tid", func() {
		Describe("Successfully", func() {
			It("should return 200 and the updated template", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var template map[string]interface{}
				err := json.Unmarshal([]byte(body), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template["id"]).ToNot(BeNil())
				Expect(template["appId"]).To(Equal(existingApp.ID.String()))
				Expect(template["name"]).To(Equal(payload["name"]))
				Expect(template["locale"]).To(Equal(payload["locale"]))
				Expect(template["createdBy"]).To(Equal(existingTemplate.CreatedBy))
				Expect(int64(template["createdAt"].(float64))).To(Equal(existingTemplate.CreatedAt))
				Expect(template["updatedAt"]).ToNot(Equal(existingTemplate.UpdatedAt))

				tempBody := template["body"].(map[string]interface{})
				plBody := payload["body"].(map[string]interface{})
				for key := range plBody {
					Expect(tempBody[key]).To(Equal(plBody[key]))
				}

				tempDefaults := template["defaults"].(map[string]interface{})
				plDefaults := payload["defaults"].(map[string]interface{})
				for key := range plDefaults {
					Expect(tempDefaults[key]).To(Equal(plDefaults[key]))
				}

				id, err := uuid.FromString(template["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbTemplate := &model.Template{
					ID: id,
				}
				err = app.DB.Select(&dbTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbTemplate.ID).ToNot(BeNil())
				Expect(dbTemplate.AppID).To(Equal(existingApp.ID))
				Expect(dbTemplate.Name).To(Equal(payload["name"]))
				Expect(dbTemplate.Locale).To(Equal(payload["locale"]))
				Expect(dbTemplate.CreatedBy).To(Equal(existingTemplate.CreatedBy))
				Expect(dbTemplate.CreatedAt).ToNot(BeNil())
				Expect(dbTemplate.UpdatedAt).ToNot(BeNil())

				for key := range plBody {
					Expect(dbTemplate.Body[key]).To(Equal(plBody[key]))
				}

				for key := range plDefaults {
					Expect(dbTemplate.Defaults[key]).To(Equal(plDefaults[key]))
				}
			})

			It("should return 200 if missing defaults", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				delete(payload, "defaults")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var template map[string]interface{}
				err := json.Unmarshal([]byte(body), &template)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, _ := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				goodDB := app.DB
				app.DB = faultyDb
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, _ := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/not-uuid/templates/%s", existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})

			It("should return 422 if template id is not UUID", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/not-uuid", baseRoute), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})

			It("should return 409 if template with same appId, name and locale already exists", func() {
				existingTemplate1 := CreateTestTemplate(app.DB, existingApp.ID)
				existingTemplate2 := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload(map[string]interface{}{
					"name":   existingTemplate1.Name,
					"locale": existingTemplate1.Locale,
				})
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate2.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("violates unique constraint \"name_locale_app\""))
			})

			It("should return 404 if template with given id does not exist", func() {
				payload := GetTemplatePayload()
				pl, _ := json.Marshal(payload)
				status, _ := Put(app, fmt.Sprintf("%s/%s", baseRoute, uuid.NewV4().String()), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if missing name", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				delete(payload, "name")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if missing locale", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				delete(payload, "locale")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid locale"))
			})

			It("should return 422 if missing body", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				delete(payload, "body")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid body"))
			})

			It("should return 422 if invalid name", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				payload["name"] = strings.Repeat("a", 256)
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if invalid locale", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				payload["locale"] = strings.Repeat("a", 11)
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid locale"))
			})

			It("should return 422 if invalid defaults", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				payload["defaults"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go struct"))
			})

			It("should return 422 if invalid body", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				payload := GetTemplatePayload()
				payload["body"] = "not-json"
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("cannot unmarshal string into Go struct"))
			})
		})
	})

	Describe("Delete /apps/:id/templates/:tid", func() {
		Describe("Sucesfully", func() {
			It("should return 204 ", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, _ := Delete(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusNoContent))

				dbTemplate := &model.Template{
					ID: existingTemplate.ID,
				}
				err := app.DB.Select(&dbTemplate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(api.RecordNotFoundString))
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Delete(app, "/apps/1234/templates/5678", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Delete(app, fmt.Sprintf("%s/%s", baseRoute, existingTemplate.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the app does not exist", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, _ := Delete(app, fmt.Sprintf("/apps/%s/templates/%s", uuid.NewV4().String(), existingTemplate.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 404 if the template does not exist", func() {
				status, _ := Delete(app, fmt.Sprintf("%s/%s", baseRoute, uuid.NewV4().String()), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				existingTemplate := CreateTestTemplate(app.DB, existingApp.ID)
				status, body := Delete(app, fmt.Sprintf("/apps/not-uuid/templates/%s", existingTemplate.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})

			It("should return 422 if template id is not UUID", func() {
				status, body := Delete(app, fmt.Sprintf("%s/not-uuid", baseRoute), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})
		})
	})
})
