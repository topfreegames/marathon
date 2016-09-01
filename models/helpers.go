package models

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	"git.topfreegames.com/topfreegames/marathon/util"
	_ "github.com/lib/pq" //This is required to use postgres with database/sql
	"github.com/uber-go/zap"
	"gopkg.in/gorp.v1"
)

// QueryLogger logs a query with zap
type QueryLogger interface {
	Printf(format string, v ...interface{})
}
type queryLogger struct {
	l zap.Logger
}

func (ql queryLogger) Printf(format string, v ...interface{}) {
	r := regexp.MustCompile("[ ]*\n")
	logStr := fmt.Sprintf(format, v...)
	logStr = r.ReplaceAllString(logStr, " ")
	ql.l.Info(logStr)
}

// DB is a gorp.DbMap with a Logger
type DB struct {
	gorp.DbMap
	Logger zap.Logger
}

var _db *DB

// GetTestDB returns a connection to the test database
func GetTestDB(l zap.Logger) (*DB, error) {
	return GetDB(l, "localhost", "marathon", 9910, "disable", "marathon", "")
}

// GetDB returns a DbMap connection to the database specified in the arguments
func GetDB(l zap.Logger, host string, user string, port int, sslmode string, dbName string, password string) (*DB, error) {
	if _db == nil {
		var err error
		_db, err = InitDb(l, host, user, port, sslmode, dbName, password)
		if err != nil {
			_db = nil
			return nil, err
		}
	}

	return _db, nil
}

// InitDb initializes a connection to the database
func InitDb(l zap.Logger, host string, user string, port int, sslmode string, dbName string, password string) (*DB, error) {
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

	dbmap := &DB{gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}, l}
	dbmap.TypeConverter = util.TypeConverter{}

	dbmap.AddTableWithName(App{}, "apps").SetKeys(false, "ID")
	dbmap.AddTableWithName(Notifier{}, "notifiers").SetKeys(false, "ID")
	dbmap.AddTableWithName(Template{}, "templates").SetKeys(false, "ID")

	// TODO: Use config
	if os.Getenv("LOG_QUERIES") == "true" {
		ql := queryLogger{zap.New(
			zap.NewJSONEncoder(),
			zap.DebugLevel,
			zap.AddCaller(),
		)}
		dbmap.TraceOn("", ql)
		// dbmap.TraceOn("[postgres]", log.New(os.Stdout, "marathon:", log.Lmicroseconds))
	}

	return dbmap, nil
}
