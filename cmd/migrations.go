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

package cmd

import (
	"database/sql"
	"fmt"
	"os"

	// pg driver
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// migrationsCmd represents the migrate command
var migrationsCmd = &cobra.Command{
	Use:   "migrations",
	Short: "use this command to work with migrations",
	Long:  "use this command to work with migrations",
}

func executeMigrationCmd(cmd string) {
	ll := zap.InfoLevel
	if debug {
		ll = zap.DebugLevel
	}

	l := zap.New(
		zap.NewJSONEncoder(),
		ll,
	)

	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		l.Panic("error loading config file", zap.Error(err))
	}

	host := viper.GetString("db.host")
	port := viper.GetInt("db.port")
	database := viper.GetString("db.database")
	user := viper.GetString("db.user")
	pass := viper.GetString("db.pass")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, pass, host, port, database)

	logger := l.With(zap.String("dbUrl", dbURL))

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Panic("error migrating database", zap.Error(err))
	}
	logger.Info("migrating database...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Panic("error migrating database", zap.Error(err))
	}

	//TODO add command to support down and up
	if err := goose.Run(cmd, db, "./migrations"); err != nil {
		logger.Fatal("error migrating database", zap.Error(err))
	}

	logger.Info("successfully migrated tables!")
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "use this command to run all migrations",
	Long:  "use this command to run all migrations",
	Run: func(cmd *cobra.Command, args []string) {
		executeMigrationCmd("up")
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "use this command to rollback a single migration",
	Long:  "use this command to rollback a single migration",
	Run: func(cmd *cobra.Command, args []string) {
		executeMigrationCmd("down")
	},
}
var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "use this command to rollback the most recently applied migration, then run it again",
	Long:  "use this command to rollback the most recently applied migration, then run it again",
	Run: func(cmd *cobra.Command, args []string) {
		executeMigrationCmd("redo")
	},
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "use this command to create a migration",
	Long:  "use this command to create a migration",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			os.Exit(1)
		}
		if err := goose.Create(nil, "./migrations", args[0], "sql"); err != nil {
			panic(err)
		}
	},
}

func init() {
	migrationsCmd.AddCommand(createCmd)
	migrationsCmd.AddCommand(upCmd)
	migrationsCmd.AddCommand(downCmd)
	migrationsCmd.AddCommand(redoCmd)
	RootCmd.AddCommand(migrationsCmd)
}
