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
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"

	"gopkg.in/redis.v5"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	"github.com/willf/bloom"
)

// CreateBatchesFromFiltersWorker is the CreateBatchesUsingFiltersWorker struct
type CreateBatchesFromFiltersWorker struct {
	Logger                    zap.Logger
	PushDB                    *extensions.PGClient
	MarathonDB                *extensions.PGClient
	Workers                   *Worker
	Config                    *viper.Viper
	DBPageSize                int
	S3Client                  s3iface.S3API
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
	b.Logger.Debug("Configured CreateBatchesFromFiltersWorker successfully")
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

func (b *CreateBatchesFromFiltersWorker) configureS3Client() {
	s3Client, err := extensions.NewS3(b.Config, b.Logger)
	checkErr(b.Logger, err)
	b.S3Client = s3Client
}

func (b *CreateBatchesFromFiltersWorker) configure() {
	b.loadConfigurationDefaults()
	b.loadConfiguration()
	b.configureDatabases()
	b.configureS3Client()
	b.configureRedisClient()
}

func (b *CreateBatchesFromFiltersWorker) getPageFromDBWithFilters(job *model.Job, page DBPage) []User {
	filters := job.Filters
	whereClause := GetWhereClauseFromFilters(filters)
	limit := b.DBPageSize
	var query string
	if (whereClause) != "" {
		query = fmt.Sprintf("SELECT user_id FROM %s WHERE seq_id > %d AND %s ORDER BY seq_id ASC LIMIT %d;", GetPushDBTableName(job.App.Name, job.Service), page.Offset, whereClause, limit)
	} else {
		query = fmt.Sprintf("SELECT user_id FROM %s WHERE seq_id > %d ORDER BY seq_id ASC LIMIT %d;", GetPushDBTableName(job.App.Name, job.Service), page.Offset, limit)
	}
	var users []User
	_, err := b.PushDB.DB.Query(&users, query)
	checkErr(b.Logger, err)
	return users
}

func (b *CreateBatchesFromFiltersWorker) preprocessPages(job *model.Job) ([]DBPage, int, int) {
	filters := job.Filters
	var count int
	whereClause := GetWhereClauseFromFilters(filters)
	var query string
	if (whereClause) != "" {
		query = fmt.Sprintf("SELECT count(1) FROM %s WHERE %s;", GetPushDBTableName(job.App.Name, job.Service), whereClause)
	} else {
		query = fmt.Sprintf("SELECT count(1) FROM %s;", GetPushDBTableName(job.App.Name, job.Service))
	}
	_, err := b.PushDB.DB.Query(&count, query)
	if count == 0 {
		panic(fmt.Errorf("no users matching the filters"))
	}
	pageCount := int(math.Ceil(float64(count) / float64(b.DBPageSize)))
	checkErr(b.Logger, err)
	pages := []DBPage{}
	nextPageOffset := 0
	for page := 0; page < pageCount; page++ {
		pages = append(pages, DBPage{
			Page:   page,
			Offset: nextPageOffset,
		})
		if (whereClause) != "" {
			query = fmt.Sprintf("SELECT max(q.seq_id) FROM (SELECT seq_id FROM %s WHERE seq_id > %d AND %s ORDER BY seq_id ASC LIMIT %d) AS q;", GetPushDBTableName(job.App.Name, job.Service), nextPageOffset, whereClause, b.DBPageSize)
		} else {
			query = fmt.Sprintf("SELECT max(q.seq_id) FROM (SELECT seq_id FROM %s WHERE seq_id > %d ORDER BY seq_id ASC LIMIT %d) AS q;", GetPushDBTableName(job.App.Name, job.Service), nextPageOffset, b.DBPageSize)
		}
		b.Logger.Info("Querying database", zap.String("query", query))
		_, err := b.PushDB.DB.Query(&nextPageOffset, query)
		checkErr(b.Logger, err)
	}
	return pages, pageCount, count
}

func (b *CreateBatchesFromFiltersWorker) processPages(c <-chan DBPage, writeToCSVCH chan<- *[]User, job *model.Job, wg *sync.WaitGroup, wgCSV *sync.WaitGroup) {
	for page := range c {
		users := b.getPageFromDBWithFilters(job, page)
		b.Logger.Info("got users from db", zap.Int("usersInBatch", len(users)))
		wgCSV.Add(1)
		writeToCSVCH <- &users
		wg.Done()
	}
}

func (b *CreateBatchesFromFiltersWorker) writeUserPageIntoCSV(c <-chan *[]User, bFilter *bloom.BloomFilter, csvWriter *io.Writer, wgCSV *sync.WaitGroup) {
	(*csvWriter).Write([]byte("userIds\n"))
	for users := range c {
		for _, user := range *users {
			if !bFilter.TestString(user.UserID) {
				(*csvWriter).Write([]byte(fmt.Sprintf("%s\n", user.UserID)))
				bFilter.AddString(user.UserID)
			}
		}
		wgCSV.Done()
	}
}

func (b *CreateBatchesFromFiltersWorker) createBatchesFromFilters(job *model.Job, csvWriter *io.Writer) error {
	pages, pageCount, usersCount := b.preprocessPages(job)
	var wg sync.WaitGroup
	var wgCSV sync.WaitGroup
	pageCH := make(chan DBPage, pageCount)
	csvWriterCH := make(chan *[]User)
	wg.Add(pageCount)

	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processPages(pageCH, csvWriterCH, job, &wg, &wgCSV)
	}
	rate := 1E-8
	bFilter := bloom.NewWithEstimates(uint(usersCount), rate)
	go b.writeUserPageIntoCSV(csvWriterCH, bFilter, csvWriter, &wgCSV)
	for i := 0; i < pageCount; i++ {
		pageCH <- DBPage{
			Page:   pages[i].Page,
			Offset: pages[i].Offset,
		}
	}
	wg.Wait()
	wgCSV.Wait()
	close(pageCH)
	close(csvWriterCH)
	return nil
}

func (b *CreateBatchesFromFiltersWorker) updateJobCSVPath(job *model.Job, csvPath string) {
	job.CSVPath = csvPath
	_, err := b.MarathonDB.DB.Model(job).Set("csv_path = ?csv_path").Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesFromFiltersWorker) sendCSVToS3AndCreateCreateBatchesJob(csvBytes *[]byte, job *model.Job) {
	folder := b.Config.GetString("s3.folder")
	bucket := b.Config.GetString("s3.bucket")
	writePath := fmt.Sprintf("%s/job-%s.csv", folder, job.ID.String())
	b.Logger.Info("uploading file to s3", zap.String("path", writePath))
	err := extensions.S3PutObject(b.Config, b.S3Client, writePath, csvBytes)
	checkErr(b.Logger, err)
	b.updateJobCSVPath(job, fmt.Sprintf("%s/%s", bucket, writePath))
	jid, err := b.Workers.CreateBatchesJob(&[]string{job.ID.String()})
	checkErr(b.Logger, err)
	b.Logger.Info("created create batches job", zap.String("jid", jid))
}

// Process processes the messages sent to worker queue
func (b *CreateBatchesFromFiltersWorker) Process(message *workers.Msg) {
	arr, err := message.Args().Array()
	checkErr(b.Logger, err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(b.Logger, err)
	l := b.Logger.With(
		zap.Int("dbPageSize", b.DBPageSize),
		zap.String("jobID", id.String()),
	)
	l.Info("starting create_batches_using_filters_worker")
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)
	csvBuffer := &bytes.Buffer{}
	csvWriter := io.Writer(csvBuffer)
	err = b.createBatchesFromFilters(job, &csvWriter)
	checkErr(l, err)
	csvBytes := csvBuffer.Bytes()
	b.sendCSVToS3AndCreateCreateBatchesJob(&csvBytes, job)
	l.Info("finished create_batches_using_filters_worker")
}
