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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"math/rand"

	"gopkg.in/pg.v5"
	"gopkg.in/redis.v5"

	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameCreateBatches = "create_batches_worker"

// CreateBatchesWorker is the CreateBatchesWorker struct
type CreateBatchesWorker struct {
	BatchSize                 int
	Config                    *viper.Viper
	DBPageSize                int
	Logger                    zap.Logger
	MarathonDB                *extensions.PGClient
	PageProcessingConcurrency int
	PushDB                    *extensions.PGClient
	RedisClient               *redis.Client
	S3Client                  interfaces.S3
	SendgridClient            *extensions.SendgridClient
	Workers                   *Worker
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
	b.configureSendgridClient()
}

func (b *CreateBatchesWorker) configureSendgridClient() {
	apiKey := b.Config.GetString("sendgrid.key")
	if apiKey != "" {
		b.SendgridClient = extensions.NewSendgridClient(b.Config, b.Logger, apiKey)
	}
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

// ReadCSVFromS3 reads CSV from S3 and return correspondent array of strings
func (b *CreateBatchesWorker) ReadCSVFromS3(csvPath string) []string {
	csvFileBytes, err := b.S3Client.GetObject(csvPath)
	checkErr(b.Logger, err)
	for i, b := range csvFileBytes {
		if b == 0x0D {
			csvFileBytes[i] = 0x0A
		}
	}
	r := csv.NewReader(bytes.NewReader(csvFileBytes))
	lines, err := r.ReadAll()
	checkErr(b.Logger, err)
	res := []string{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		res = append(res, line[0])
	}
	return res
}

func (b *CreateBatchesWorker) updateTotalBatches(totalBatches int, job *model.Job) {
	job.TotalBatches = totalBatches
	// coalesce is necessary since total_batches can be null
	_, err := b.MarathonDB.DB.Model(job).Set("total_batches = coalesce(total_batches, 0) + ?", totalBatches).Where("id = ?", job.ID).Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) updateTotalTokens(totalTokens int, job *model.Job) {
	job.TotalTokens = totalTokens
	// coalesce is necessary since total_tokens can be null
	_, err := b.MarathonDB.DB.Model(job).Set("total_tokens = coalesce(total_tokens, 0) + ?", totalTokens).Where("id = ?", job.ID).Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) computeTotalTokensAndBatchesSent(c <-chan *SentBatches, job *model.Job, wg *sync.WaitGroup) {
	for sent := range c {
		b.updateTotalBatches((*sent).NumBatches, job)
		b.updateTotalTokens((*sent).TotalTokens, job)
		wg.Done()
	}
}

func (b *CreateBatchesWorker) getCSVUserBatchFromPG(userIds *[]string, appName, service string) *[]User {
	var users []User
	start := time.Now()
	_, err := b.PushDB.DB.Query(&users, fmt.Sprintf("SELECT user_id, token, locale, tz, fiu, adid, vendor_id FROM %s WHERE user_id IN (?)", GetPushDBTableName(appName, service)), pg.In(*userIds))
	labels := []string{fmt.Sprintf("game:%s", appName), fmt.Sprintf("platform:%s", service)}
	b.Workers.Statsd.Timing("get_csv_batch_from_pg", time.Now().Sub(start), labels, 1)
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
			NumBatches:  len(bucketsByTZ),
			TotalTokens: numUsersFromBatch,
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

func (b *CreateBatchesWorker) sendControlGroupToS3(job *model.Job, controlGroup []string) {
	folder := b.Config.GetString("s3.controlGroupFolder")
	bucket := b.Config.GetString("s3.bucket")
	csvBuffer := &bytes.Buffer{}
	csvWriter := io.Writer(csvBuffer)
	csvWriter.Write([]byte("controlGroupUserIds\n"))
	for _, user := range controlGroup {
		csvWriter.Write([]byte(fmt.Sprintf("%s\n", user)))
	}
	writePath := fmt.Sprintf("%s/job-%s.csv", folder, job.ID.String())
	csvBytes := csvBuffer.Bytes()
	_, err := b.S3Client.PutObject(writePath, &csvBytes)
	checkErr(b.Logger, err)
	b.updateJobControlGroupCSVPath(job, fmt.Sprintf("%s/%s", bucket, writePath))
}

func (b *CreateBatchesWorker) updateJobControlGroupCSVPath(job *model.Job, csvPath string) {
	job.ControlGroupCSVPath = csvPath
	_, err := b.MarathonDB.DB.Model(job).Set("control_group_csv_path = ?control_group_csv_path").Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) updateTotalUsers(job *model.Job, totalUsers int) {
	job.TotalUsers = totalUsers
	_, err := b.MarathonDB.DB.Model(job).Set("total_users = coalesce(?total_users,0)").Update()
	checkErr(b.Logger, err)
}

func (b *CreateBatchesWorker) createBatchesUsingCSV(job *model.Job, isReexecution bool, dbPageSize int) error {
	l := b.Logger
	userIds := b.ReadCSVFromS3(job.CSVPath)
	controlGroupSize := int(math.Ceil(float64(len(userIds)) * job.ControlGroup))
	if controlGroupSize > 0 {
		if controlGroupSize >= len(userIds) {
			panic("control group size cannot be higher than number of users")
		}
		log.I(l, "this job has a control group!", func(cm log.CM) {
			cm.Write(
				zap.Int("controlGroupSize", controlGroupSize),
				zap.String("jobID", job.ID.String()),
			)
		})
		// shuffle slice in place
		for i := range userIds {
			j := rand.Intn(i + 1)
			userIds[i], userIds[j] = userIds[j], userIds[i]
		}
		// grab control group
		controlGroup := userIds[len(userIds)-controlGroupSize:]
		b.sendControlGroupToS3(job, controlGroup)
		// cut control group from the slice
		userIds = append(userIds[:len(userIds)-controlGroupSize], userIds[len(userIds):]...)
		log.I(l, "control group cut from the users", func(cm log.CM) {
			cm.Write(
				zap.Int("usersRemaining", len(userIds)),
			)
		})
	}
	numPushes := len(userIds)
	log.D(l, "finished reading csv from s3", func(cm log.CM) {
		cm.Write(zap.Int("numPushes", numPushes),
			zap.Int("dbPageSize", dbPageSize),
		)
	})
	b.updateTotalUsers(job, numPushes)
	pages := int(math.Max(math.Ceil(float64(numPushes)/float64(dbPageSize)), 1))
	if numPushes == 0 {
		pages = 0
	}
	l.Info("grabing pages from pg", zap.Int("pagesToComplete", pages))
	var wg sync.WaitGroup
	var wgBatchesSent sync.WaitGroup
	pgCH := make(chan *Batch, pages)
	batchesSentCH := make(chan *SentBatches)
	wg.Add(pages)
	for i := 0; i < b.PageProcessingConcurrency; i++ {
		go b.processBatch(pgCH, batchesSentCH, job, &wg, &wgBatchesSent)
	}
	go b.computeTotalTokensAndBatchesSent(batchesSentCH, job, &wgBatchesSent)
	for i := 0; i < pages; i++ {
		if isReexecution && isPageProcessed(i, job.ID, b.RedisClient, b.Logger) {
			log.I(l, "job is reexecution and page is already processed", func(cm log.CM) {
				cm.Write(zap.String("jobID", job.ID.String()), zap.Int("page", i))
			})
			wg.Done()
			continue
		}
		userBatch := b.getPage(i, dbPageSize, userIds)
		pgCH <- &Batch{
			UserIds: &userBatch,
			PageID:  i,
		}
	}
	wg.Wait()
	wgBatchesSent.Wait()
	close(pgCH)
	close(batchesSentCH)
	return nil
}

func (b *CreateBatchesWorker) getPage(page, dbPageSize int, users []string) []string {
	start := page * dbPageSize
	end := (page + 1) * dbPageSize
	if start >= len(users) {
		return nil
	}
	if end > len(users) {
		end = len(users)
	}
	return users[start:end]
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
		zap.String("worker", nameCreateBatches),
	)
	log.I(l, "starting")
	job := &model.Job{
		ID: id,
	}
	job.TagRunning(b.MarathonDB, nameCreateBatches, "starting")
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)
	if job.Status == stoppedJobStatus {
		l.Info("stopped job")
		return
	}
	dbPageSize := b.DBPageSize
	if job.DBPageSize == 0 {
		b.MarathonDB.DB.Model(job).Set("db_page_size = ?", b.DBPageSize).Returning("*").Update()
	} else if job.DBPageSize != b.DBPageSize {
		dbPageSize = job.DBPageSize
		log.I(l, "Using job DBPageSize value", func(cm log.CM) {
			cm.Write(zap.Int("dbPageSize", job.DBPageSize))
		})
	}
	if len(job.CSVPath) > 0 {
		err = b.createBatchesUsingCSV(job, isReexecution, dbPageSize)
		if err != nil {
			job.TagError(b.MarathonDB, nameCreateBatches, err.Error())
			checkErr(l, err)
		}
		b.RedisClient.Expire(fmt.Sprintf("%s-processedpages", job.ID.String()), time.Second*3600)
		updatedJob := &model.Job{
			ID: id,
		}
		err = b.MarathonDB.DB.Model(updatedJob).Column("job.*").Where("job.id = ?", updatedJob.ID).Select()
		if err != nil {
			job.TagError(b.MarathonDB, nameCreateBatches, err.Error())
			checkErr(l, err)
		}
		if updatedJob.TotalTokens == 0 {
			_, err := b.MarathonDB.DB.Model(job).Set("status = 'stopped', updated_at = ?", time.Now().UnixNano()).Where("id = ?", updatedJob.ID).Update()
			checkErr(l, err)

			if b.SendgridClient != nil {
				msg := "Your job was automatically stopped because no tokens were found matching the user ids given in the CSV file. Please verify the uploaded CSV and create a new job."
				err = email.SendStoppedJobEmail(b.SendgridClient, updatedJob, job.App.Name, msg)
				b.checkErr(job, err)

			}
		}
		log.I(l, "finished")
		job.TagSuccess(b.MarathonDB, nameCreateBatches, "finished")
	} else {
		log.I(l, "panicked")
		checkErr(l, fmt.Errorf("no csvPath passed to worker"))
		b.checkErr(job, err)
	}
}

func (b *CreateBatchesWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.MarathonDB, nameCreateBatches, err.Error())
		checkErr(b.Logger, err)
	}
}
