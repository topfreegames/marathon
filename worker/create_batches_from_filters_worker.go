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
	"fmt"
	"math"
	"sync"

	"gopkg.in/redis.v5"

	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// CreateBatchesFromFiltersWorker is the CreateBatchesUsingFiltersWorker struct
type CreateBatchesFromFiltersWorker struct {
	Logger                    zap.Logger
	PushDB                    *extensions.PGClient
	MarathonDB                *extensions.PGClient
	Workers                   *Worker
	Config                    *viper.Viper
	DBPageSize                int
	PageProcessingConcurrency int
	RedisClient               *redis.Client
}

// NewCreateBatchesFromFiltersWorker gets a new CreateBatchesFromFiltersWorker
func NewCreateBatchesFromFiltersWorker(config *viper.Viper, logger zap.Logger, workers *Worker) *CreateBatchesFromFiltersWorker {
	b := &CreateBatchesFromFiltersWorker{
		Config:  config,
		Logger:  logger.With(zap.String("worker", "CreateBatchesFromFiltersWorker")),
		Workers: workers,
	}
	b.configure()
	log.D(logger, "Configured CreateBatchesFromFiltersWorker successfully.")
	return b
}

func (b *CreateBatchesFromFiltersWorker) configurePushDatabase() {
	var err error
	b.PushDB, err = extensions.NewPGClient("push.db", b.Config, b.Logger)
	checkErr(b.Logger, err)
}

func (b *CreateBatchesFromFiltersWorker) configureMarathonDatabase() {
	var err error
	b.MarathonDB, err = extensions.NewPGClient("db", b.Config, b.Logger)
	checkErr(b.Logger, err)
}

func (b *CreateBatchesFromFiltersWorker) loadConfigurationDefaults() {
	b.Config.SetDefault("workers.createBatchesFromFilters.dbPageSize", 1000)
	b.Config.SetDefault("workers.createBatchesFromFilters.pageProcessingConcurrency", 1)
}

func (b *CreateBatchesFromFiltersWorker) loadConfiguration() {
	b.DBPageSize = b.Config.GetInt("workers.createBatchesFromFilters.dbPageSize")
	b.PageProcessingConcurrency = b.Config.GetInt("workers.createBatchesFromFilters.pageProcessingConcurrency")
}

func (b *CreateBatchesFromFiltersWorker) configureDatabases() {
	b.configureMarathonDatabase()
	b.configurePushDatabase()
}

func (b *CreateBatchesFromFiltersWorker) configureRedisClient() {
	r, err := extensions.NewRedis("workers", b.Config, b.Logger)
	checkErr(b.Logger, err)
	b.RedisClient = r
}

func (b *CreateBatchesFromFiltersWorker) configure() {
	b.loadConfigurationDefaults()
	b.loadConfiguration()
	b.configureDatabases()
	b.configureRedisClient()
}

func (b *CreateBatchesFromFiltersWorker) getPageFromDBWithFilters(job *model.Job, page DBPage) []User {
	filters := job.Filters
	whereClause := GetWhereClauseFromFilters(filters)
	limit := b.DBPageSize
	query := fmt.Sprintf("SELECT user_id, token, locale, tz FROM %s WHERE %s AND seq_id > %d LIMIT %d;", GetPushDBTableName(job.App.Name, job.Service), whereClause, page.Offset, limit)
	var users []User
	_, err := b.PushDB.DB.Query(&users, query)
	checkErr(b.Logger, err)
	return users
}

func (b *CreateBatchesFromFiltersWorker) processPages(c <-chan DBPage, batchesSentCH chan<- int, job *model.Job, wg *sync.WaitGroup) {
	for page := range c {
		users := b.getPageFromDBWithFilters(job, page)
		log.I(b.Logger, "got users from db", func(cm log.CM) {
			cm.Write(zap.Int("usersInBatch", len(users)))
		})
		bucketsByTZ := SplitUsersInBucketsByTZ(&users)
		for tz, users := range bucketsByTZ {
			log.D(b.Logger, "batch of users for tz", func(cm log.CM) {
				cm.Write(zap.Int("numUsers", len(users)), zap.String("tz", tz))
			})
		}
		markProcessedPage(page.Page, job.ID, b.RedisClient)
		sendBatches(&bucketsByTZ, job, b.Logger, b.Workers)
		batchesSentCH <- len(bucketsByTZ)
		wg.Done()
	}
}

func (b *CreateBatchesFromFiltersWorker) preprocessPages(job *model.Job) ([]DBPage, int) {
	filters := job.Filters
	var count int
	whereClause := GetWhereClauseFromFilters(filters)
	query := fmt.Sprintf("SELECT count(1) FROM %s WHERE %s;", GetPushDBTableName(job.App.Name, job.Service), whereClause)
	_, err := b.PushDB.DB.Query(&count, query)
	pageCount := int(math.Ceil(float64(count) / float64(b.DBPageSize)))
	checkErr(b.Logger, err)
	pages := []DBPage{}
	nextPageOffset := 0
	for page := 0; page < pageCount; page++ {
		pages = append(pages, DBPage{
			Page:   page,
			Offset: nextPageOffset,
		})
		query := fmt.Sprintf("SELECT max(q.seq_id) FROM (SELECT seq_id FROM %s WHERE seq_id > %d AND %s LIMIT %d) AS q;", GetPushDBTableName(job.App.Name, job.Service), nextPageOffset, whereClause, b.DBPageSize)
		log.I(b.Logger, "Querying database.", func(cm log.CM) {
			cm.Write(zap.String("query", query))
		})
		_, err := b.PushDB.DB.Query(&nextPageOffset, query)
		checkErr(b.Logger, err)
	}
	return pages, pageCount
}

func (b *CreateBatchesFromFiltersWorker) computeBatchesSent(c <-chan int, job *model.Job) {
	batchesSent := 0
	for sent := range c {
		batchesSent += sent
	}
	updateTotalBatches(batchesSent, job, b.MarathonDB.DB, b.Logger)
}

func (b *CreateBatchesFromFiltersWorker) createBatchesFromFilters(job *model.Job, isReexecution bool) error {
	pages, pageCount := b.preprocessPages(job)
	var wg sync.WaitGroup
	pageCH := make(chan DBPage, pageCount)
	batchesSentCH := make(chan int)
	wg.Add(pageCount)
	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processPages(pageCH, batchesSentCH, job, &wg)
	}
	go b.computeBatchesSent(batchesSentCH, job)
	for i := 0; i < pageCount; i++ {
		if isReexecution && isPageProcessed(i, job.ID, b.RedisClient, b.Logger) {
			log.I(b.Logger, "job is reexecution and page is already processed", func(cm log.CM) {
				cm.Write(zap.String("jobID", job.ID.String()), zap.Int("page", i))
			})
			wg.Done()
			continue
		}
		pageCH <- DBPage{
			Page:   pages[i].Page,
			Offset: pages[i].Offset,
		}
	}
	wg.Wait()
	close(pageCH)
	close(batchesSentCH)
	return nil
}

// Process processes the messages sent to worker queue
func (b *CreateBatchesFromFiltersWorker) Process(message *workers.Msg) {
	arr, err := message.Args().Array()
	checkErr(b.Logger, err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(b.Logger, err)
	isReexecution := checkIsReexecution(id, b.RedisClient, b.Logger)
	l := b.Logger.With(
		zap.Int("dbPageSize", b.DBPageSize),
		zap.String("jobID", id.String()),
		zap.Bool("isReexecution", isReexecution),
	)
	log.I(l, "starting create_batches_using_filters_worker")
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)
	if len(job.Filters) > 0 {
		err := b.createBatchesFromFilters(job, isReexecution)
		checkErr(l, err)
	} else {
		panic(fmt.Errorf("no filters passed to worker"))
	}
}
