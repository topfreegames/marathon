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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/api"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
	"net/http"
)

var _ = Describe("API Application", func() {
	var logger zap.Logger
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app := api.GetApplication("127.0.0.1", 9999, true, logger, GetConfPath())
			Expect(app).NotTo(BeNil())
			Expect(app.Host).To(Equal("127.0.0.1"))
			Expect(app.Port).To(Equal(9999))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App authentication", func() {
		var app *api.Application

		BeforeEach(func() {
			app = api.GetApplication("127.0.0.1", 9999, false, logger, GetConfPath())
		})

		It("should fail if no user", func() {
			status, _ := Get(app, "/apps", "")
			Expect(status).To(Equal(http.StatusUnauthorized))
		})

		It("should work if valid user", func() {
			status, _ := Get(app, "/apps", "test@test.com")
			Expect(status).To(Equal(200))
		})
	})

	Describe("Error Handler", func() {
		var sink *TestBuffer
		BeforeEach(func() {
			sink = &TestBuffer{}
			logger = zap.New(
				zap.NewJSONEncoder(zap.NoTime()),
				zap.Output(sink),
				zap.WarnLevel,
			)
		})

		It("should handle errors when they are error objects", func() {
			app := api.GetApplication("127.0.0.1", 9999, false, logger, GetConfPath())

			app.OnErrorHandler(fmt.Errorf("some other error occurred"), []byte("stack"))
			result := sink.Buffer.String()
			var obj map[string]interface{}
			err := json.Unmarshal([]byte(result), &obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj["level"]).To(Equal("error"))
			Expect(obj["msg"]).To(Equal("Panic occurred."))
			Expect(obj["operation"]).To(Equal("OnErrorHandler"))
			Expect(obj["stack"]).To(Equal("stack"))
		})

		It("should panic on DB connect error", func() {
			Expect(func() { api.GetApplication("127.0.0.1", 9999, false, logger, GetFaultyDBConfPath()) }).To(PanicWith("cannot configure application"))
		})
	})
})
