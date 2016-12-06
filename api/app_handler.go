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

// ListAppsHandler is the method called when a get to /apps is called
func (a *Application) ListAppsHandler(c echo.Context) error {
	apps := []model.App{}
	if err := a.DB.Model(&apps).Select(); err != nil {
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	return c.JSON(http.StatusOK, apps)
}

// PostAppHandler is the method called when a post to /apps is called
func (a *Application) PostAppHandler(c echo.Context) error {
	app := &model.App{
		ID: uuid.NewV4(),
	}
	email := c.Get("user-email").(string)
	app.CreatedBy = email
	err := decodeAndValidate(c, app)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: app})
	}
	if err = a.DB.Insert(&app); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: app})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	return c.JSON(http.StatusCreated, app)
}

// GetAppHandler is the method called when a get to /apps/:id is called
func (a *Application) GetAppHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app := &model.App{ID: id}
	if err := a.DB.Select(&app); err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, app)
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	return c.JSON(http.StatusOK, app)
}

// PutAppHandler is the method called when a put to /apps/:id is called
func (a *Application) PutAppHandler(c echo.Context) error {
	app := &model.App{}
	email := c.Get("user-email").(string)
	app.CreatedBy = email
	err := decodeAndValidate(c, app)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: app})
	}
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app.ID = id
	if err = a.DB.Update(&app); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: app})
		}
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	return c.JSON(http.StatusOK, app)
}

// DeleteAppHandler is the method called when a delete to /apps/:id is called
func (a *Application) DeleteAppHandler(c echo.Context) error {
	id, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app := &model.App{ID: id}
	if err := a.DB.Delete(&app); err != nil {
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	return c.JSON(http.StatusNoContent, "")
}
