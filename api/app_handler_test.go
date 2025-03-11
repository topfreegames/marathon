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

var _ = Describe("App Handler", func() {
	logger := zap.New(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		zap.FatalLevel,
	)
	app := GetDefaultTestApp(logger)
	faultyDb := GetFaultyTestDB(app)
	BeforeEach(func() {
		app.DB.Exec("DELETE FROM apps;")
		app.DB.Exec("DELETE FROM users;")
		CreateTestUser(app.DB, map[string]interface{}{"email": "test@test.com", "isAdmin": true})
	})

	Describe("Get /apps", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list of apps if there are no apps", func() {
				status, body := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of apps", func() {
				testApps := CreateTestApps(app.DB, 10)
				status, body := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				var response []map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, app := range response {
					Expect(app["id"]).ToNot(BeNil())
					Expect(app["name"]).To(Equal(testApps[idx].Name))
					Expect(app["bundleId"]).To(Equal(testApps[idx].BundleID))
					Expect(app["createdBy"]).To(Equal(testApps[idx].CreatedBy))
					Expect(app["createdAt"]).ToNot(BeNil())
					Expect(app["createdAt"]).ToNot(Equal(0))
					Expect(app["updatedAt"]).ToNot(BeNil())
					Expect(app["updatedAt"]).ToNot(Equal(0))
				}
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Get(app, "/apps", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})
		})
	})

	Describe("Post /apps", func() {
		Describe("Sucesfully", func() {
			It("should return 201 and the created app", func() {
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["id"]).ToNot(BeNil())
				Expect(response["name"]).To(Equal(payload["name"]))
				Expect(response["bundleId"]).To(Equal(payload["bundleId"]))
				Expect(response["createdBy"]).To(Equal("test@test.com"))
				Expect(response["createdAt"]).ToNot(BeNil())
				Expect(response["createdAt"]).ToNot(Equal(0))
				Expect(response["updatedAt"]).ToNot(BeNil())
				Expect(response["updatedAt"]).ToNot(Equal(0))

				id, err := uuid.FromString(response["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbApp := &model.App{
					ID: id,
				}
				err = app.DB.Model(dbApp).Select()
				Expect(err).NotTo(HaveOccurred())
				Expect(dbApp).NotTo(BeNil())
				Expect(dbApp.Name).To(Equal(payload["name"]))
				Expect(dbApp.BundleID).To(Equal(payload["bundleId"]))
				Expect(dbApp.CreatedBy).To(Equal("test@test.com"))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, "/apps", "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				app.DB = faultyDb
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, "/apps", string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 409 if app with same bundleId already exists", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload(map[string]interface{}{"bundleId": existingApp.BundleID})
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uix_apps_bundle_id"))
			})

			It("should return 422 if missing name", func() {
				payload := GetAppPayload()
				delete(payload, "name")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if missing bundleId", func() {
				payload := GetAppPayload()
				delete(payload, "bundleId")
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid bundleId"))
			})

			It("should return 422 if invalid name", func() {
				payload := GetAppPayload()
				payload["name"] = strings.Repeat("a", 256)
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if invalid bundleId", func() {
				payload := GetAppPayload()
				payload["bundleId"] = "invalidformat"
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid bundleId"))
			})
		})
	})

	Describe("Get /apps/:id", func() {
		Describe("Successfully", func() {
			It("should return 200 and the requested app if admin", func() {
				existingApp := CreateTestApp(app.DB)
				status, body := Get(app, fmt.Sprintf("/apps/%s", existingApp.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["id"]).To(Equal(existingApp.ID.String()))
				Expect(response["name"]).To(Equal(existingApp.Name))
				Expect(response["bundleId"]).To(Equal(existingApp.BundleID))
				Expect(response["createdBy"]).To(Equal(existingApp.CreatedBy))
				Expect(response["createdAt"]).ToNot(BeNil())
				Expect(response["createdAt"]).ToNot(Equal(0))
				Expect(response["updatedAt"]).ToNot(BeNil())
				Expect(response["updatedAt"]).ToNot(Equal(0))
			})

			It("should return 200 and the requested app if not admin but allowed", func() {
				existingApp := CreateTestApp(app.DB)
				CreateTestUser(app.DB, map[string]interface{}{"email": "test2@test.com", "isAdmin": false, "allowedApps": []uuid.UUID{existingApp.ID}})
				status, body := Get(app, fmt.Sprintf("/apps/%s", existingApp.ID), "test2@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["id"]).To(Equal(existingApp.ID.String()))
				Expect(response["name"]).To(Equal(existingApp.Name))
				Expect(response["bundleId"]).To(Equal(existingApp.BundleID))
				Expect(response["createdBy"]).To(Equal(existingApp.CreatedBy))
				Expect(response["createdAt"]).ToNot(BeNil())
				Expect(response["createdAt"]).ToNot(Equal(0))
				Expect(response["updatedAt"]).ToNot(BeNil())
				Expect(response["updatedAt"]).ToNot(Equal(0))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				CreateTestUser(app.DB, map[string]interface{}{"email": "test2@test.com", "isAdmin": false})
				status, _ := Get(app, fmt.Sprintf("/apps/%s", uuid.NewV4()), "test2@test.com")

				Expect(status).To(Equal(http.StatusForbidden))
			})

			It("should return 401 if player not allowed for app", func() {
				status, _ := Get(app, "/apps/1234", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				goodDB := app.DB
				existingApp := CreateTestApp(app.DB)
				app.DB = faultyDb
				status, _ := Get(app, fmt.Sprintf("/apps/%s", existingApp.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the app does not exist", func() {
				status, _ := Get(app, fmt.Sprintf("/apps/%s", uuid.NewV4().String()), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				status, body := Get(app, "/apps/not-uuid", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})
		})
	})

	Describe("Put /apps/:id", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and the updated app without updating createdBy", func() {
				CreateTestUser(app.DB, map[string]interface{}{"email": "update@test.com", "isAdmin": true})
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "update@test.com")
				Expect(status).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["id"]).ToNot(BeNil())
				Expect(response["name"]).To(Equal(payload["name"]))
				Expect(response["bundleId"]).To(Equal(payload["bundleId"]))
				Expect(response["createdBy"]).To(Equal(existingApp.CreatedBy))
				Expect(int64(response["createdAt"].(float64))).To(Equal(existingApp.CreatedAt))
				Expect(response["updatedAt"]).ToNot(Equal(existingApp.UpdatedAt))

				id, err := uuid.FromString(response["id"].(string))
				Expect(err).NotTo(HaveOccurred())
				dbApp := &model.App{
					ID: id,
				}
				err = app.DB.Select(dbApp)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbApp).NotTo(BeNil())
				Expect(dbApp.Name).To(Equal(payload["name"]))
				Expect(dbApp.BundleID).To(Equal(payload["bundleId"]))
				Expect(dbApp.CreatedBy).To(Equal(existingApp.CreatedBy))
			})
		})

		Describe("Unsuccessfully", func() {
			It("should return 401 if no authenticated user", func() {
				existingApp := CreateTestApp(app.DB)
				status, _ := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "update@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 422 if app id is not UUID", func() {
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)

				status, body := Put(app, "/apps/not-uuid", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})

			It("should return 409 if app with same bundleId already exists", func() {
				existingApp1 := CreateTestApp(app.DB)
				existingApp2 := CreateTestApp(app.DB)
				payload := GetAppPayload(map[string]interface{}{"bundleId": existingApp1.BundleID})
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp2.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uix_apps_bundle_id"))
			})

			It("should return 422 if missing name", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				delete(payload, "name")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if missing bundleId", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				delete(payload, "bundleId")
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid bundleId"))
			})

			It("should return 422 if invalid name", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				payload["name"] = strings.Repeat("a", 256)
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid name"))
			})

			It("should return 422 if invalid bundleId", func() {
				existingApp := CreateTestApp(app.DB)
				payload := GetAppPayload()
				payload["bundleId"] = "invalidformat"
				pl, _ := json.Marshal(payload)
				status, body := Put(app, fmt.Sprintf("/apps/%s", existingApp.ID), string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid bundleId"))
			})
		})
	})

	Describe("Delete /apps/:id", func() {
		Describe("Sucesfully", func() {
			It("should return 204 ", func() {
				existingApp := CreateTestApp(app.DB)
				status, _ := Delete(app, fmt.Sprintf("/apps/%s", existingApp.ID), "test@test.com")
				Expect(status).To(Equal(http.StatusNoContent))

				dbApp := &model.App{
					ID: existingApp.ID,
				}
				err := app.DB.Select(dbApp)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(api.RecordNotFoundString))
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Delete(app, "/apps/1234", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				existingApp := CreateTestApp(app.DB)
				goodDB := app.DB
				app.DB = faultyDb
				status, _ := Delete(app, fmt.Sprintf("/apps/%s", existingApp.ID), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
				app.DB = goodDB
			})

			It("should return 404 if the app does not exist", func() {
				status, _ := Delete(app, fmt.Sprintf("/apps/%s", uuid.NewV4().String()), "test@test.com")

				Expect(status).To(Equal(http.StatusNotFound))
			})

			It("should return 422 if app id is not UUID", func() {
				status, body := Delete(app, "/apps/not-uuid", "test@test.com")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(ContainSubstring("uuid: incorrect UUID length: not-uuid"))
			})
		})
	})
})
