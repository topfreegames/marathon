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

package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo"
	newrelic "github.com/newrelic/go-agent"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/model"
)

// RecordNotFoundString is the string returned when a record is not found
var RecordNotFoundString = "pg: no rows in result set"

//Error is a struct to help return errors
type Error struct {
	Reason string          `json:"reason"`
	Value  InputValidation `json:"value"`
}

//GetTX returns new relic transaction
func GetTX(c echo.Context) newrelic.Transaction {
	tx := c.Get("txn")
	if tx == nil {
		return nil
	}

	return tx.(newrelic.Transaction)
}

//WithSegment adds a segment to new relic transaction
func WithSegment(name string, c echo.Context, f func() error) error {
	tx := GetTX(c)
	if tx == nil {
		return f()
	}
	segment := newrelic.StartSegment(tx, name)
	defer segment.End()
	return f()
}

//SendCreatedJobEmail builds a created job email message and sends it with sendgrid
func SendCreatedJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, app *model.App) error {
	action := "created"
	if job.StartsAt != 0 {
		action = "scheduled"
	}
	subject := fmt.Sprintf("New push job %s", action)

	strategy := ""
	if job.PastTimeStrategy != "" {
		strategy = fmt.Sprintf("(strategy for push in the past: %s)", job.PastTimeStrategy)
	}

	var extraInfo string

	if len(job.CSVPath) > 0 {
		extraInfo = fmt.Sprintf("This job uses the following csvPath: %s.", job.CSVPath)
	} else if len(job.Filters) > 0 {
		filters, _ := json.MarshalIndent(job.Filters, "", "  ")
		extraInfo = fmt.Sprintf("This job uses the following filters: \n%s.", string(filters))
	} else {
		extraInfo = "This job has no specified filters or csvPath."
	}

	var platform string
	if job.Service == "apns" {
		platform = "iOS"
	} else if job.Service == "gcm" {
		platform = "Android"
	} else {
		platform = fmt.Sprintf("Unknown platform for service %s", job.Service)
	}

	var scheduledInfo string
	if job.StartsAt != 0 {
		scheduledInfo = fmt.Sprintf("(%s)", time.Unix(0, job.StartsAt).UTC().Format(time.RFC1123))
	} else {
		scheduledInfo = ""
	}

	message := fmt.Sprintf(`
Hi there, a new push job was %s.

App: %s
Template: %s
Platform: %s
JobID: %s
Scheduled: %t %s
Localized: %t %s

%s
`, action, app.Name, job.TemplateName, platform, job.ID, job.StartsAt != 0, scheduledInfo, job.Localized, strategy, extraInfo)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message)
}

//SendPausedJobEmail builds a paused job email message and sends it with sendgrid
func SendPausedJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName string, expireAt int64) error {
	subject := "Push job entered paused state"

	var platform string
	if job.Service == "apns" {
		platform = "iOS"
	} else if job.Service == "gcm" {
		platform = "Android"
	} else {
		platform = fmt.Sprintf("Unknown platform for service %s", job.Service)
	}

	expireAtDate := fmt.Sprintf("(%s)", time.Unix(0, expireAt).UTC().Format(time.RFC1123))

	message := fmt.Sprintf(`
Hello, your push job status has changed to paused.

App: %s
Template: %s
Platform: %s
JobID: %s

This job will be removed from the paused queue on %s. After this date the job will no longer be available.
Please resume or stop it before then.
`, appName, job.TemplateName, platform, job.ID, expireAtDate)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message)
}
