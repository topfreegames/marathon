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
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
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
	"time"

	"gopkg.in/pg.v5"
	"gopkg.in/redis.v5"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
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
	S3Client                  s3iface.S3API
	PageProcessingConcurrency int
	RedisClient               *redis.Client
}

// NewCreateBatchesWorker gets a new CreateBatchesWorker
func NewCreateBatchesWorker(config *viper.Viper, logger zap.Logger, workers *Worker) *CreateBatchesWorker {
	b := &CreateBatchesWorker{
		Config:  config,
		Logger:  logger,
		Workers: workers,
	}
	b.configure()
	log.D(logger, "Configured CreateBatchesWorker successfully.")
	return b
}

func (b *CreateBatchesWorker) configure() {
	b.loadConfigurationDefaults()
	b.loadConfiguration()
	b.configureDatabases()
	b.configureS3Client()
	b.configureRedisClient()
}

func (b *CreateBatchesWorker) configureS3Client() {
	s3Client, err := extensions.NewS3(b.Config, b.Logger)
	checkErr(b.Logger, err)
	b.S3Client = s3Client
}

func (b *CreateBatchesWorker) configureRedisClient() {
	r, err := extensions.NewRedis("workers", b.Config, b.Logger)
	checkErr(b.Logger, err)
	b.RedisClient = r
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
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) configureMarathonDatabase() {
	var err error
	b.MarathonDB, err = extensions.NewPGClient("db", b.Config, b.Logger)
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) configureDatabases() {
	b.configureMarathonDatabase()
	b.configurePushDatabase()
}

func (b *CreateBatchesWorker) readCSVFromS3(csvPath string) *[]string {
	csvFile, err := extensions.S3GetObject(b.S3Client, csvPath)
	checkErr(b.Logger, err)
	r := csv.NewReader(*csvFile)
	lines, err := r.ReadAll()
	checkErr(b.Logger, err)
	res := []string{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		res = append(res, line[0])
	}
	return &res
}

func (b *CreateBatchesWorker) updateTotalBatches(totalBatches int, job *model.Job) {
	job.TotalBatches = totalBatches
	_, err := b.MarathonDB.DB.Model(job).Set("total_batches = total_batches + ?total_batches").Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) setTotalUsers(totalUsers int, job *model.Job) {
	job.TotalUsers = totalUsers
	_, err := b.MarathonDB.DB.Model(job).Set("total_users = ?", totalUsers).Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) computeTotalUsersAndBatchesSent(c <-chan *SentBatches, job *model.Job, batchesSent *int, totalUsers *int, wg *sync.WaitGroup) {
	for sent := range c {
		*batchesSent += (*sent).NumBatches
		*totalUsers += (*sent).TotalUsers
		wg.Done()
	}
}

func (b *CreateBatchesWorker) getCSVUserBatchFromPG(userIds *[]string, appName, service string) *[]User {
	var users []User
	_, err := b.PushDB.DB.Query(&users, fmt.Sprintf("SELECT user_id, token, locale, tz FROM %s WHERE user_id IN (?)", GetPushDBTableName(appName, service)), pg.In(*userIds))
	checkErr(b.Logger, err)
	return &users
}

func (b *CreateBatchesWorker) processBatch(c <-chan *Batch, batchesSentCH chan<- *SentBatches, job *model.Job, wg *sync.WaitGroup, wgBatchesSent *sync.WaitGroup) {
	l := b.Logger
	for batch := range c {
		usersFromBatch := b.getCSVUserBatchFromPG((*batch).UserIds, job.App.Name, job.Service)
		numUsersFromBatch := len(*usersFromBatch)
		log.I(l, "got users from db", func(cm log.CM) {
			cm.Write(zap.Int("usersInBatch", numUsersFromBatch))
		})
		bucketsByTZ := SplitUsersInBucketsByTZ(usersFromBatch)
		for tz, users := range bucketsByTZ {
			log.D(l, "batch of users for tz", func(cm log.CM) {
				cm.Write(zap.Int("numUsers", len(*users)), zap.String("tz", tz))
			})
		}
		markProcessedPage((*batch).PageID, job.ID, b.RedisClient)
		if job.Localized {
			b.sendLocalizedBatches(bucketsByTZ, job)
		} else {
			b.sendBatches(bucketsByTZ, job)
		}
		wgBatchesSent.Add(1)
		batchesSentCH <- &SentBatches{
			NumBatches: len(bucketsByTZ),
			TotalUsers: numUsersFromBatch,
		}
		wg.Done()
	}
}

func (b *CreateBatchesWorker) sendLocalizedBatches(batches map[string]*[]User, job *model.Job) {
	l := b.Logger
	for tz, users := range batches {
		offset, err := GetTimeOffsetFromUTCInSeconds(tz, b.Logger)
		checkErr(b.Logger, err)
		t := time.Unix(0, job.StartsAt)
		localizedTime := t.Add(time.Duration(offset) * time.Second)
		log.I(l, "scheduling batch of users to process batches worker", func(cm log.CM) {
			cm.Write(zap.Int("numUsers", len(*users)),
				zap.String("at", localizedTime.String()),
			)
		})
		isLocalizedTimeInPast := time.Now().After(localizedTime)
		if isLocalizedTimeInPast {
			if job.PastTimeStrategy == "skip" {
				continue
			} else {
				localizedTime = localizedTime.Add(time.Duration(24) * time.Hour)
			}
		}
		_, err = b.Workers.ScheduleProcessBatchJob(job.ID.String(), job.App.Name, users, localizedTime.UnixNano())
		checkErr(l, err)
	}
}

func (b *CreateBatchesWorker) sendBatches(batches map[string]*[]User, job *model.Job) {
	l := b.Logger
	for tz, users := range batches {
		log.I(l, "sending batch of users to process batches worker", func(cm log.CM) {
			cm.Write(zap.Int("numUsers", len(*users)), zap.String("tz", tz))
		})
		_, err := b.Workers.CreateProcessBatchJob(job.ID.String(), job.App.Name, users)
		checkErr(l, err)
	}
}

func (b *CreateBatchesWorker) createBatchesUsingCSV(job *model.Job, isReexecution bool) error {
	l := b.Logger
	userIds := b.readCSVFromS3(job.CSVPath)
	numPushes := len(*userIds)
	pages := int(math.Ceil(float64(numPushes) / float64(b.DBPageSize)))
	l.Info("grabing pages from pg", zap.Int("pagesToComplete", pages))
	var wg sync.WaitGroup
	var wgBatchesSent sync.WaitGroup
	pgCH := make(chan *Batch, pages)
	batchesSentCH := make(chan *SentBatches)
	wg.Add(pages)
	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processBatch(pgCH, batchesSentCH, job, &wg, &wgBatchesSent)
	}
	batchesSent := 0
	totalUsers := 0
	go b.computeTotalUsersAndBatchesSent(batchesSentCH, job, &batchesSent, &totalUsers, &wgBatchesSent)
	for i := 0; i < pages; i++ {
		if isReexecution && isPageProcessed(i, job.ID, b.RedisClient, b.Logger) {
			log.I(l, "job is reexecution and page is already processed", func(cm log.CM) {
				cm.Write(zap.String("jobID", job.ID.String()), zap.Int("page", i))
			})
			wg.Done()
			continue
		}
		userBatch := b.getPage(i, userIds)
		pgCH <- &Batch{
			UserIds: &userBatch,
			PageID:  i,
		}
	}
	wg.Wait()
	wgBatchesSent.Wait()
	close(pgCH)
	close(batchesSentCH)
	updateTotalBatches(batchesSent, job, b.MarathonDB.DB, b.Logger)
	b.setTotalUsers(totalUsers, job)
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
	arr, err := message.Args().Array()
	checkErr(b.Logger, err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(b.Logger, err)
	isReexecution := checkIsReexecution(id, b.RedisClient, b.Logger)
	l := b.Logger.With(
		zap.Int("batchSize", b.BatchSize),
		zap.Int("dbPageSize", b.DBPageSize),
		zap.String("jobID", id.String()),
		zap.Bool("isReexecution", isReexecution),
	)
	log.I(l, "starting create_batches_worker")
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)
	if job.Status == stoppedJobStatus {
		l.Info("stopped job create_batches_worker")
		return
	}
	if job.DBPageSize == 0 {
		b.MarathonDB.DB.Model(job).Set("db_page_size = ?", b.DBPageSize).Returning("*").Update()
	} else if job.DBPageSize != b.DBPageSize {
		b.DBPageSize = job.DBPageSize
		log.I(l, "Changing DBPageSize to the job value", func(cm log.CM) {
			cm.Write(zap.Int("dbPageSize", job.DBPageSize))
		})
	}
	if len(job.CSVPath) > 0 {
		err := b.createBatchesUsingCSV(job, isReexecution)
		checkErr(l, err)
		b.RedisClient.Expire(fmt.Sprintf("%s-processedpages", job.ID.String()), time.Second*3600)
		log.I(l, "finished create_batches_worker")
	} else {
		log.I(l, "panicked create_batches_worker")
		panic(fmt.Errorf("no csvPath passed to worker"))
	}
}
