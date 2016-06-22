package cmd

import (
	"fmt"
	"path"
	"runtime"

	"gopkg.in/gorp.v1"

  "github.com/golang/glog"
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/topfreegames/goose/lib/goose"
)

var migrationVersion int64

func getDatabase() (*gorp.DbMap, error) {
	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	dbName := viper.GetString("postgres.dbname")
	password := viper.GetString("postgres.password")
	port := viper.GetInt("postgres.port")
	sslMode := viper.GetString("postgres.sslMode")

	glog.Infof(
		"\nConnecting to %s:%d as %s using sslMode=%s to db %s...\n\n",
		host, port, user, sslMode, dbName,
	)
	db, err := models.GetDB(host, user, port, sslMode, dbName, password)
	return db.(*gorp.DbMap), err
}

func getGooseConf(migrationsDir string) *goose.DBConf {
	if migrationsDir == "" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			panic("Could not find configuration file...")
		}
		migrationsDir = path.Join(path.Dir(filename), "../db/migrations/")
	}

	return &goose.DBConf{
		MigrationsDir: migrationsDir,
		Env:           "production",
		Driver: goose.DBDriver{
			Name:    "postgres",
			OpenStr: "",
			Import:  "github.com/jinzhu/gorm/dialects/postgres",
			Dialect: &goose.PostgresDialect{},
		},
	}
}

// MigrationError identified rigrations running error
type MigrationError struct {
	Message string
}

func (err *MigrationError) Error() string {
	return fmt.Sprintf("Could not run migrations: %s", err.Message)
}

func runMigrations(migrationsDir string, migrationVersion int64) error {
	conf := getGooseConf(migrationsDir)
	db, err := getDatabase()
	if err != nil {
		return &MigrationError{fmt.Sprintf("could not connect to database: %s", err.Error())}
	}

	targetVersion := migrationVersion
	if targetVersion == -1 {
		// Get the latest possible migration
		latest, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
		if err != nil {
			return &MigrationError{fmt.Sprintf("could not get migrations at %s: %s", conf.MigrationsDir, err.Error())}
		}
		targetVersion = latest
	}

	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(conf, conf.MigrationsDir, targetVersion, db.Db)
	if err != nil {
		return &MigrationError{fmt.Sprintf("could not run migrations to %d: %s", targetVersion, err.Error())}
	}
	glog.Info("Migrated database successfully to version %d.\n", targetVersion)
	return nil
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrates the database up or down",
	Long:  `Migrate the database specified in the configuration file to the given version (or latest if none provided)`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := runMigrations("", migrationVersion)
		if err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().Int64VarP(&migrationVersion, "target", "t", -1, "Version to run up to or down to")
}
