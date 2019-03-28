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
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	workers "github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	"github.com/willf/bloom"
	redis "gopkg.in/redis.v5"
)

const nameCreateBatchesFromFilters = "create_batches_using_filters_worker"

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

func (b *CreateBatchesFromFiltersWorker) getPageFromDBWithFilters(job *model.Job, page DBPage) *[]User {
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
	start := time.Now()
	_, err := b.PushDB.DB.Query(&users, query)
	labels := []string{fmt.Sprintf("game:%s", job.App.Name), fmt.Sprintf("platform:%s", job.Service)}
	b.Workers.Statsd.Timing("get_page_with_filters", time.Now().Sub(start), labels, 1)

	checkErr(b.Logger, err)
	return &users
}

func (b *CreateBatchesFromFiltersWorker) preprocessPages(job *model.Job, stageStatus *StageStatus) ([]DBPage, int, int) {
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
		checkErr(b.Logger, fmt.Errorf("no users matching the filters"))
	}
	pageCount := int(math.Ceil(float64(count) / float64(b.DBPageSize)))
	checkErr(b.Logger, err)
	pages := []DBPage{}
	nextPageOffset := 0

	preProcessStats, err := stageStatus.NewSubStage("pre processing pages", pageCount)
	checkErr(b.Logger, err)

	tx, err := b.PushDB.DB.Begin()
	checkErr(b.Logger, err)

	defer tx.Rollback()

	if (whereClause) != "" {
		query = fmt.Sprintf("DECLARE cursor CURSOR FOR SELECT seq_id FROM %s WHERE %s;", GetPushDBTableName(job.App.Name, job.Service), whereClause)
	} else {
		query = fmt.Sprintf("DECLARE cursor CURSOR FOR SELECT seq_id FROM %s;", GetPushDBTableName(job.App.Name, job.Service))
	}
	tx.Query(nil, query)

	for page := 0; page < pageCount; page++ {
		pages = append(pages, DBPage{
			Page:   page,
			Offset: nextPageOffset,
		})
		_, err := tx.Query(&nextPageOffset, fmt.Sprintf("FETCH RELATIVE +%d FROM cursor;", b.DBPageSize))
		checkErr(b.Logger, err)
		preProcessStats.IncrProgress()
	}
	tx.Commit()

	_, err = b.MarathonDB.DB.Model(job).Set("total_tokens = ?", count).Where("id = ?", job.ID).Update()
	checkErr(b.Logger, err)

	return pages, pageCount, count
}

func (b *CreateBatchesFromFiltersWorker) processPages(c <-chan DBPage, writeToCSVCH chan<- *[]User, job *model.Job, wg *sync.WaitGroup, wgCSV *sync.WaitGroup, stageStatus *StageStatus) {
	for page := range c {
		users := b.getPageFromDBWithFilters(job, page)
		b.Logger.Info("got users from db", zap.Int("usersInBatch", len(*users)), zap.Int("Offset", page.Offset))
		labels := []string{fmt.Sprintf("game:%s", job.App.Name), fmt.Sprintf("platform:%s", job.Service)}
		b.Workers.Statsd.Incr("process_pages", labels, 1)
		wgCSV.Add(1)
		writeToCSVCH <- users
		wg.Done()

		err := stageStatus.IncrProgress()
		checkErr(b.Logger, err)
	}
}

func (b *CreateBatchesFromFiltersWorker) writeUserPageIntoCSV(job *model.Job, c <-chan *[]User, bFilter *bloom.BloomFilter, csvWriter *io.Writer, wgCSV *sync.WaitGroup, stageStatus *StageStatus) {
	(*csvWriter).Write([]byte("userIds\n"))
	for users := range c {
		labels := []string{fmt.Sprintf("game:%s", job.App.Name), fmt.Sprintf("platform:%s", job.Service)}
		b.Workers.Statsd.Incr("write_user_page_into_csv", labels, 1)
		for _, user := range *users {
			if IsUserIDValid(user.UserID) && !bFilter.TestString(user.UserID) {
				(*csvWriter).Write([]byte(fmt.Sprintf("%s\n", user.UserID)))
				bFilter.AddString(user.UserID)
			}
		}
		wgCSV.Done()

		err := stageStatus.IncrProgress()
		checkErr(b.Logger, err)
	}
}

func (b *CreateBatchesFromFiltersWorker) createBatchesFromFilters(job *model.Job, csvWriter *io.Writer, stageStatus *StageStatus) error {
	pages, pageCount, usersCount := b.preprocessPages(job, stageStatus)
	var wg sync.WaitGroup
	var wgCSV sync.WaitGroup
	pageCH := make(chan DBPage, pageCount)
	csvWriterCH := make(chan *[]User)
	wg.Add(pageCount)

	processPageStatus, err := stageStatus.NewSubStage(
		"processing pages", pageCount)
	checkErr(b.Logger, err)

	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processPages(pageCH, csvWriterCH, job, &wg, &wgCSV, processPageStatus)
	}
	rate := 1E-8
	bFilter := bloom.NewWithEstimates(uint(usersCount), rate)

	writeCSVStats, err := stageStatus.NewSubStage(
		"writing to csv", pageCount)
	checkErr(b.Logger, err)

	go b.writeUserPageIntoCSV(job, csvWriterCH, bFilter, csvWriter, &wgCSV, writeCSVStats)
	for i := 0; i < pageCount; i++ {
		pageCH <- DBPage{
			Page:   pages[i].Page,
			Offset: pages[i].Offset,
		}
	}

	checkErr(b.Logger, err)

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

func (b *CreateBatchesFromFiltersWorker) sendCSVToS3AndCreateCreateBatchesJob(csvBytes *[]byte, job *model.Job) error {
	folder := b.Config.GetString("s3.folder")
	bucket := b.Config.GetString("s3.bucket")
	writePath := fmt.Sprintf("%s/job-%s.csv", folder, job.ID.String())
	b.Logger.Info("uploading file to s3", zap.String("path", writePath))
	err := extensions.S3PutObject(b.Config, b.S3Client, writePath, csvBytes)
	if err != nil {
		return err
	}
	b.updateJobCSVPath(job, fmt.Sprintf("%s/%s", bucket, writePath))
	jid, err := b.Workers.CreateBatchesJob(&[]string{job.ID.String()})
	if err != nil {
		return err
	}
	b.Logger.Info("created create batches job", zap.String("jid", jid))
	return nil
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
		zap.String("worker", nameCreateBatchesFromFilters),
	)
	l.Info("starting")

	job := &model.Job{
		ID: id,
	}
	job.TagRunning(b.MarathonDB, nameCreateBatchesFromFilters, "starting")
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	if err != nil {
		job.TagError(b.MarathonDB, nameCreateBatchesFromFilters, err.Error())
		return
	}
	if job.Status == stoppedJobStatus {
		l.Info("stopped job")
		return
	}

	processStats, err := NewStageStatus(b.RedisClient, job.ID.String(),
		"1", "create batches from filter worker", 1)
	b.checkErr(job, err)

	csvBuffer := &bytes.Buffer{}
	csvWriter := io.Writer(csvBuffer)

	createBatchStats, err := processStats.NewSubStage(
		"creating batches from filters", 1)
	b.checkErr(job, err)

	err = b.createBatchesFromFilters(job, &csvWriter, createBatchStats)
	b.checkErr(job, err)
	createBatchStats.IncrProgress()

	uploadCSVStats, err := processStats.NewSubStage("uploading csv to s3", 1)
	b.checkErr(job, err)

	csvBytes := csvBuffer.Bytes()
	err = b.sendCSVToS3AndCreateCreateBatchesJob(&csvBytes, job)
	b.checkErr(job, err)

	l.Info("finished")
	err = uploadCSVStats.IncrProgress()
	b.checkErr(job, err)

	job.TagSuccess(b.MarathonDB, nameCreateBatchesFromFilters, "finished")
	processStats.IncrProgress()
}

func (b *CreateBatchesFromFiltersWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.MarathonDB, nameCreateBatchesFromFilters, err.Error())
		checkErr(b.Logger, err)
	}
}
