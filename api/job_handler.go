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
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
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
	templateName := c.QueryParam("template")
	if templateName == "" {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: "template name must be specified"})
	}
	email := c.Get("user-email").(string)
	job := &model.Job{
		ID:           uuid.NewV4(),
		AppID:        aid,
		TemplateName: templateName,
		CreatedBy:    email,
		CreatedAt:    time.Now().UnixNano(),
		UpdatedAt:    time.Now().UnixNano(),
	}
	err = WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, job)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
	}

	template := &model.Template{}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Model(&template).Column("template.*").Where("template.app_id = ?", aid).Where("template.name = ?", templateName).First()
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
		}
		log.E(l, "Failed to create job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}

	err = WithSegment("db-insert", c, func() error {
		return a.DB.Insert(&job)
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, job)
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
		}
		log.E(l, "Failed to create job.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	a.Logger.Debug("job successfully created! creating job in create_batches_worker")
	var wJobID string
	err = WithSegment("create-job", c, func() error {
		if job.StartsAt != 0 {
			if len(job.CSVPath) == 0 {
				wJobID, err = a.Worker.ScheduleCreateBatchesJob(&[]string{job.ID.String()}, job.StartsAt)
			} else {
				wJobID, err = a.Worker.ScheduleCreateBatchesFromFiltersJob(&[]string{job.ID.String()}, job.StartsAt)
			}
		} else {
			if len(job.CSVPath) == 0 {
				wJobID, err = a.Worker.CreateBatchesJob(&[]string{job.ID.String()})
			} else {
				wJobID, err = a.Worker.CreateBatchesFromFiltersJob(&[]string{job.ID.String()})
			}
		}
		return err
	})

	if err != nil {
		a.DB.Delete(&job)
		log.E(l, "Failed to send job to create_batches_worker.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	log.I(l, "Job successfully sent to create_batches_worker", func(cm log.CM) {
		cm.Write(zap.String("workerJobId", wJobID))
	})
	return c.JSON(http.StatusCreated, job)
}

// GetJobHandler is the method called when a get to /apps/:aid/templates/:templateName/jobs/:jid is called
func (a *Application) GetJobHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "jobHandler"),
		zap.String("operation", "getJob"),
		zap.String("appId", c.Param("aid")),
		zap.String("template", c.QueryParam("template")),
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
