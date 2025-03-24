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

	"github.com/labstack/echo/v4"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

// Health is a struct to help healthcheck handler
type Health struct {
	Healthy bool `json:"healthy"`
}

func (a *Application) checkPostgres() error {
	_, err := a.DB.Exec("SELECT 1")
	return err
}

func (a *Application) checkPushDB() error {
	_, err := a.PushDB.Exec("SELECT 1")
	return err
}

// HealthcheckHandler is the method called when a get to /healthcheck is called
func (a *Application) HealthcheckHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "healthcheckHandler"),
		zap.String("operation", "healthcheck"),
	)

	err := a.checkPostgres()
	if err != nil {
		log.E(l, "Failed Postgres healthcheck.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Health{Healthy: false})
	}

	err = a.checkPushDB()
	if err != nil {
		log.E(l, "Failed PushDB healthcheck.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Health{Healthy: false})
	}

	return c.JSON(http.StatusOK, &Health{Healthy: true})
}
