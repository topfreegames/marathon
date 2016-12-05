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
	"github.com/topfreegames/marathon/api"
	"github.com/topfreegames/marathon/model"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("App Handler", func() {
	var logger zap.Logger
	var faultyDb *gorm.DB
	var app *api.Application
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		app = GetDefaultTestApp(logger)
		faultyDb = GetFaultyTestDB()
	})

	Describe("Get /apps", func() {
		Describe("Sucesfully", func() {
			It("should return 200 and an empty list of apps if there are no apps", func() {
				status, body := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				response := []model.App{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(0))
			})

			It("should return 200 and a list of apps", func() {
				testApps := CreateTestApps(app.DB, 10)
				status, body := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusOK))

				response := []model.App{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(HaveLen(10))

				for idx, app := range response {
					Expect(app.ID).ToNot(BeNil())
					Expect(app.Name).To(Equal(testApps[idx].Name))
					Expect(app.BundleID).To(Equal(testApps[idx].BundleID))
					Expect(app.CreatedBy).To(Equal(testApps[idx].CreatedBy))
					Expect(app.CreatedAt).ToNot(BeNil())
					Expect(app.UpdatedAt).ToNot(BeNil())
				}
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Get(app, "/apps", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				app.DB = faultyDb
				status, _ := Get(app, "/apps", "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("Post /apps", func() {
		Describe("Sucesfully", func() {
			It("should return 201 and the created app", func() {
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "success@test.com")
				Expect(status).To(Equal(http.StatusCreated))

				var response model.App
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.ID).ToNot(BeNil())
				Expect(response.Name).To(Equal(payload["name"]))
				Expect(response.BundleID).To(Equal(payload["bundleId"]))
				Expect(response.CreatedBy).To(Equal("success@test.com"))
				Expect(response.CreatedAt).ToNot(BeNil())
				Expect(response.UpdatedAt).ToNot(BeNil())
			})
		})

		Describe("Unsucesfully", func() {
			It("should return 401 if no authenticated user", func() {
				status, _ := Post(app, "/apps", "", "")

				Expect(status).To(Equal(http.StatusUnauthorized))
			})

			It("should return 500 if some error occured", func() {
				app.DB = faultyDb
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, _ := Post(app, "/apps", string(pl), "test@test.com")

				Expect(status).To(Equal(http.StatusInternalServerError))
			})

			It("should return 409 if app with same bundleId already exists", func() {
				extistingApp := CreateTestApp(app.DB)
				payload := GetAppPayload(map[string]interface{}{"bundleId": extistingApp.BundleID})
				fmt.Println(extistingApp.BundleID)
				fmt.Println(payload["bundleId"])
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "test@test.com")
				Expect(status).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("pq: duplicate key value violates unique constraint \"uix_apps_bundle_id\""))
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

			It("should return 422 if invalid auth header", func() {
				payload := GetAppPayload()
				pl, _ := json.Marshal(payload)
				status, body := Post(app, "/apps", string(pl), "not-a-valid-email")
				Expect(status).To(Equal(http.StatusUnprocessableEntity))

				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["reason"]).To(Equal("invalid createdBy"))
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
})
