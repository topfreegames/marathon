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

package worker

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"gopkg.in/pg.v5"

	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/model"
)

// CreateBatchesWorker is the CreateBatchesWorker struct
type CreateBatchesWorker struct {
	MarathonDB *pg.DB
	PushDB     *pg.DB
	Config     *viper.Viper
	BatchSize  int
	DBPageSize int
}

// User is the struct that will keep users before sending them to send batches worker
type User struct {
	UserID string `json:"user_id" sql:"user_id"`
	Token  string `json:"token" sql:"token"`
	Locale string `json:"locale" sql:"locale"`
	Tz     string `json:"tz" sql:"tz"`
}

// GetCreateBatchesWorker gets a new CreateBatchesWorker
func GetCreateBatchesWorker(config *viper.Viper) *CreateBatchesWorker {
	b := &CreateBatchesWorker{
		Config: config,
	}
	b.configure()
	return b
}

func (b *CreateBatchesWorker) configure() {
	b.loadConfigurationDefaults()
	b.loadConfiguration()
	b.configureDatabases()
}

func (b *CreateBatchesWorker) loadConfigurationDefaults() {
	b.Config.SetDefault("workers.createBatches.batchSize", 1000)
	b.Config.SetDefault("workers.createBatches.dbPageSize", 1000)
}

func (b *CreateBatchesWorker) loadConfiguration() {
	b.BatchSize = b.Config.GetInt("workers.createBatches.batchSize")
	b.DBPageSize = b.Config.GetInt("workers.createBatches.dbPageSize")
}

func (b *CreateBatchesWorker) configurePushDatabase() {
	pushDBUser := b.Config.GetString("push.db.user")
	pushDBPass := b.Config.GetString("push.db.pass")
	pushDBHost := b.Config.GetString("push.db.host")
	pushDBDatabase := b.Config.GetString("push.db.database")
	pushDBPort := b.Config.GetInt("push.db.port")
	pushDBPoolSize := b.Config.GetInt("push.db.poolSize")
	pushDBMaxRetries := b.Config.GetInt("push.db.maxRetries")
	pushDB := pg.Connect(&pg.Options{
		Addr:       fmt.Sprintf("%s:%d", pushDBHost, pushDBPort),
		User:       pushDBUser,
		Password:   pushDBPass,
		Database:   pushDBDatabase,
		PoolSize:   pushDBPoolSize,
		MaxRetries: pushDBMaxRetries,
	})
	b.PushDB = pushDB
}

func (b *CreateBatchesWorker) configureMarathonDatabase() {
	host := b.Config.GetString("db.host")
	user := b.Config.GetString("db.user")
	pass := b.Config.GetString("db.pass")
	database := b.Config.GetString("db.database")
	port := b.Config.GetInt("db.port")
	poolSize := b.Config.GetInt("db.poolSize")
	maxRetries := b.Config.GetInt("db.maxRetries")
	marathonDB := pg.Connect(&pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	})
	b.MarathonDB = marathonDB
}

func (b *CreateBatchesWorker) configureDatabases() {
	b.configureMarathonDatabase()
	b.configurePushDatabase()
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (b *CreateBatchesWorker) readRemoteCSV(csvURL string) []string {
	res := []string{}
	resp, err := http.Get(csvURL)
	checkErr(err)
	defer resp.Body.Close()
	r := csv.NewReader(resp.Body)
	lines, err := r.ReadAll()
	checkErr(err)
	for i, line := range lines {
		if i == 0 {
			continue
		}
		res = append(res, line[0])
	}
	return res
}

func (b *CreateBatchesWorker) getCSVUserBatchFromPG(userIds *[]string, appName, service string) []User {
	var users []User
	fmt.Printf("Getting users from push %s", *userIds)
	_, err := b.PushDB.Query(&users, fmt.Sprintf("SELECT user_id, token, locale, tz FROM %s_%s WHERE user_id IN (?)", appName, service), pg.In(*userIds))
	if err != nil {
		panic(err)
	}
	return users
}

func (b *CreateBatchesWorker) createBatchesUsingCSV(csvURL, appName, service string) error {
	l := workers.Logger
	userIds := b.readRemoteCSV(csvURL)
	numPushes := len(userIds)
	pages := numPushes / (b.DBPageSize * 1.0)
	l.Printf("%d batches to complete", pages)
	for i := 0; true; i++ {
		userBatch := b.getPage(i, &userIds)
		if userBatch == nil {
			break
		}
		usersFromBatch := b.getCSVUserBatchFromPG(&userBatch, appName, service)
		l.Printf("got some users from db %s", usersFromBatch)
	}
	return nil
}

func (b *CreateBatchesWorker) getPage(page int, users *[]string) []string {
	start := page * b.DBPageSize
	end := (page + 1) * b.DBPageSize
	if start >= len(*users) {
		return nil
	}
	if end > len(*users) {
		end = len(*users)
	}
	return (*users)[start:end]
}

// Process processes the messages sent to batch worker queue
func (b *CreateBatchesWorker) Process(message *workers.Msg) {
	l := workers.Logger
	l.Printf("starting create_batches_worker with batchSize %d and dbBatchSize %d", b.BatchSize, b.DBPageSize)
	arr, err := message.Args().Array()
	if err != nil {
		checkErr(err)
	}
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(err)
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(err)
	if len(job.CsvURL) > 0 {
		err := b.createBatchesUsingCSV(job.CsvURL, job.App.Name, job.Service)
		if err != nil {
			panic(err)
		}
	} else {
		// Find the ids based on filters
	}
}
