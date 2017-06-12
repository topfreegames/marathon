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

package email

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/model"
)

func getPlatformFromService(service string) string {
	var platform string
	if service == "apns" {
		platform = "iOS"
	} else if service == "gcm" {
		platform = "Android"
	} else {
		platform = fmt.Sprintf("Unknown platform for service %s", service)
	}
	return platform
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

	platform := getPlatformFromService(job.Service)

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
CreatedBy: %s
Scheduled: %t %s
Localized: %t %s

%s
`, action, app.Name, job.TemplateName, platform, job.ID, job.CreatedBy, job.StartsAt != 0, scheduledInfo, job.Localized, strategy, extraInfo)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message, false)
}

//SendPausedJobEmail builds a paused job email message and sends it with sendgrid
func SendPausedJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName string, expireAt int64) error {
	subject := "Push job entered paused state"
	platform := getPlatformFromService(job.Service)
	expireAtDate := fmt.Sprintf("(%s)", time.Unix(0, expireAt).UTC().Format(time.RFC1123))

	message := fmt.Sprintf(`
Hello, your push job status has changed to paused.

App: %s
Template: %s
Platform: %s
JobID: %s
CreatedBy: %s

This job will be removed from the paused queue on %s. After this date the job will no longer be available.
Please resume or stop it before then.
`, appName, job.TemplateName, platform, job.ID, job.CreatedBy, expireAtDate)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message, false)
}

//SendStoppedJobEmail builds a stopped job email message and sends it with sendgrid
func SendStoppedJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName, stoppedBy string) error {
	subject := "Push job stopped"
	platform := getPlatformFromService(job.Service)

	message := fmt.Sprintf(`
Hello, your push job status has changed to stopped.

StoppedBy: %s

App: %s
Template: %s
Platform: %s
JobID: %s
CreatedBy: %s

This action is irreversible and this job's push notifications will no longer be sent.
`, stoppedBy, appName, job.TemplateName, platform, job.ID, job.CreatedBy)

	skipBlacklist := strings.Contains(stoppedBy, "automatically")
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message, skipBlacklist)
}

//SendCircuitBreakJobEmail builds a circuit break job email message and sends it with sendgrid
func SendCircuitBreakJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName string, expireAt int64) error {
	subject := "Push job entered circuit break state"
	platform := getPlatformFromService(job.Service)
	expireAtDate := fmt.Sprintf("(%s)", time.Unix(0, expireAt).UTC().Format(time.RFC1123))

	message := fmt.Sprintf(`
Hello, your push job status has changed to circuit break.

App: %s
Template: %s
Platform: %s
JobID: %s
CreatedBy: %s

This job will be removed from the paused queue on %s. After this date the job will no longer be available.
Please fix the issues causing the circuit break and resume or stop it before then.
`, appName, job.TemplateName, platform, job.ID, job.CreatedBy, expireAtDate)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message, true)
}

//SendJobCompletedEmail builds a job complete email message and sends it with sendgrid
func SendJobCompletedEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName string) error {
	subject := "Push job completed"
	platform := getPlatformFromService(job.Service)

	var ack int
	if val, ok := job.Feedbacks["ack"]; ok {
		ack = int(val.(float64))
	}
	feedbacks := make([]string, len(job.Feedbacks))
	idx := 0
	for key, val := range job.Feedbacks {
		intVal := int(val.(float64))
		feedbacks[idx] = fmt.Sprintf("- %s: %.2f%% (%d)\n", key, float64(100*intVal)/float64(job.TotalTokens), intVal)
		idx++
	}

	stats := fmt.Sprintf(`
  Messages sent to Kafka:
    Batches: %.2f%% (%d/%d)
    Tokens: %.2f%% (%d/%d)

  %s Feedbacks
    Success/Total Tokens: %.2f%% (%d)

Feedbacks:
%s
`,
		float64(100*job.CompletedBatches)/float64(job.TotalBatches),
		job.CompletedBatches,
		job.TotalBatches,
		float64(100*job.CompletedTokens)/float64(job.TotalTokens),
		job.CompletedTokens,
		job.TotalTokens,
		strings.ToUpper(job.Service),
		float64(100*ack)/float64(job.TotalTokens),
		ack,
		strings.Join(feedbacks, ""),
	)

	message := fmt.Sprintf(`
Hello, your push job is complete.

App: %s
Template: %s
Platform: %s
JobID: %s
CreatedBy: %s

Stats:
%s
`, appName, job.TemplateName, platform, job.ID, job.CreatedBy, stats)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message, false)
}
