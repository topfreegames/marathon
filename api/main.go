/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
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

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"

	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/worker"
)

// Application is the api main struct
type Application struct {
	Debug          bool
	API            *echo.Echo
	Logger         zap.Logger
	Port           int
	Host           string
	DB             interfaces.DB
	PushDB         interfaces.DB
	ConfigPath     string
	Config         *viper.Viper
	NewRelic       newrelic.Application
	Worker         *worker.Worker
	S3Client       s3iface.S3API
	SendgridClient *extensions.SendgridClient
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
		log.P(l, "cannot configure application", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
	return a
}

func (a *Application) loadConfiguration() error {
	a.Config.SetConfigFile(a.ConfigPath)
	a.Config.SetEnvPrefix("marathon")
	a.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	a.Config.AutomaticEnv()

	if err := a.Config.ReadInConfig(); err == nil {
		log.I(a.Logger, "Loaded config file.", func(cm log.CM) {
			cm.Write(zap.String("configFile", a.Config.ConfigFileUsed()))
		})
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

	err = a.configurePushDatabase()
	if err != nil {
		return err
	}

	a.configureApplication()
	a.configureWorker()
	a.configureSentry()
	a.configureSendgrid()

	err = a.configureNewRelic()
	if err != nil {
		return err
	}

	err = a.configureS3Client()
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) configureS3Client() error {
	s3, err := extensions.NewS3(a.Config, a.Logger)
	if err != nil {
		log.E(a.Logger, "Failed to initialize S3.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}
	a.S3Client = s3
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
	a.Worker = worker.NewWorker(a.Debug, a.Logger, a.ConfigPath)
}

func (a *Application) configureApplication() {
	e := echo.New()
	_, w, _ := os.Pipe()
	e.Logger.SetOutput(w)
	e.Pre(middleware.RemoveTrailingSlash())

	// Base Routes
	e.GET("/healthcheck", a.HealthcheckHandler)

	uploadGroup := e.Group("/uploadurl")
	// AuthMiddleware MUST be the first middleware
	uploadGroup.Use(NewUploadAuthMiddleware(a).Serve)
	uploadGroup.Use(NewLoggerMiddleware(a.Logger).Serve)
	uploadGroup.Use(NewRecoveryMiddleware(a.OnErrorHandler).Serve)
	uploadGroup.Use(NewVersionMiddleware().Serve)
	uploadGroup.Use(NewSentryMiddleware(a).Serve)
	uploadGroup.Use(NewNewRelicMiddleware(a, a.Logger).Serve)

	// Upload Routes
	uploadGroup.GET("", a.GetUploadURL)

	appGroup := e.Group("/apps")
	// AuthMiddleware MUST be the first middleware
	appGroup.Use(NewAppAuthMiddleware(a).Serve)
	appGroup.Use(NewLoggerMiddleware(a.Logger).Serve)
	appGroup.Use(NewRecoveryMiddleware(a.OnErrorHandler).Serve)
	appGroup.Use(NewVersionMiddleware().Serve)
	appGroup.Use(NewSentryMiddleware(a).Serve)
	appGroup.Use(NewNewRelicMiddleware(a, a.Logger).Serve)

	// Apps Routes
	appGroup.POST("", a.PostAppHandler)
	appGroup.GET("", a.ListAppsHandler)
	appGroup.GET("/:aid", a.GetAppHandler)
	appGroup.PUT("/:aid", a.PutAppHandler)
	appGroup.DELETE("/:aid", a.DeleteAppHandler)

	// Templates Routes
	appGroup.POST("/:aid/templates", a.PostTemplateHandler)
	appGroup.GET("/:aid/templates", a.ListTemplatesHandler)
	appGroup.GET("/:aid/templates/:tid", a.GetTemplateHandler)
	appGroup.PUT("/:aid/templates/:tid", a.PutTemplateHandler)
	appGroup.DELETE("/:aid/templates/:tid", a.DeleteTemplateHandler)

	// Jobs Routes
	appGroup.POST("/:aid/jobs", a.PostJobHandler)
	appGroup.GET("/:aid/jobs", a.ListJobsHandler)
	appGroup.GET("/:aid/jobs/:jid", a.GetJobHandler)
	appGroup.PUT("/:aid/jobs/:jid/pause", a.PauseJobHandler)
	appGroup.PUT("/:aid/jobs/:jid/stop", a.StopJobHandler)
	appGroup.PUT("/:aid/jobs/:jid/resume", a.ResumeJobHandler)

	userGroup := e.Group("/users")
	// AuthMiddleware MUST be the first middleware
	userGroup.Use(NewUserAuthMiddleware(a).Serve)
	userGroup.Use(NewLoggerMiddleware(a.Logger).Serve)
	userGroup.Use(NewRecoveryMiddleware(a.OnErrorHandler).Serve)
	userGroup.Use(NewVersionMiddleware().Serve)
	userGroup.Use(NewSentryMiddleware(a).Serve)
	userGroup.Use(NewNewRelicMiddleware(a, a.Logger).Serve)

	// User Routes
	userGroup.GET("", a.ListUsersHandler)
	userGroup.POST("", a.CreateUserHandler)
	userGroup.GET("/:uid", a.GetUserHandler)
	userGroup.PUT("/:uid", a.UpdateUserHandler)
	userGroup.DELETE("/:uid", a.DeleteUserHandler)

	a.API = e
}

func (a *Application) configureDatabase() error {
	PgClient, err := extensions.NewPGClient("db", a.Config, a.Logger)
	a.DB = PgClient.DB
	if err == nil {
		log.I(a.Logger, "successfully connected to the marathon database")
	}
	return err
}

func (a *Application) configurePushDatabase() error {
	PgClient, err := extensions.NewPGClient("push.db", a.Config, a.Logger)
	a.PushDB = PgClient.DB
	if err == nil {
		log.I(a.Logger, "successfully connected to the push database")
	}
	return err
}

func (a *Application) configureSendgrid() {
	l := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureSendgrid"),
	)
	apiKey := a.Config.GetString("sendgrid.key")
	if apiKey != "" {
		a.SendgridClient = extensions.NewSendgridClient(a.Config, a.Logger, apiKey)
		log.I(l, "Configured sendgrid successfully.")
	}
}

func (a *Application) configureSentry() {
	l := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := a.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	log.I(l, "Configured sentry successfully.")
}

func (a *Application) configureNewRelic() error {
	newRelicKey := a.Config.GetString("newrelic.key")

	l := a.Logger.With(
		zap.String("source", "main"),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig("Marathon", newRelicKey)
	if newRelicKey == "" {
		log.I(l, "New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		log.E(a.Logger, "Failed to initialize New Relic.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	a.NewRelic = nr
	log.I(l, "Initialized New Relic successfully.")

	return nil
}

// Start starts the api
func (a *Application) Start() {
	err := a.API.Start(fmt.Sprintf("%s:%d", a.Host, a.Port))
	if err != nil {
		log.P(a.Logger, "Cannot start api.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
}
