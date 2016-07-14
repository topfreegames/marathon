package models

import (
	"database/sql"
	"fmt"
	"os"

	"git.topfreegames.com/topfreegames/marathon/util"
	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	"github.com/uber-go/zap"
	"gopkg.in/gorp.v1"
)

func getLogLevel() zap.Level {
	var level = zap.WarnLevel
	var environment = os.Getenv("ENV")
	if environment == "test" {
		level = zap.FatalLevel
	}
	return level
}

// Logger is the models logger
var Logger = zap.NewJSON(getLogLevel())

// DB is the contract for all the operations we use from either a connection or transaction
// This is required for automatic transactions
type DB interface {
	Get(interface{}, ...interface{}) (interface{}, error)
	Select(interface{}, string, ...interface{}) ([]interface{}, error)
	SelectOne(interface{}, string, ...interface{}) error
	SelectInt(string, ...interface{}) (int64, error)
	Insert(...interface{}) error
	Update(...interface{}) (int64, error)
	Delete(...interface{}) (int64, error)
	AddTableWithName(interface{}, string) *gorp.TableMap
	CreateTables() error
	Exec(string, ...interface{}) (sql.Result, error)
}

var _db DB

// GetTestDB returns a connection to the test database
func GetTestDB() (DB, error) {
	return GetDB("localhost", "marathon_test", 5432, "disable", "marathon_test", "")
}

// GetDB returns a DbMap connection to the database specified in the arguments
func GetDB(host string, user string, port int, sslmode string, dbName string, password string) (DB, error) {
	if _db == nil {
		var err error
		_db, err = InitDb(host, user, port, sslmode, dbName, password)
		if err != nil {
			_db = nil
			return nil, err
		}
	}

	return _db, nil
}

// InitDb initializes a connection to the database
func InitDb(host string, user string, port int, sslmode string, dbName string, password string) (DB, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s port=%d sslmode=%s dbname=%s",
		host, user, port, sslmode, dbName,
	)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.TypeConverter = util.TypeConverter{}

	dbmap.AddTableWithName(Organization{}, "organizations").SetKeys(false, "ID")
	dbmap.AddTableWithName(App{}, "apps").SetKeys(false, "ID")
	dbmap.AddTableWithName(Notifier{}, "notifiers").SetKeys(false, "ID")
	dbmap.AddTableWithName(Template{}, "templates").SetKeys(false, "ID")

	return dbmap, nil
}
