package api

import (
	"fmt"
	"os"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	"gopkg.in/gorp.v1"

	"github.com/getsentry/raven-go"
	_ "github.com/jinzhu/gorm/dialects/postgres" // This is required to use postgres with gorm
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Application is a struct that represents a Marathon API Applicationlication
type Application struct {
	Debug          bool
	Port           int
	Host           string
	ConfigPath     string
	Application    *iris.Framework
	Db             models.DB
	Config         *viper.Viper
	Logger         zap.Logger
	ReadBufferSize int
}

// GetApplication returns a new Marathon API Applicationlication
func GetApplication(host string, port int, configPath string, debug bool, logger zap.Logger) *Application {
	application := &Application{
		Host:           host,
		Port:           port,
		ConfigPath:     configPath,
		Config:         viper.New(),
		Debug:          debug,
		Logger:         logger,
		ReadBufferSize: 30000,
	}

	application.Configure()
	return application
}

func getLogLevel() zap.Level {
	var level = zap.WarnLevel
	var environment = os.Getenv("ENV")
	if environment == "test" {
		level = zap.FatalLevel
	}
	return level
}

// Configure instantiates the required dependencies for Marathon Api Applicationlication
func (application *Application) Configure() {
	application.setConfigurationDefaults()
	application.loadConfiguration()
	application.configureSentry()
	application.connectDatabase()
	application.configureApplicationlication()
}

func (application *Application) setConfigurationDefaults() {
	l := application.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "setConfigurationDefaults"),
	)
	application.Config.SetDefault("healthcheck.workingText", "WORKING")
	application.Config.SetDefault("postgres.host", "localhost")
	application.Config.SetDefault("postgres.user", "marathon")
	application.Config.SetDefault("postgres.dbName", "marathon")
	application.Config.SetDefault("postgres.port", 5432)
	application.Config.SetDefault("postgres.sslMode", "disable")
	l.Debug("Configuration defaults set.")
}

func (application *Application) loadConfiguration() {
	l := application.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "loadConfiguration"),
		zap.String("configPath", application.ConfigPath),
	)

	application.Config.SetConfigFile(application.ConfigPath)
	application.Config.SetEnvPrefix("marathon")
	application.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	application.Config.AutomaticEnv()

	if err := application.Config.ReadInConfig(); err == nil {
		l.Info("Using config file", zap.String("file", application.Config.ConfigFileUsed()))
	} else {
		l.Panic(fmt.Sprintf("Could not load configuration file", zap.String("path", application.ConfigPath)))
	}
}

func (application *Application) configureSentry() {
	l := application.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := application.Config.GetString("sentry.url")
	l.Info("Configuring sentry", zap.String("url", sentryURL))
	raven.SetDSN(sentryURL)
	raven.SetRelease(VERSION)
}

func (application *Application) connectDatabase() {
	host := application.Config.GetString("postgres.host")
	user := application.Config.GetString("postgres.user")
	dbName := application.Config.GetString("postgres.dbname")
	password := application.Config.GetString("postgres.password")
	port := application.Config.GetInt("postgres.port")
	sslMode := application.Config.GetString("postgres.sslMode")

	l := application.Logger.With(
		zap.String("source", "application"),
		zap.String("operation", "connectDatabase"),
		zap.String("host", host),
		zap.String("user", user),
		zap.String("dbName", dbName),
		zap.Int("port", port),
		zap.String("sslMode", sslMode),
	)

	l.Debug("Connecting to database...")
	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		l.Panic(
			"Could not connect to postgres...",
			zap.String("error", err.Error()),
		)
	}

	_, err = db.SelectInt("select 1")
	if err != nil {
		l.Panic(
			"Could not connect to postgres...",
			zap.String("error", err.Error()),
		)
	}

	l.Info("Connected to database successfully.")

	application.Db = db
}

func (application *Application) configureApplicationlication() {
	application.Application = iris.New()
	a := application.Application

	a.Use(&TransactionMiddleware{Application: application})

	a.Get("/healthcheck", HealthCheckHandler(application))

	// Routes

	// Create an app
	a.Post("/apps", CreateAppHandler(application))

	// Send a push notification by filters
	a.Post("/apps/:appName/users/notifications", SendNotificationHandler(application))

	// a.Put("/apps/:app_name/users/notification", CreateAppHandler(application))
	// a.Post("/notifiers", controllers.CreateNotifierHandler(application))
	// a.Post("/organizations", controllers.CreateOrganizationHandler(application))
	// a.Post("/templates", controllers.CreateTemplatesHandler(application))
	// a.Post("/userTokens", controllers.CreateUserTokensHandler(application))
	// a.Post("/userTokensTable", controllers.CreateUserTokensTableHandler(application))
}

func (application *Application) finalizeApplication() {
	application.Db.(*gorp.DbMap).Db.Close()
}

// Start starts listening for web requests at specified host and port
func (application *Application) Start() {
	defer application.finalizeApplication()
	application.Application.Listen(fmt.Sprintf("%s:%d", application.Host, application.Port))
}
