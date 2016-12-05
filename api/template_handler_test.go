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
	var existingApp *model.App
	var baseRoute string
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		app = GetDefaultTestApp(logger)
		faultyDb = GetFaultyTestDB(app)

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
		})
	})
})
