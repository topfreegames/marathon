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

package lib

import (
	"context"
	"net/http"
	"time"
)

// Marathon is the marathon client.
// Implements MarathonInterface.
type Marathon struct {
	httpClient *http.Client
	url        string
	userEmail  string
	appID      string
}

// Config is the configuration struct
// for Marathon http client.
type Config struct {
	Timeout   time.Duration
	URL       string
	UserEmail string
	AppID     string
}

// NewMarathon returns a new Marathon lib
func NewMarathon(config *Config) *Marathon {
	if config.Timeout.Seconds() == 0 {
		config.Timeout = 500 * time.Millisecond
	}

	m := &Marathon{
		httpClient: getHTTPClient(config),
		url:        config.URL,
		userEmail:  config.UserEmail,
		appID:      config.AppID,
	}

	return m
}

// CreateJob access Marathon API to create a job
// using template and payload.
func (m *Marathon) CreateJob(
	ctx context.Context,
	template string,
	payload *CreateJobPayload,
) (*Job, error) {
	route := m.buildCreateJobURL(template)
	response := &Job{}

	err := m.sendTo(ctx, http.MethodPost, route, payload, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ListJobs access Marathon API to list jobs
// of template.
func (m *Marathon) ListJobs(
	ctx context.Context,
	template string,
) ([]*Job, error) {
	route := m.buildListJobsURL(template)
	var response []*Job

	err := m.sendTo(ctx, http.MethodGet, route, nil, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
