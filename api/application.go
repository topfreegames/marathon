package api

import (
	"fmt"
	"os"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	"gopkg.in/gorp.v1"

	_ "github.com/jinzhu/gorm/dialects/postgres" // This is required to use postgres with gorm
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Application is a struct that represents a Marathon API Applicationlication
type Application struct {
	Debug       bool
	Port        int
	Host        string
	ConfigPath  string
	Application *iris.Framework
	Db          models.DB
	Config      *viper.Viper
	Logger      zap.Logger
}

// GetApplication returns a new Marathon API Applicationlication
func GetApplication(host string, port int, configPath string, debug bool) *Application {
	application := &Application{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
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
	application.Logger = zap.NewJSON(getLogLevel())
	application.setConfigurationDefaults()
	application.loadConfiguration()
	application.connectDatabase()
	application.configureApplicationlication()
}

func (application *Application) setConfigurationDefaults() {
	application.Config.SetDefault("healthcheck.workingText", "WORKING")
	application.Config.SetDefault("postgres.host", "localhost")
	application.Config.SetDefault("postgres.user", "marathon")
	application.Config.SetDefault("postgres.dbName", "marathon")
	application.Config.SetDefault("postgres.port", 5432)
	application.Config.SetDefault("postgres.sslMode", "disable")
}

func (application *Application) loadConfiguration() {
	application.Config.SetConfigFile(application.ConfigPath)
	application.Config.SetEnvPrefix("marathon")
	application.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	application.Config.AutomaticEnv()

	if err := application.Config.ReadInConfig(); err == nil {
		application.Logger.Info("Using config file", zap.String("file", application.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", application.ConfigPath))
	}
}

func (application *Application) connectDatabase() {
	host := application.Config.GetString("postgres.host")
	user := application.Config.GetString("postgres.user")
	dbName := application.Config.GetString("postgres.dbname")
	password := application.Config.GetString("postgres.password")
	port := application.Config.GetInt("postgres.port")
	sslMode := application.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		application.Logger.Error(
			"Could not connect to postgres...",
			zap.String("host", host),
			zap.Int("port", port),
			zap.String("user", user),
			zap.String("dbName", dbName),
			zap.Error(err),
		)
		panic(err)
	}
	application.Db = db
}

func (application *Application) configureApplicationlication() {
	application.Application = iris.New()
	a := application.Application

	a.Use(&TransactionMiddleware{Application: application})

	a.Get("/healthcheck", HealthCheckHandler(application))

	// Application Routes
	// a.Post("/applications", CreateGameHandler(application))
}

func (application *Application) finalizeApplication() {
	application.Db.(*gorp.DbMap).Db.Close()
}

// Start starts listening for web requests at specified host and port
func (application *Application) Start() {
	defer application.finalizeApplication()
	application.Application.Listen(fmt.Sprintf("%s:%d", application.Host, application.Port))
}
