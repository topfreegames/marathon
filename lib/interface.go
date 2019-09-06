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
)

// MarathonInterface defines the interface of marathon client
// to access the API.
type MarathonInterface interface {
	CreateJob(
		ctx context.Context,
		template string,
		payload *CreateJobPayload,
	) (*Job, error)

	ListJobs(
		ctx context.Context,
		template string,
	) ([]*Job, error)
}

// JSON is a generic json map
type JSON map[string]interface{}

// CreateJobPayload contains the parameters for CreateJob method
type CreateJobPayload struct {
	Localized        bool        `json:"localized"`
	ExpiresAt        int64       `json:"expiresAt"`
	StartsAt         int64       `json:"startsAt"`
	Context          JSON        `json:"context"`
	Service          string      `json:"service"`
	Filters          JSON        `json:"filters"`
	Metadata         JSON        `json:"metadata"`
	CSVPath          string      `json:"csvPath"`
	PastTimeStrategy interface{} `json:"pastTimeStrategy"`
	ControlGroup     float64     `json:"controlGroup"`
}

// Job contains job information
type Job struct {
	ID                  string  `json:"id"`
	TotalBatches        int     `json:"totalBatches"`
	CompletedBatches    int     `json:"completedBatches"`
	TotalUsers          int     `json:"totalUsers"`
	CompletedUsers      int     `json:"completedUsers"`
	CompletedTokens     int     `json:"completedTokens"`
	DBPageSize          int     `json:"dbPageSize"`
	Localized           bool    `json:"localized"`
	CompletedAt         int64   `json:"completedAt"`
	ExpiresAt           int64   `json:"expiresAt"`
	StartsAt            int64   `json:"startsAt"`
	Context             JSON    `json:"context"`
	Service             string  `json:"service"`
	Filters             JSON    `json:"filters"`
	Metadata            JSON    `json:"metadata"`
	CSVPath             string  `json:"csvPath"`
	TemplateName        string  `json:"templateName"`
	PastTimeStrategy    string  `json:"pastTimeStrategy"`
	Status              string  `json:"status"`
	AppID               string  `json:"appId"`
	CreatedBy           string  `json:"createdBy"`
	CreatedAt           int64   `json:"createdAt"`
	UpdatedAt           int64   `json:"updatedAt"`
	ControlGroup        float64 `json:"controlGroup"`
	ControlGroupCsvPath string  `json:"controlGroupCsvPath"`
}
