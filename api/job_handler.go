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
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/pg.v5/types"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

// ListJobsHandler is the method called when a get to /apps/:aid/templates/:templateName/jobs is called
func (a *Application) ListJobsHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "listJobs"),
		zap.String("appId", c.Param("aid")),
		zap.String("template", c.QueryParam("template")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	templateName := c.QueryParam("template")
	jobs := []model.Job{}
	query := a.DB.Model(&jobs).Column("job.*", "App").Where("job.app_id = ?", aid)
	if templateName != "" {
		query.Where("job.template_name = ?", templateName)
	}
	err = WithSegment("db-select", c, func() error {
		return query.Select()
	})
	if err != nil {
		log.E(l, "Failed to list jobs.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	log.D(l, "Listed jobs successfully.", func(cm log.CM) {
		cm.Write(zap.Object("jobs", jobs))
	})
	return c.JSON(http.StatusOK, jobs)
}

// PostJobHandler is the method called when a post to /apps/:aid/templates/:templateName/jobs is called
func (a *Application) PostJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "postJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("template", c.QueryParam("template")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}

	app := &model.App{ID: aid}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Select(&app)
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: "App not found with given id."})
		}
		log.E(l, "Failed to retrieve app.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}

	templateName := c.QueryParam("template")
	if templateName == "" {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: "template name must be specified"})
	}
	userEmail := c.Get("user-email").(string)
	job := &model.Job{
		ID:           uuid.NewV4(),
		AppID:        aid,
		TemplateName: templateName,
		CreatedBy:    userEmail,
		CreatedAt:    time.Now().UnixNano(),
		UpdatedAt:    time.Now().UnixNano(),
		App:          *app,
	}
	err = WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, job)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
	}

	err = a.checkFilters(job, c)
	if err != nil {
		return err
	}

	err = a.checkTemplateName(templateName, job, c)
	if err != nil {
		return err
	}

	if job.StartsAt == 0 && job.Localized {
		localeErr := "Job can not be localized and don't have an start time"
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: localeErr, Value: job})
	}

	// check if was an timezone conflict
	if job.StartsAt != 0 && job.Localized {
		// get smallast timezone
		localizedAt := time.Unix(0, job.StartsAt).Add(-12 * time.Hour)
		localizedAt = localizedAt.Add(-15 * time.Minute)
		// we'll use 1 to have some marging when creating the jobs
		localizedAt = localizedAt.Add(-1 * time.Minute)
		if localizedAt.Before(time.Now()) {
			localeErr := "The selected time is invalid because it already has started in some places."
			localeErr += "\nUse at least the current UTC time -12 hours."
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: localeErr, Value: job})
		}
	}

	err = WithSegment("create-job", c, func() error {
		scheduleJob := job.StartsAt
		if scheduleJob != 0 && job.Localized {
			// create a job for each tz
			for i := -12; i <= 14; i++ {
				tzs := []string{
					fmt.Sprintf("%+.4d", i*100-55), // 100 - 55 = 45
					fmt.Sprintf("%+.4d", i*100),
					fmt.Sprintf("%+.4d", i*100+15),
					fmt.Sprintf("%+.4d", i*100+30),
				}
				sendTime := time.Unix(0, scheduleJob).Add(time.Duration(i) * time.Hour)

				if sendTime.Before(time.Now()) {
					if job.PastTimeStrategy == "skip" {
						continue
					}
					sendTime = sendTime.Add(time.Duration(i) * time.Hour)
				}

				job.StartsAt = sendTime.UnixNano()
				job.Filters["tz"] = strings.Join(tzs, ",")
				job.ID = uuid.NewV4()
				log.I(l, "Create a timezone job.")

				err = a.createJob(job, c)
				if err != nil {
					return err
				}
			}
		} else {
			log.I(l, "Create a simple job.")
			return a.createJob(job, c)
		}
		return nil
	})

	if err != nil {
		log.E(l, "Failed to send job to create_batches_worker.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}

	if a.SendgridClient != nil {
		log.D(l, "sending email with job info")
		app := &model.App{ID: aid}
		a.DB.Select(&app)

		err := email.SendCreatedJobEmail(a.SendgridClient, job, app)
		if err != nil {
			log.E(l, "Failed to send email with job info.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
		}
		log.I(l, "Successfully sent email with job info.")
	}
	return c.JSON(http.StatusCreated, job)
}

func (a *Application) checkFilters(job *model.Job, c echo.Context) error {
	if job.Filters["region"] != nil || job.Filters["NOTregion"] != nil || job.Filters["locale"] != nil || job.Filters["NOTlocale"] != nil {
		var users []worker.User
		query := fmt.Sprintf("SELECT locale, region FROM %s WHERE locale is not NULL AND region is not NULL LIMIT 1;", worker.GetPushDBTableName(job.App.Name, job.Service))
		a.PushDB.Query(&users, query)
		if len(users) != 1 {
			return c.JSON(http.StatusInternalServerError, &Error{Reason: "Failed to check filters in Push DB"})
		}
		locale := users[0].Locale
		region := users[0].Region

		localeSettings := map[string]bool{
			"isUpperCase": strings.ToUpper(locale) == locale,
			"isLowerCase": strings.ToLower(locale) == locale,
		}

		regionSettings := map[string]bool{
			"isUpperCase": strings.ToUpper(region) == region,
			"isLowerCase": strings.ToLower(region) == region,
		}

		if job.Filters["locale"] != nil {
			if localeSettings["isUpperCase"] && !localeSettings["isLowerCase"] {
				job.Filters["locale"] = strings.ToUpper(job.Filters["locale"].(string))
			} else if localeSettings["isLowerCase"] && !localeSettings["isUpperCase"] {
				job.Filters["locale"] = strings.ToLower(job.Filters["locale"].(string))
			} else {
				return c.JSON(http.StatusInternalServerError, &Error{Reason: "Locale case check failed in Push DB"})
			}
		}

		if job.Filters["NOTlocale"] != nil {
			if localeSettings["isUpperCase"] && !localeSettings["isLowerCase"] {
				job.Filters["NOTlocale"] = strings.ToUpper(job.Filters["NOTlocale"].(string))
			} else if localeSettings["isLowerCase"] && !localeSettings["isUpperCase"] {
				job.Filters["NOTlocale"] = strings.ToLower(job.Filters["NOTlocale"].(string))
			} else {
				return c.JSON(http.StatusInternalServerError, &Error{Reason: "Locale case check failed in Push DB"})
			}
		}

		if job.Filters["region"] != nil {
			if regionSettings["isUpperCase"] && !regionSettings["isLowerCase"] {
				job.Filters["region"] = strings.ToUpper(job.Filters["region"].(string))
			} else if regionSettings["isLowerCase"] && !regionSettings["isUpperCase"] {
				job.Filters["region"] = strings.ToLower(job.Filters["region"].(string))
			} else {
				return c.JSON(http.StatusInternalServerError, &Error{Reason: "Region case check failed in Push DB"})
			}
		}

		if job.Filters["NOTregion"] != nil {
			if regionSettings["isUpperCase"] && !regionSettings["isLowerCase"] {
				job.Filters["NOTregion"] = strings.ToUpper(job.Filters["NOTregion"].(string))
			} else if regionSettings["isLowerCase"] && !regionSettings["isUpperCase"] {
				job.Filters["NOTregion"] = strings.ToLower(job.Filters["NOTregion"].(string))
			} else {
				return c.JSON(http.StatusInternalServerError, &Error{Reason: "Region case check failed in Push DB"})
			}
		}
	}
	return nil
}

func (a *Application) checkTemplateName(templateName string, job *model.Job, c echo.Context) error {
	for _, tpl := range strings.Split(templateName, ",") {
		template := &model.Template{}
		err := WithSegment("db-select", c, func() error {
			return a.DB.Model(&template).Column("template.*").Where("template.app_id = ?", job.AppID).Where("template.name = ?", tpl).First()
		})
		if err != nil {
			if err.Error() == RecordNotFoundString {
				return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
			}
			return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
		}

		err = WithSegment("db-select", c, func() error {
			return a.DB.Model(&template).Column("template.*").Where("template.app_id = ?", job.AppID).Where("template.name = ? AND template.locale='en'", tpl).First()
		})
		if err != nil {
			if err.Error() == RecordNotFoundString {
				localeErr := "Cannot create job if there is no template for locale 'en'."
				return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: localeErr, Value: job})
			}
			return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
		}
	}
	return nil
}

func (a *Application) createJobWorkers(job *model.Job, c echo.Context) error {
	var err error
	if job.StartsAt != 0 {
		if len(job.CSVPath) > 0 {
			_, err = a.Worker.ScheduleCSVSplitJob(job, job.StartsAt)
		} else {
			err = a.Worker.ScheduleDirectBatchesJob(job, job.StartsAt)
		}
	} else {
		if len(job.CSVPath) > 0 {
			_, err = a.Worker.CreateCSVSplitJob(job)
		} else {
			err = a.Worker.CreateDirectBatchesJob(job)
		}
	}
	return err
}

func (a *Application) createJob(job *model.Job, c echo.Context) error {
	err := WithSegment("db-insert", c, func() error {
		return a.DB.Insert(&job)
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, job)
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	a.createJobWorkers(job, c)

	return nil
}

// GetJobHandler is the method called when a get to /apps/:aid/templates/:templateName/jobs/:jid is called
func (a *Application) GetJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "getJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("jobId", c.Param("jid")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	jid, err := uuid.FromString(c.Param("jid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	job := &model.Job{
		ID:    jid,
		AppID: aid,
	}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Model(&job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	})
	a.DB.Model(&job.StatusEvents).Where("job_id = ?", job.ID).Column("status.*", "Events").Select()
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, job)
		}
		log.E(l, "Failed to retrieve job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	log.D(l, "Retrieved job successfully.", func(cm log.CM) {
		cm.Write(zap.Object("job", job))
	})
	return c.JSON(http.StatusOK, job)
}

// PauseJobHandler is the method called when a put to apps/:id/jobs/:jid/pause is called
func (a *Application) PauseJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "pauseJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("jobId", c.Param("jid")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	jid, err := uuid.FromString(c.Param("jid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	userEmail := c.Get("user-email").(string)
	job := &model.Job{
		ID:        jid,
		AppID:     aid,
		CreatedBy: userEmail,
		Status:    "paused",
		UpdatedAt: time.Now().UnixNano(),
	}
	prevJob := &model.Job{}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Model(&prevJob).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, job)
		}
		log.E(l, "Failed to retrieve job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: prevJob})
	}
	if prevJob.Status != "" {
		return c.JSON(http.StatusForbidden, &Error{Reason: fmt.Sprintf("cannot pause %s job", prevJob.Status)})
	}
	err = WithSegment("db-update", c, func() error {
		_, err = a.DB.Model(&job).Column("status").Column("updated_at").Returning("*").Update()
		return err
	})
	if err != nil {
		log.E(l, "Failed to pause job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	log.D(l, "Updated job successfully.", func(cm log.CM) {
		cm.Write(zap.Object("job", job))
	})

	if a.SendgridClient != nil {
		log.D(l, "sending email with paused job info")
		app := &model.App{ID: aid}
		a.DB.Select(&app)

		expireAt := time.Now().Add(7 * 24 * time.Hour).UnixNano()
		err := email.SendPausedJobEmail(a.SendgridClient, job, app.Name, expireAt)
		if err != nil {
			log.E(l, "Failed to send email with paused job info.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
		}
		log.I(l, "Successfully sent email with paused job info.")
	}
	return c.JSON(http.StatusOK, job)
}

// StopJobHandler is the method called when a put to apps/:id/jobs/:jid/stop is called
func (a *Application) StopJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "stopJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("jobId", c.Param("jid")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	jid, err := uuid.FromString(c.Param("jid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	userEmail := c.Get("user-email").(string)
	job := &model.Job{
		ID:        jid,
		AppID:     aid,
		CreatedBy: userEmail,
		Status:    "stopped",
		UpdatedAt: time.Now().UnixNano(),
	}
	var values *types.Result
	err = WithSegment("db-update", c, func() error {
		values, err = a.DB.Model(&job).Column("status").Column("updated_at").Returning("*").Update()
		return err
	})
	if err != nil {
		log.E(l, "Failed to stop job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	if values.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{})
	}
	log.D(l, "Updated job successfully.", func(cm log.CM) {
		cm.Write(zap.Object("job", job))
	})

	if a.SendgridClient != nil {
		log.D(l, "sending email with stopped job info")
		app := &model.App{ID: aid}
		a.DB.Select(&app)

		err := email.SendStoppedJobEmail(a.SendgridClient, job, app.Name, userEmail)
		if err != nil {
			log.E(l, "Failed to send email with stopped job info.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
		}
		log.I(l, "Successfully sent email with stopped job info.")
	}
	return c.JSON(http.StatusOK, job)
}

// ResumeJobHandler is the method called when a put to apps/:id/jobs/:jid/resume is called
func (a *Application) ResumeJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "resumeJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("jobId", c.Param("jid")),
	)
	aid, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	jid, err := uuid.FromString(c.Param("jid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	userEmail := c.Get("user-email").(string)
	prevJob := &model.Job{}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Model(&prevJob).Column("job.*", "App").Where("job.id = ?", jid).Select()
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, prevJob)
		}
		log.E(l, "Failed to retrieve job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: prevJob})
	}
	if prevJob.Status != "paused" && prevJob.Status != "circuitbreak" {
		return c.JSON(http.StatusForbidden, &Error{Reason: "cannot resume job with status other than paused/circuitbreak"})
	}

	var wJobID string
	err = WithSegment("resume-job", c, func() error {
		wJobID, err = a.Worker.CreateResumeJob(&[]string{prevJob.ID.String()})
		return err
	})

	if err != nil {
		log.E(l, "Failed to send job to resume_job_worker.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}

	log.I(l, "Job successfully sent to resume_job_worker", func(cm log.CM) {
		cm.Write(zap.String("workerJobId", wJobID))
	})

	job := &model.Job{
		ID:        jid,
		AppID:     aid,
		CreatedBy: userEmail,
		Status:    "",
		UpdatedAt: time.Now().UnixNano(),
	}
	err = WithSegment("db-update", c, func() error {
		_, err = a.DB.Model(&job).Column("status").Column("updated_at").Returning("*").Update()
		return err
	})
	if err != nil {
		log.E(l, "Failed to resume job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	log.D(l, "Resumed job successfully.", func(cm log.CM) {
		cm.Write(zap.Object("job", job))
	})
	return c.JSON(http.StatusOK, job)
}
