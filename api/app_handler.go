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
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// ListAppsHandler is the method called when a get to /apps is called
func (a *Application) ListAppsHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "appHandler"),
		zap.String("operation", "listApps"),
	)
	apps := []model.App{}
	err := WithSegment("db-select", c, func() error {
		return a.DB.Model(&apps).Select()
	})
	if err != nil {
		log.E(l, "Failed to list apps.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	log.D(l, "Listed apps successfully.", func(cm log.CM) {
		cm.Write(zap.Object("apps", apps))
	})
	return c.JSON(http.StatusOK, apps)
}

// PostAppHandler is the method called when a post to /apps is called
func (a *Application) PostAppHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "appHandler"),
		zap.String("operation", "postApp"),
	)
	app := &model.App{
		ID:        uuid.NewV4(),
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
	}
	email := c.Get("user-email").(string)
	app.CreatedBy = email
	err := WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, app)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: app})
	}
	err = WithSegment("db-insert", c, func() error {
		return a.DB.Insert(&app)
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: app})
		}
		log.E(l, "Failed to create app.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	log.D(l, "Created app successfully.", func(cm log.CM) {
		cm.Write(zap.Object("app", app))
	})
	return c.JSON(http.StatusCreated, app)
}

// GetAppHandler is the method called when a get to /apps/:aid is called
func (a *Application) GetAppHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "appHandler"),
		zap.String("operation", "getApp"),
		zap.String("appId", c.Param("aid")),
	)
	id, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app := &model.App{ID: id}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Select(&app)
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, map[string]string{})
		}
		log.E(l, "Failed to retrieve app.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	log.D(l, "Retrieved app successfully.", func(cm log.CM) {
		cm.Write(zap.Object("app", app))
	})
	return c.JSON(http.StatusOK, app)
}

// PutAppHandler is the method called when a put to /apps/:aid is called
func (a *Application) PutAppHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "appHandler"),
		zap.String("operation", "putApp"),
		zap.String("appId", c.Param("aid")),
	)
	app := &model.App{}
	email := c.Get("user-email").(string)
	app.CreatedBy = email
	app.UpdatedAt = time.Now().UnixNano()
	err := WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, app)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: app})
	}
	id, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app.ID = id
	err = WithSegment("db-update", c, func() error {
		_, err = a.DB.Model(&app).Column("name").Column("bundle_id").Column("updated_at").Returning("*").Update()
		return err
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: app})
		}
		log.E(l, "Failed to update app.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	log.D(l, "Updated app successfully.", func(cm log.CM) {
		cm.Write(zap.Object("app", app))
	})
	return c.JSON(http.StatusOK, app)
}

// DeleteAppHandler is the method called when a delete to /apps/:aid is called
func (a *Application) DeleteAppHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "appHandler"),
		zap.String("operation", "deleteApp"),
		zap.String("appId", c.Param("aid")),
	)
	id, err := uuid.FromString(c.Param("aid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	app := &model.App{ID: id}
	err = WithSegment("db-delete", c, func() error {
		return a.DB.Delete(&app)
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, map[string]string{})
		}
		log.E(l, "Failed to delete app.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: app})
	}
	log.D(l, "Deleted app successfully.", func(cm log.CM) {
		cm.Write(zap.Object("app", app))
	})
	return c.JSON(http.StatusNoContent, "")
}
