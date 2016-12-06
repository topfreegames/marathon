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
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/api"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

// TODO: find out how to test the other cases
var _ = Describe("Upload Handler", func() {
	var logger zap.Logger
	var app *api.Application
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		app = GetDefaultTestApp(logger)
	})

	Describe("Get /uploadurl", func() {
		It("should return 500 if AWS credentials are invalid", func() {
			status, body := Get(app, "/uploadurl", "test@testmail.com")

			Expect(status).To(Equal(http.StatusInternalServerError))

			var response map[string]interface{}
			err := json.Unmarshal([]byte(body), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["reason"]).To(Equal("The AWS Access Key Id you provided does not exist in our records."))
		})

		It("should return 401 if no authenticated user", func() {
			status, _ := Get(app, "/uploadurl", "")

			Expect(status).To(Equal(http.StatusUnauthorized))
		})

	})
})
