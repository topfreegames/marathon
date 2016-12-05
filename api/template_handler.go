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
)

// ListTemplatesHandler is the method called when a get to /apps/:id/templates is called
func (a *Application) ListTemplatesHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	templates := []model.Template{}
	where := map[string]interface{}{
		"app_id": id,
	}
	if err := a.DB.Where(where).Find(&templates).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	return c.JSON(http.StatusOK, templates)
}

// PostTemplateHandler is the method called when a post to /apps/:id/templates is called
func (a *Application) PostTemplateHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	email := c.Get("user-email").(string)
	template := &model.Template{
		AppID:     id,
		CreatedBy: email,
	}
	err = decodeAndValidate(c, template)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: template})
	}
	if err = a.DB.Create(&template).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: template})
		}
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: template})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	return c.JSON(http.StatusCreated, template)
}

// GetTemplateHandler is the method called when a get to /apps/:id/templates/:tid is called
func (a *Application) GetTemplateHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	tid, err := uuid.FromString(c.Param("tid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	template := &model.Template{ID: tid, AppID: id}
	if err := a.DB.Where(template).First(&template).Error; err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, template)
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	return c.JSON(http.StatusOK, template)
}

// PutTemplateHandler is the method called when a put to /apps/:id/templates/:tid is called
func (a *Application) PutTemplateHandler(c echo.Context) error {
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
	queryTemplate := &model.Template{ID: tid, AppID: id}
	if err = a.DB.Where(queryTemplate).First(&queryTemplate).Error; err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, &Error{Reason: err.Error(), Value: template})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	template.ID = queryTemplate.ID
	template.AppID = queryTemplate.AppID
	template.CreatedBy = queryTemplate.CreatedBy
	if err = a.DB.Save(&template).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: template})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	return c.JSON(http.StatusOK, template)
}

// DeleteTemplateHandler is the method called when a delete to /apps/:id/templates/:tid is called
func (a *Application) DeleteTemplateHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	tid, err := uuid.FromString(c.Param("tid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	template := &model.Template{
		ID:    tid,
		AppID: id,
	}
	if err := a.DB.Where(template).First(&template).Error; err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, template)
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	if err := a.DB.Delete(&template).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: template})
	}
	return c.JSON(http.StatusOK, template)
}
