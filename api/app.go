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
	"github.com/labstack/echo"
	"github.com/uber-go/zap"

	"github.com/jinzhu/gorm"
	// for gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

// App is the api main struct
type App struct {
	Debug  bool
	API    *echo.Echo
	Logger zap.Logger
	Port   int
	Host   string
	DB     *gorm.DB
}

// GetApp returns a configured api
func GetApp(host string, port int, debug bool, l zap.Logger) *App {
	app := &App{Port: port,
		Debug:  debug,
		Logger: l,
		Host:   host,
	}

	app.configure()
	return app
}

func (a *App) configure() {
	a.configureDatabase()
	a.configureApplication()
}

func (a *App) configureApplication() {
	e := echo.New()
	e.POST("/app", a.PostApp)
	e.GET("/app/:id", a.GetApp)
	e.PUT("/app/:bundleId", a.PutApp)
	e.DELETE("/app/:bundleId", a.DeleteApp)
	e.GET("/healthcheck", a.Healthcheck)
	a.API = e
}

func (a *App) configureDatabase() {
	dbURL := viper.GetString("database.url")
	logger := a.Logger.With(zap.String("dbUrl", dbURL))
	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		logger.Panic("cannot connect to db", zap.Error(err))
	} else {
		logger.Info("successfully connected to the database")
	}
	a.DB = db
}

// Start starts the api
func (a *App) Start() {
	err := a.API.Start(fmt.Sprintf("%s:%d", a.Host, a.Port))
	if err != nil {
		a.Logger.Panic("cannot start api", zap.Error(err))
	}
}
