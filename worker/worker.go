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
	"os"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	raven "github.com/getsentry/raven-go"
	"github.com/jrallison/go-workers"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/uber-go/zap"
	redis "gopkg.in/redis.v5"
)

// Worker is the struct that will configure workers
type Worker struct {
	Logger                    zap.Logger
	PushDB                    *extensions.PGClient
	MarathonDB                *extensions.PGClient
	Config                    *viper.Viper
	DBPageSize                int
	S3Client                  interfaces.S3
	PageProcessingConcurrency int
	Statsd                    *statsd.Client
	RedisClient               *redis.Client
	ConfigPath                string
	SendgridClient            *extensions.SendgridClient
	Kafka                     interfaces.PushProducer
}

// NewWorker returns a configured worker
func NewWorker(l zap.Logger, configPath string) *Worker {
	worker := &Worker{
		Logger:     l,
		ConfigPath: configPath,
	}

	worker.configure()
	return worker
}

func (w *Worker) configure() {
	w.Config = viper.New()

	w.Config.SetConfigFile(w.ConfigPath)
	w.Config.SetEnvPrefix("marathon")
	w.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	w.Config.AutomaticEnv()

	if err := w.Config.ReadInConfig(); err == nil {
		w.Logger.Info("Loaded config file.", zap.String("configFile", w.Config.ConfigFileUsed()))
	} else {
		panic(err)
	}
	w.loadConfigurationDefaults()
	w.configureSentry()
	w.configureRedis()
	w.configureStatsd()
	w.configureWorkers()
	w.configureStatsd()
	w.configurePushDatabase()
	w.configureMarathonDatabase()
	w.configureS3Client()
	w.configureSendgrid()
	w.configureKafkaProducer()
}

func (w *Worker) loadConfigurationDefaults() {
	w.Config.SetDefault("workers.redis.server", "localhost:6379")
	w.Config.SetDefault("workers.redis.database", "0")
	w.Config.SetDefault("workers.redis.poolSize", "10")
	w.Config.SetDefault("workers.statsPort", 8081)
	w.Config.SetDefault("workers.concurrency", 10)
	w.Config.SetDefault("database.url", "postgres://localhost:5432/marathon?sslmode=disable")
	w.Config.SetDefault("workers.statsd.host", "127.0.0.1:8125")
	w.Config.SetDefault("workers.statsd.prefix", "marathon.")
}

func (w *Worker) configureSendgrid() {
	apiKey := w.Config.GetString("sendgrid.key")
	if apiKey != "" {
		w.SendgridClient = extensions.NewSendgridClient(w.Config, w.Logger, apiKey)
	}
}

func (w *Worker) configurePushDatabase() {
	var err error
	w.PushDB, err = extensions.NewPGClient("push.db", w.Config, w.Logger)
	checkErr(w.Logger, err)
}

func (w *Worker) configureMarathonDatabase() {
	var err error
	w.MarathonDB, err = extensions.NewPGClient("db", w.Config, w.Logger)
	checkErr(w.Logger, err)
}

func (w *Worker) configureStatsd() {
	host := w.Config.GetString("workers.statsd.host")
	prefix := w.Config.GetString("workers.statsd.prefix")

	client, err := statsd.New(host)
	if err != nil {
		return
	}
	client.Namespace = prefix
	w.Statsd = client
}

func (w *Worker) configureRedis() {
	redisHost := w.Config.GetString("workers.redis.host")
	redisPort := w.Config.GetInt("workers.redis.port")
	redisDatabase := w.Config.GetString("workers.redis.db")
	redisPassword := w.Config.GetString("workers.redis.pass")
	redisPoolsize := w.Config.GetString("workers.redis.poolSize")

	logger := w.Logger.With(
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.String("redisDB", redisDatabase),
		zap.String("redisPoolsize", redisPoolsize),
	)

	logger.Info("connecting to workers redis")

	// unique process id for this instance of workers (for recovery of inprogress jobs on crash)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	workers.Configure(map[string]string{
		"server":   fmt.Sprintf("%s:%d", redisHost, redisPort),
		"database": redisDatabase,
		"pool":     redisPoolsize,
		"process":  hostname,
		"password": redisPassword,
	})
	r, err := extensions.NewRedis("workers", w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.RedisClient = r
}

func (w *Worker) configureS3Client() {
	s3Client, err := extensions.NewS3(w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.S3Client = s3Client
}

func (w *Worker) configureWorkers() {
	p := NewProcessBatchWorker(w)
	d := NewDbToCsvWorker(w)
	k := NewCSVSplitWorker(w)
	c := NewCreateBatchesWorker(w)
	f := NewCreateBatchesFromFiltersWorker(w)
	r := NewResumeJobWorker(w)
	j := NewJobCompletedWorker(w)

	createBatchesWorkerConcurrency := w.Config.GetInt("workers.createBatches.concurrency")
	createDbToCsvWorkerConcurrency := w.Config.GetInt("workers.dbToCsv.concurrency")
	createCSVSplitWorkerConcurrency := w.Config.GetInt("workers.csvSplitWorker.concurrency")
	createBatchesFromFiltersWorkerConcurrency := w.Config.GetInt("workers.createBatchesFromFilters.concurrency")
	processBatchWorkerConcurrency := w.Config.GetInt("workers.processBatch.concurrency")
	resumeJobWorkerConcurrency := w.Config.GetInt("workers.resume.concurrency")
	jobCompletedWorkerConcurrency := w.Config.GetInt("workers.jobCompleted.concurrency")

	workers.Process("create_batches_from_filters_worker", f.Process, createBatchesFromFiltersWorkerConcurrency)
	workers.Process("db_to_csv_worker", d.Process, createDbToCsvWorkerConcurrency)

	workers.Process("csv_split_worker", k.Process, createCSVSplitWorkerConcurrency)
	workers.Process("create_batches_worker", c.Process, createBatchesWorkerConcurrency)
	workers.Process("process_batch_worker", p.Process, processBatchWorkerConcurrency)
	workers.Process("resume_job_worker", r.Process, resumeJobWorkerConcurrency)
	workers.Process("job_completed_worker", j.Process, jobCompletedWorkerConcurrency)
}

func (w *Worker) configureSentry() {
	l := w.Logger.With(
		zap.String("source", "worker"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := w.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (w *Worker) configureKafkaProducer() {
	var kafka *extensions.KafkaProducer
	var err error
	kafka, err = extensions.NewKafkaProducer(w.Config, w.Logger, w.Statsd)
	checkErr(w.Logger, err)
	w.Kafka = kafka
}

// CSVSplitJob creates a new CSVSplitWorker job
func (w *Worker) CSVSplitJob(jobID string) (string, error) {
	maxRetries := w.Config.GetInt("workers.csvSplitWorker.maxRetries")
	return workers.EnqueueWithOptions("csv_split_worker", "Add", jobID, workers.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// DbToCsvJob creates a new DbToCsvWorker job
func (w *Worker) DbToCsvJob(msg *ToCSVMessage) (string, error) {
	maxRetries := w.Config.GetInt("workers.dbToCsv.maxRetries")
	return workers.EnqueueWithOptions("db_to_csv_worker", "Add", msg, workers.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// CreateBatchesJob creates a new CreateBatchesWorker job
func (w *Worker) CreateBatchesJob(part *BatchPart) (string, error) {
	maxRetries := w.Config.GetInt("workers.createBatches.maxRetries")
	return workers.EnqueueWithOptions("create_batches_worker", "Add", part, workers.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// CreateBatchesFromFiltersJob creates a new CreateBatchesFromFiltersWorker job
func (w *Worker) CreateBatchesFromFiltersJob(jobID *[]string) (string, error) {
	maxRetries := w.Config.GetInt("workers.createBatchesFromFilters.maxRetries")
	return workers.EnqueueWithOptions("create_batches_from_filters_worker", "Add", jobID, workers.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// CreateProcessBatchJob creates a new ProcessBatchWorker job
func (w *Worker) CreateProcessBatchJob(jobID string, appName string, users *[]User) (string, error) {
	compressedUsers, err := CompressUsers(users)
	if err != nil {
		return "", err
	}
	return workers.Enqueue(
		"process_batch_worker",
		"Add",
		[]interface{}{jobID, appName, compressedUsers},
	)
}

// CreateResumeJob creates a new ResumeJobWorker job
func (w *Worker) CreateResumeJob(jobID *[]string) (string, error) {
	maxRetries := w.Config.GetInt("workers.resume.maxRetries")
	return workers.EnqueueWithOptions("resume_job_worker", "Add", jobID, workers.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// ScheduleCreateBatchesJob schedules a new CreateBatchesWorker job
func (w *Worker) ScheduleCreateBatchesJob(jobID *[]string, at int64) (string, error) {
	return workers.EnqueueWithOptions(
		"create_batches_worker",
		"Add",
		jobID,
		workers.EnqueueOptions{
			Retry: true,
			At:    float64(at) / workers.NanoSecondPrecision,
		})
}

// ScheduleCreateBatchesFromFiltersJob schedules a new CreateBatchesWorker job
func (w *Worker) ScheduleCreateBatchesFromFiltersJob(jobID *[]string, at int64) (string, error) {
	return workers.EnqueueWithOptions(
		"create_batches_from_filters_worker",
		"Add",
		jobID,
		workers.EnqueueOptions{
			Retry: true,
			At:    float64(at) / workers.NanoSecondPrecision,
		})
}

// ScheduleProcessBatchJob schedules a new ProcessBatchWorker job
func (w *Worker) ScheduleProcessBatchJob(jobID string, appName string, users *[]User, at int64) (string, error) {
	compressedUsers, err := CompressUsers(users)
	if err != nil {
		return "", err
	}
	return workers.EnqueueWithOptions(
		"process_batch_worker",
		"Add",
		[]interface{}{jobID, appName, compressedUsers},
		workers.EnqueueOptions{
			At: float64(at) / workers.NanoSecondPrecision,
		})
}

// ScheduleJobCompletedJob schedules a new JobCompletedWorker job
func (w *Worker) ScheduleJobCompletedJob(jobID string, at int64) (string, error) {
	maxRetries := w.Config.GetInt("workers.jobCompleted.maxRetries")
	return workers.EnqueueWithOptions(
		"job_completed_worker",
		"Add",
		[]interface{}{jobID},
		workers.EnqueueOptions{
			Retry:      true,
			RetryCount: maxRetries,
			At:         float64(at) / workers.NanoSecondPrecision,
		})
}

// Start starts the worker
func (w *Worker) Start() {
	jobsStatsPort := w.Config.GetInt("workers.statsPort")
	go workers.StatsServer(jobsStatsPort)
	workers.Run()
}
