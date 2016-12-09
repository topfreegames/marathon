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
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// ListTemplatesHandler is the method called when a get to /apps/:id/templates is called
func (a *Application) ListTemplatesHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "templateHandler"),
		zap.String("operation", "listTemplates"),
		zap.String("appId", c.Param("id")),
	)
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	templates := []model.Template{}
	if err := a.DB.Model(&templates).Column("template.*", "App").Where("template.app_id = ?", id).Select(); err != nil {
		log.E(l, "Failed to list templates.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	log.D(l, "Listed templates successfully.", func(cm log.CM) {
		cm.Write(zap.Object("templates", templates))
	})
	return c.JSON(http.StatusOK, templates)
}

// PostTemplateHandler is the method called when a post to /apps/:id/templates is called
func (a *Application) PostTemplateHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "templateHandler"),
		zap.String("operation", "postTemplate"),
		zap.String("appId", c.Param("id")),
	)
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	email := c.Get("user-email").(string)
	template := &model.Template{
		ID:        uuid.NewV4(),
		AppID:     id,
		CreatedBy: email,
	}
	err = decodeAndValidate(c, template)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: template})
	}
	if err = a.DB.Insert(&template); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: template})
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: template})
		}
		log.E(l, "Failed to create template.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	log.D(l, "Created template successfully.", func(cm log.CM) {
		cm.Write(zap.Object("template", template))
	})
	return c.JSON(http.StatusCreated, template)
}

// GetTemplateHandler is the method called when a get to /apps/:id/templates/:tid is called
func (a *Application) GetTemplateHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "templateHandler"),
		zap.String("operation", "getTemplate"),
		zap.String("appId", c.Param("id")),
		zap.String("templateId", c.Param("tid")),
	)
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	tid, err := uuid.FromString(c.Param("tid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	template := &model.Template{ID: tid, AppID: id}
	if err := a.DB.Model(&template).Column("template.*", "App").Where("template.id = ?", template.ID).Select(); err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, template)
		}
		log.E(l, "Failed to retrieve template.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	return c.JSON(http.StatusOK, template)
}

// PutTemplateHandler is the method called when a put to /apps/:id/templates/:tid is called
func (a *Application) PutTemplateHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "templateHandler"),
		zap.String("operation", "putTemplate"),
		zap.String("appId", c.Param("id")),
		zap.String("templateId", c.Param("tid")),
	)
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	tid, err := uuid.FromString(c.Param("tid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	email := c.Get("user-email").(string)
	template := &model.Template{
		ID:        tid,
		AppID:     id,
		CreatedBy: email,
	}
	err = decodeAndValidate(c, template)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: template})
	}
	template.ID = tid
	template.AppID = id
	values, err := a.DB.Model(&template).Column("name").Column("locale").Column("defaults").Column("body").Returning("*").Update()
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: template})
		}
		log.E(l, "Failed to update template.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	if values.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{})
	}
	log.D(l, "Updated template successfully.", func(cm log.CM) {
		cm.Write(zap.Object("template", template))
	})
	return c.JSON(http.StatusOK, template)
}

// DeleteTemplateHandler is the method called when a delete to /apps/:id/templates/:tid is called
func (a *Application) DeleteTemplateHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "templateHandler"),
		zap.String("operation", "deleteTemplate"),
		zap.String("appId", c.Param("id")),
		zap.String("templateId", c.Param("tid")),
	)
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	tid, err := uuid.FromString(c.Param("tid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	template := &model.Template{}
	res, err := a.DB.Model(&template).Where("id = ? AND app_id = ?", tid, id).Delete()
	if err != nil {
		log.E(l, "Failed to delete template.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	if res.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{})
	}
	log.D(l, "Deleted template successfully.", func(cm log.CM) {
		cm.Write(zap.Object("template", template))
	})
	return c.JSON(http.StatusNoContent, "")
}
