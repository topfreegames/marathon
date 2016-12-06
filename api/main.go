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
	"os"
	"strings"

	"gopkg.in/pg.v5"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/uber-go/zap"

	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/worker"
)

// Application is the api main struct
type Application struct {
	Debug      bool
	API        *echo.Echo
	Logger     zap.Logger
	Port       int
	Host       string
	DB         *pg.DB
	ConfigPath string
	Config     *viper.Viper
	NewRelic   newrelic.Application
	Worker     *worker.Worker
}

// GetApplication returns a configured api
func GetApplication(host string, port int, debug bool, l zap.Logger, configPath string) *Application {
	a := &Application{
		Config:     viper.New(),
		ConfigPath: configPath,
		Debug:      debug,
		Host:       host,
		Logger:     l,
		Port:       port,
	}

	err := a.configure()
	if err != nil {
		l.Panic("cannot configure application", zap.Error(err))
	}
	return a
}

func (a *Application) loadConfiguration() error {
	a.Config.SetConfigFile(a.ConfigPath)
	a.Config.SetEnvPrefix("marathon")
	a.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	a.Config.AutomaticEnv()

	if err := a.Config.ReadInConfig(); err == nil {
		a.Logger.Info("Loaded config file.", zap.String("configFile", a.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", a.ConfigPath)
	}

	return nil
}

func (a *Application) configure() error {
	err := a.loadConfiguration()
	if err != nil {
		return err
	}

	err = a.configureDatabase()
	if err != nil {
		return err
	}

	a.configureApplication()
	a.configureWorker()
	a.configureSentry()

	err = a.configureNewRelic()
	if err != nil {
		return err
	}
	return nil
}

//OnErrorHandler handles panics
func (a *Application) OnErrorHandler(err error, stack []byte) {
	a.Logger.Error(
		"Panic occurred.",
		zap.String("operation", "OnErrorHandler"),
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (a *Application) configureWorker() {
	a.Worker = worker.GetWorker(a.Debug, a.Logger)
}

func (a *Application) configureApplication() {
	e := echo.New()
	_, w, _ := os.Pipe()
	e.Logger.SetOutput(w)

	// AuthMiddleware MUST be the first middleware
	e.Use(NewAuthMiddleware(a).Serve)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(NewLoggerMiddleware(a.Logger).Serve)
	e.Use(NewRecoveryMiddleware(a.OnErrorHandler).Serve)
	e.Use(NewVersionMiddleware().Serve)
	e.Use(NewSentryMiddleware(a).Serve)
	e.Use(NewNewRelicMiddleware(a, a.Logger).Serve)

	// Base Routes
	e.GET("/healthcheck", a.HealthcheckHandler)

	// Apps Routes
	e.POST("/apps", a.PostAppHandler)
	e.GET("/apps", a.ListAppsHandler)
	e.GET("/apps/:id", a.GetAppHandler)
	e.PUT("/apps/:id", a.PutAppHandler)
	e.DELETE("/apps/:id", a.DeleteAppHandler)

	//TODO tudo que for appId nas rotas de template e jobs deveriam user :appId
	// Templates Routes
	e.POST("/apps/:id/templates", a.PostTemplateHandler)
	e.GET("/apps/:id/templates", a.ListTemplatesHandler)
	e.GET("/apps/:id/templates/:tid", a.GetTemplateHandler)
	e.PUT("/apps/:id/templates/:tid", a.PutTemplateHandler)
	e.DELETE("/apps/:id/templates/:tid", a.DeleteTemplateHandler)

	// Jobs Routes
	e.POST("/apps/:id/templates/:tid/jobs", a.PostJobHandler)
	e.GET("/apps/:id/templates/:tid/jobs", a.ListJobsHandler)
	e.GET("/apps/:id/templates/:tid/jobs/:jid", a.GetJobHandler)
	// TODO: implement these routes
	// e.PUT("/apps/:id/templates/:tid/jobs/:jid/stop", a.StopJobHandler)
	// e.DELETE("/apps/:id/templates/:tid/:jid/resume", a.ResumeJobHandler)
	a.API = e
}

func (a *Application) configureDatabase() error {
	host := a.Config.GetString("db.host")
	user := a.Config.GetString("db.user")
	pass := a.Config.GetString("db.pass")
	database := a.Config.GetString("db.database")
	port := a.Config.GetInt("db.port")
	poolSize := a.Config.GetInt("db.poolSize")
	maxRetries := a.Config.GetInt("db.maxRetries")
	logger := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureDatabase"),
	)
	db := pg.Connect(&pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	})
	logger.Info("successfully connected to the database")
	a.DB = db
	return nil
}

func (a *Application) configureSentry() {
	l := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := a.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (a *Application) configureNewRelic() error {
	newRelicKey := a.Config.GetString("newrelic.key")

	l := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig("Marathon", newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.Error("Failed to initialize New Relic.", zap.Error(err))
		return err
	}

	a.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

// Start starts the api
func (a *Application) Start() {
	err := a.API.Start(fmt.Sprintf("%s:%d", a.Host, a.Port))
	if err != nil {
		a.Logger.Panic("cannot start api", zap.Error(err))
	}
}
