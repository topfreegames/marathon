package api

import (
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	"git.topfreegames.com/topfreegames/marathon/util"
	"github.com/getsentry/raven-go"
	_ "github.com/jinzhu/gorm/dialects/postgres" // This is required to use postgres with gorm
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
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
	Db             *models.DB
	Config         *viper.Viper
	Logger         zap.Logger
	ReadBufferSize int
	RedisClient    *util.RedisClient
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

// Configure instantiates the required dependencies for Marathon Api Applicationlication
func (application *Application) Configure() error {
	application.setConfigurationDefaults()
	application.loadConfiguration()
	application.configureSentry()
	application.connectDatabase()

	err := application.configureApplicationlication()
	if err != nil {
		return err
	}

	return nil
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

	application.Config.SetDefault("redis.host", "localhost")
	application.Config.SetDefault("redis.port", 6379)
	application.Config.SetDefault("redis.password", "")
	application.Config.SetDefault("redis.db", 0)
	application.Config.SetDefault("redis.maxPoolSize", 20)

	application.Config.SetDefault("s3.bucket", "tfg-push-notifications")
	application.Config.SetDefault("s3.folder", "test/files")
	application.Config.SetDefault("s3.daysExpiry", 1)
	application.Config.SetDefault("s3.accessKey", "")
	application.Config.SetDefault("s3.secretAccessKey", "")

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
		l.Panic("Could not load configuration file", zap.String("path", application.ConfigPath))
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
	db, err := models.GetDB(l, host, user, port, sslMode, dbName, password)

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

func (application *Application) onErrorHandler(err error, stack []byte) {
	application.Logger.Error(
		"Panic occurred.",
		zap.String("source", "app"),
		zap.String("panicText", err.Error()),
		zap.String("stack", string(stack)),
	)
	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (application *Application) configureApplicationlication() error {
	l := application.Logger.With(
		zap.String("operation", "configureApplication"),
	)

	c := config.Iris{
		DisableBanner: true,
	}

	application.Application = iris.New(c)
	a := application.Application

	a.Use(NewLoggerMiddleware(application.Logger))
	a.Use(&RecoveryMiddleware{OnError: application.onErrorHandler})
	//a.Use(&TransactionMiddleware{App: app})
	a.Use(&VersionMiddleware{Application: application})
	a.Use(&SentryMiddleware{Application: application})

	a.OnError(iris.StatusInternalServerError, func(ctx *iris.Context) {
		application.Logger.Error(
			"Internal server error happened.",
			zap.String("error", string(ctx.Response.Body())),
			zap.String("source", "app"),
		)
		ctx.Write(fmt.Sprintf("INTERNAL SERVER ERROR: %s", ctx.Response.Body()))
	})

	a.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		application.Logger.Warn(
			"Route not found.", zap.String("url", ctx.Request.URI().String()),
			zap.String("source", "app"),
		)
		ctx.Write("Not Found")
	})

	// Routes

	// Healthcheck
	a.Get("/healthcheck", HealthCheckHandler(application))

	// Create an app
	a.Post("/apps", CreateAppHandler(application))
	a.Get("/apps", GetAppsHandler(application))

	// Send a push notification by filters
	a.Post("/notifiers/:notifierID/notifications", SendNotifierNotificationHandler(application))
	a.Get("/notifiers/:notifierID/notifications/:notificationID", GetNotifierNotificationStatusHandler(application))
	a.Get("/notifiers/:notifierID/notifications", GetNotifierNotifications(application))

	// Send push notification through csv file
	a.Get("/uploadurl", UploadHandler(application))

	redisHost := application.Config.GetString("redis.host")
	redisPort := application.Config.GetInt("redis.port")
	redisPass := application.Config.GetString("redis.password")
	redisDB := application.Config.GetInt("redis.db")
	redisMaxPoolSize := application.Config.GetInt("redis.maxPoolSize")

	rl := l.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
		zap.Int("db", redisDB),
		zap.Int("maxPoolSize", redisMaxPoolSize),
	)
	rl.Debug("Connecting to redis...")
	cli, err := util.GetRedisClient(redisHost, redisPort, redisPass, redisDB, redisMaxPoolSize, l)
	if err != nil {
		return err
	}
	application.RedisClient = cli
	rl.Info("Connected to redis successfully.")
	return nil
}

func (application *Application) finalizeApplication() {
	application.Db.Db.Close()
}

// Start starts listening for web requests at specified host and port
func (application *Application) Start() {
	defer application.finalizeApplication()
	application.Application.Listen(fmt.Sprintf("%s:%d", application.Host, application.Port))
}
