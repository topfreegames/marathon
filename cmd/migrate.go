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
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "use this command to run marathon migrations",
	Long:  "use this command to run marathon migrations",
	Run: func(cmd *cobra.Command, args []string) {
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
		dbURL := viper.GetString("database.url")

		logger := l.With(zap.String("dbUrl", dbURL))
		db, err := gorm.Open("postgres", dbURL)
		db.LogMode(true)
		if err != nil {
			logger.Panic("cannot connect to db", zap.Error(err))
		} else {
			logger.Info("successfully connected to the database")
		}
		logger.Info("migrating app, template and job tables...")
		err = createAppsTableMigration(db)
		checkErr(err)
		err = createTemplatesTableMigration(db)
		checkErr(err)
		err = createJobsTableMigration(db)
		checkErr(err)
		logger.Info("successfully migrated tables!")
	},
}

func createAppsTableMigration(db *gorm.DB) error {
	return db.CreateTable(&model.App{}).Error
}

func createTemplatesTableMigration(db *gorm.DB) error {
	err := db.CreateTable(&model.Template{}).Error
	err = db.Model(&model.Template{}).AddForeignKey("app_id", "apps(id)", "CASCADE", "CASCADE").Error
	return err
}

func createJobsTableMigration(db *gorm.DB) error {
	err := db.CreateTable(&model.Job{}).Error
	err = db.Model(&model.Job{}).AddForeignKey("app_id", "apps(id)", "CASCADE", "CASCADE").Error
	err = db.Model(&model.Job{}).AddForeignKey("template_id", "templates(id)", "CASCADE", "CASCADE").Error
	return err
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
