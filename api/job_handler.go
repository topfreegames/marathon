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

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// ListJobsHandler is the method called when a get to /apps/:id/templates/:templateName/jobs is called
func (a *Application) ListJobsHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	templateName := c.QueryParam("template")
	if templateName == "" {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: "template name must be specified"})
	}
	jobs := []model.Job{}
	if err := a.DB.Model(&jobs).Column("job.*", "App").Where("job.template_name = ?", templateName).Where("job.app_id = ?", id).Select(); err != nil {
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	return c.JSON(http.StatusOK, jobs)
}

// PostJobHandler is the method called when a post to /apps/:id/templates/:templateName/jobs is called
func (a *Application) PostJobHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
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
		AppID:        id,
		TemplateName: templateName,
		CreatedBy:    email,
	}
	err = decodeAndValidate(c, job)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
	}

	template := &model.Template{}
	if err := a.DB.Model(&template).Column("template.*").Where("template.app_id = ?", id).Where("template.name = ?", templateName).First(); err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}

	if err = a.DB.Insert(&job); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, job)
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: job})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	a.Logger.Debug("job successfully created! creating job in create_batches_worker")
	jid, err := a.Worker.CreateBatchesJob(&[]string{job.ID.String()})
	if err != nil {
		a.DB.Delete(&job)
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	a.Logger.Info("job successfully sent to create_batches_worker!", zap.String("jid", jid))
	return c.JSON(http.StatusCreated, job)
}

// GetJobHandler is the method called when a get to /apps/:id/templates/:templateName/jobs/:jid is called
func (a *Application) GetJobHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	jid, err := uuid.FromString(c.Param("jid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	job := &model.Job{
		ID:    jid,
		AppID: id,
	}
	if err := a.DB.Model(&job).Column("job.*", "App").Where("job.id = ?", job.ID).Select(); err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, job)
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: job})
	}
	return c.JSON(http.StatusOK, job)
}
