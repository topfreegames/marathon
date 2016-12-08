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
	"math"
	"sync"

	"gopkg.in/pg.v5"

	"github.com/jrallison/go-workers"
	"github.com/minio/minio-go"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// CreateBatchesWorker is the CreateBatchesWorker struct
type CreateBatchesWorker struct {
	Logger                    zap.Logger
	MarathonDB                *extensions.PGClient
	PushDB                    *extensions.PGClient
	Workers                   *Worker
	Config                    *viper.Viper
	BatchSize                 int
	DBPageSize                int
	S3Client                  *minio.Client
	PageProcessingConcurrency int
}

// User is the struct that will keep users before sending them to send batches worker
type User struct {
	UserID string `json:"user_id" sql:"user_id"`
	Token  string `json:"token" sql:"token"`
	Locale string `json:"locale" sql:"locale"`
	Tz     string `json:"tz" sql:"tz"`
}

// NewCreateBatchesWorker gets a new CreateBatchesWorker
func NewCreateBatchesWorker(config *viper.Viper, logger zap.Logger, workers *Worker) *CreateBatchesWorker {
	b := &CreateBatchesWorker{
		Config:  config,
		Logger:  logger,
		Workers: workers,
	}
	b.configure()
	return b
}

func (b *CreateBatchesWorker) configure() {
	b.loadConfigurationDefaults()
	b.loadConfiguration()
	b.configureDatabases()
	b.configureS3Client()
}

func (b *CreateBatchesWorker) configureS3Client() {
	s3AccessKeyID := b.Config.GetString("s3.accessKey")
	s3SecretAccessKey := b.Config.GetString("s3.secretAccessKey")
	ssl := true
	s3Client, err := minio.New("s3.amazonaws.com", s3AccessKeyID, s3SecretAccessKey, ssl)
	checkErr(err)
	b.S3Client = s3Client
}

func (b *CreateBatchesWorker) loadConfigurationDefaults() {
	b.Config.SetDefault("workers.createBatches.batchSize", 1000)
	b.Config.SetDefault("workers.createBatches.dbPageSize", 1000)
	b.Config.SetDefault("workers.createBatches.pageProcessingConcurrency", 1)
}

func (b *CreateBatchesWorker) loadConfiguration() {
	b.BatchSize = b.Config.GetInt("workers.createBatches.batchSize")
	b.DBPageSize = b.Config.GetInt("workers.createBatches.dbPageSize")
	b.PageProcessingConcurrency = b.Config.GetInt("workers.createBatches.pageProcessingConcurrency")
}

func (b *CreateBatchesWorker) configurePushDatabase() {
	var err error
	b.PushDB, err = extensions.NewPGClient("push.db", b.Config, b.Logger)
	checkErr(err)
}

func (b *CreateBatchesWorker) configureMarathonDatabase() {
	var err error
	b.MarathonDB, err = extensions.NewPGClient("db", b.Config, b.Logger)
	checkErr(err)
}

func (b *CreateBatchesWorker) configureDatabases() {
	b.configureMarathonDatabase()
	b.configurePushDatabase()
}

func (b *CreateBatchesWorker) readCSVFromS3(csvPath string) []string {
	bucket := b.Config.GetString("s3.bucket")
	folder := b.Config.GetString("s3.folder")
	csvFile, err := b.S3Client.GetObject(bucket, fmt.Sprintf("/%s/%s", folder, csvPath))
	checkErr(err)
	r := csv.NewReader(csvFile)
	lines, err := r.ReadAll()
	checkErr(err)
	res := []string{}
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
	_, err := b.PushDB.DB.Query(&users, fmt.Sprintf("SELECT user_id, token, locale, tz FROM %s_%s WHERE user_id IN (?)", appName, service), pg.In(*userIds))
	checkErr(err)
	return users
}

func (b *CreateBatchesWorker) sendBatch(batches *map[string][]User) {
	l := b.Logger
	for tz, users := range *batches {
		l.Info("sending batch of users to process batches worker", zap.Int("numUsers", len(users)), zap.String("tz", tz))

	}
}

func (b *CreateBatchesWorker) processBatch(c <-chan *[]string, job *model.Job, wg *sync.WaitGroup) {
	l := b.Logger
	bucketsByTZ := map[string][]User{}
	for userIds := range c {
		usersFromBatch := b.getCSVUserBatchFromPG(userIds, job.App.Name, job.Service)
		l.Info("got users from db", zap.Int("usersInBatch", len(usersFromBatch)))
		for _, user := range usersFromBatch {
			userTz := user.Tz
			if len(userTz) == 0 {
				userTz = "-0500"
			}
			if res, ok := bucketsByTZ[userTz]; ok {
				bucketsByTZ[userTz] = append(res, user)
			} else {
				bucketsByTZ[userTz] = []User{user}
			}
		}
		for tz, users := range bucketsByTZ {
			l.Debug("batch of users for tz", zap.Int("numUsers", len(users)), zap.String("tz", tz))
		}
		//TODO for now I'll just ignore timezones and send all pushes
		b.sendBatch(&bucketsByTZ)
		wg.Done()
	}
}

func (b *CreateBatchesWorker) createBatchesUsingCSV(job *model.Job) error {
	l := b.Logger
	userIds := b.readCSVFromS3(job.CSVPath)
	numPushes := len(userIds)
	pages := int(math.Ceil(float64(numPushes) / float64(b.DBPageSize)))
	l.Info("grabing pages from pg", zap.Int("pagesToComplete", pages))
	var wg sync.WaitGroup
	pgCH := make(chan *[]string, pages)
	wg.Add(pages)
	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processBatch(pgCH, job, &wg)
	}
	for i := 0; i < pages; i++ {
		userBatch := b.getPage(i, &userIds)
		pgCH <- &userBatch
	}
	wg.Wait()
	close(pgCH)
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
	checkErr(err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(err)
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(err)
	if len(job.CSVPath) > 0 {
		err := b.createBatchesUsingCSV(job)
		checkErr(err)
	} else {
		// Find the ids based on filters
	}
}
