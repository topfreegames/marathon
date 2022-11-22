/*
 * Copyright (c) 2019 TFG Co <backend@tfgco.com>
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

package lib_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jarcoal/httpmock"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/lib"
)

var _ = Describe("Lib", func() {
	var m lib.MarathonInterface

	ctx := context.Background()
	now := time.Now()
	appID := uuid.NewV4()
	template := "template"
	jobID := uuid.NewV4()
	config := &lib.Config{
		Timeout:   100 * time.Millisecond,
		URL:       "http://marathon",
		UserEmail: "user@email.com",
		AppID:     appID.String(),
	}

	BeforeSuite(func() {
		httpmock.Activate()
	})

	BeforeEach(func() {
		//default configs for each test
		m = lib.NewMarathon(config)
		httpmock.Reset()
	})

	toString := func(m interface{}) string {
		bytes, err := json.Marshal(m)
		if err != nil {
			Fail(err.Error())
		}

		return string(bytes)
	}

	Describe("Creating new job", func() {
		It("should create job", func() {
			url := fmt.Sprintf("http://marathon/apps/%s/jobs", appID)
			response := &lib.Job{
				ID:           jobID.String(),
				TotalBatches: 10,
				TotalUsers:   100,
				ExpiresAt:    now.Add(time.Hour).Unix(),
				StartsAt:     now.Unix(),
				AppID:        appID.String(),
				CSVPath:      "protocol://host/path",
			}

			httpmock.RegisterResponder(
				"POST", url,
				httpmock.NewStringResponder(200, toString(response)))

			job, err := m.CreateJob(ctx, template, &lib.CreateJobPayload{
				Localized: true,
				ExpiresAt: now.Add(time.Hour).Unix(),
				StartsAt:  now.Unix(),
				CSVPath:   "protocol://host/path",
			})

			Expect(err).To(BeNil())
			Expect(job).To(Equal(response))
		})

		It("should return error on status 500", func() {
			url := fmt.Sprintf("http://marathon/apps/%s/jobs", appID)
			httpmock.RegisterResponder(
				"POST", url,
				httpmock.NewStringResponder(500, toString(map[string]interface{}{
					"error": "error",
				})))

			job, err := m.CreateJob(ctx, template, &lib.CreateJobPayload{
				Localized: true,
				ExpiresAt: now.Add(time.Hour).Unix(),
				StartsAt:  now.Unix(),
				CSVPath:   "protocol://host/path",
			})

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(`Request error. Status code: 500. Body: {"error":"error"}`))
			Expect(job).To(BeNil())
		})
	})

	Describe("Listing jobs", func() {
		It("should create job", func() {
			url := fmt.Sprintf("http://marathon/apps/%s/jobs", appID)
			response := []*lib.Job{
				{
					ID:           jobID.String(),
					TotalBatches: 10,
					TotalUsers:   100,
					ExpiresAt:    now.Add(time.Hour).Unix(),
					StartsAt:     now.Unix(),
					AppID:        appID.String(),
				},
			}

			httpmock.RegisterResponder(
				"GET", url,
				httpmock.NewStringResponder(200, toString(response)))

			jobs, err := m.ListJobs(ctx, template)

			Expect(err).To(BeNil())
			Expect(jobs).To(Equal(response))
		})
	})
})
