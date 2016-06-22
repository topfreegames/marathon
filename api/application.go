package api

import (
	"fmt"
	"strings"

	"gopkg.in/gorp.v1"

	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/golang/glog"
	_ "github.com/jinzhu/gorm/dialects/postgres" // This is required to use postgres with gorm
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/spf13/viper"
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

// Configure instantiates the required dependencies for Marathon Api Applicationlication
func (application *Application) Configure() {
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
		glog.Info("Using config file:", application.Config.ConfigFileUsed())
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
		glog.Errorf(
			"Could not connect to Postgres at %s:%d with user %s and db %s with password %s (%s)\n",
			host, port, user, dbName, password, err,
		)
		panic(err)
	}
	application.Db = db
}

func (application *Application) configureApplicationlication() {
	application.Application = iris.New()
	a := application.Application

	if application.Debug {
		a.Use(logger.New(iris.Logger))
	}
	// a.Use(recovery.New(os.Stderr))
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
