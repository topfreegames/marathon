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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	goworkers2 "github.com/digitalocean/go-workers2"
	raven "github.com/getsentry/raven-go"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	redis "gopkg.in/redis.v5"
)

// Worker is the struct that will configure workers
type Worker struct {
	Logger                    zap.Logger
	PushDB                    interfaces.DB
	MarathonDB                interfaces.DB
	Config                    *viper.Viper
	DBPageSize                int
	S3Client                  interfaces.S3
	PageProcessingConcurrency int
	Statsd                    *statsd.Client
	RedisClient               *redis.Client
	ConfigPath                string
	SendgridClient            *extensions.SendgridClient
	Kafka                     interfaces.PushProducer

	Manager *goworkers2.Manager
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
	connection, err := extensions.NewPGClient("push.db", w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.PushDB = connection.DB
}

func (w *Worker) configureMarathonDatabase() {
	connection, err := extensions.NewPGClient("db", w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.MarathonDB = connection.DB
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
	redisDatabase := w.Config.GetInt("workers.redis.db")
	redisPassword := w.Config.GetString("workers.redis.pass")
	redisPoolsize := w.Config.GetInt("workers.redis.poolSize")
	tlsEnabled := w.Config.GetBool("workers.redis.tlsEnabled")

	logger := w.Logger.With(
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDB", redisDatabase),
		zap.Int("redisPoolsize", redisPoolsize),
	)

	logger.Info("connecting to workers redis")

	// unique process id for this instance of workers (for recovery of inprogress jobs on crash)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	r, err := extensions.NewRedis("workers", w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.RedisClient = r

	opt := goworkers2.Options{
		ServerAddr: fmt.Sprintf("%s:%d", redisHost, redisPort),
		Database:   redisDatabase,
		PoolSize:   redisPoolsize,
		ProcessID:  hostname,
		Password:   redisPassword,
	}
	if tlsEnabled {
		opt.RedisTLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	manager, err := goworkers2.NewManager(opt)
	checkErr(w.Logger, err)

	w.Manager = manager
}

func (w *Worker) configureS3Client() {
	s3Client, err := extensions.NewS3(w.Config, w.Logger)
	checkErr(w.Logger, err)
	w.S3Client = s3Client
}

func (w *Worker) configureWorkers() {
	p := NewProcessBatchWorker(w)
	k := NewCSVSplitWorker(w)
	c := NewCreateBatchesWorker(w)
	r := NewResumeJobWorker(w)
	j := NewJobCompletedWorker(w)
	directWorker := NewDirectWorker(w)

	createCSVSplitWorkerConcurrency := w.Config.GetInt("workers.csvSplitWorker.concurrency")
	processBatchWorkerConcurrency := w.Config.GetInt("workers.processBatch.concurrency")
	resumeJobWorkerConcurrency := w.Config.GetInt("workers.resume.concurrency")
	jobCompletedWorkerConcurrency := w.Config.GetInt("workers.jobCompleted.concurrency")
	createBatchesWorkerConcurrency := w.Config.GetInt("workers.createBatches.concurrency")

	jobDirectWorkerConcurrency := w.Config.GetInt("workers.direct.concurrency")

	w.Manager.AddWorker("csv_split_worker", createCSVSplitWorkerConcurrency, k.Process)
	w.Manager.AddWorker("create_batches_worker", createBatchesWorkerConcurrency, c.Process)
	w.Manager.AddWorker("process_batch_worker", processBatchWorkerConcurrency, p.Process)
	w.Manager.AddWorker("resume_job_worker", resumeJobWorkerConcurrency, r.Process)
	w.Manager.AddWorker("job_completed_worker", jobCompletedWorkerConcurrency, j.Process)
	w.Manager.AddWorker("direct_worker", jobDirectWorkerConcurrency, directWorker.Process)
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

// CreateCSVSplitJob creates a new CSVSplitWorker job
func (w *Worker) CreateCSVSplitJob(job *model.Job) (string, error) {
	maxRetries := w.Config.GetInt("workers.csvSplitWorker.maxRetries")
	producer := w.Manager.Producer()

	return producer.EnqueueWithOptions(
		"csv_split_worker",
		"Add",
		job.ID.String(),
		goworkers2.EnqueueOptions{
			Retry:      true,
			RetryCount: maxRetries,
		})
}

// ScheduleCSVSplitJob schedules a new CSVSplitWorker job
func (w *Worker) ScheduleCSVSplitJob(job *model.Job, at int64) (string, error) {
	maxRetries := w.Config.GetInt("workers.csvSplitWorker.maxRetries")
	producer := w.Manager.Producer()
	return producer.EnqueueWithOptions(
		"csv_split_worker",
		"Add",
		job.ID.String(),
		goworkers2.EnqueueOptions{
			Retry:      true,
			RetryCount: maxRetries,
			At:         float64(at) / goworkers2.NanoSecondPrecision,
		})
}

// CreateBatchesJob creates a new CreateBatchesWorker job
func (w *Worker) CreateBatchesJob(part *BatchPart) (string, error) {
	maxRetries := w.Config.GetInt("workers.createBatches.maxRetries")
	producer := w.Manager.Producer()
	return producer.EnqueueWithOptions("create_batches_worker", "Add", part, goworkers2.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// CreateDirectBatchesJob schedules a new DirectWorker job
func (w *Worker) CreateDirectBatchesJob(job *model.Job) error {
	maxRetries := w.Config.GetInt("workers.direct.maxRetries")
	return w.createDirectBatchesJobWithOption(job, goworkers2.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// ScheduleDirectBatchesJob schedules a new DirectWorker job
func (w *Worker) ScheduleDirectBatchesJob(job *model.Job, at int64) error {
	maxRetries := w.Config.GetInt("workers.direct.maxRetries")
	return w.createDirectBatchesJobWithOption(job, goworkers2.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
		At:         float64(at) / goworkers2.NanoSecondPrecision,
	})
}

func (w *Worker) createDirectBatchesJobWithOption(job *model.Job, options goworkers2.EnqueueOptions) error {
	var testBatchSize uint64
	var maxSeqID uint64
	var rownsEstimative uint64
	var i uint64

	job.GetJobInfoAndApp(w.MarathonDB)
	tableName := GetPushDBTableName(job.App.Name, job.Service)
	query := fmt.Sprintf("SELECT reltuples::BIGINT AS estimate FROM pg_class WHERE relname = '%s';", tableName)
	_, err := w.PushDB.QueryOne(&rownsEstimative, query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("SELECT max(seq_id) FROM %s;", tableName)
	_, err = w.PushDB.QueryOne(&maxSeqID, query)
	if err != nil {
		return err
	}
	if rownsEstimative == 0 {
		rownsEstimative = 1
	}

	//testBatchSize = (200000 * maxSeqID) / rownsEstimative
	testBatchSize = 100000

	producer := w.Manager.Producer()

	for i = 0; i < maxSeqID+1; {
		_, err = producer.EnqueueWithOptions("direct_worker", "Add",
			DirectPartMsg{
				SmallestSeqID: i,
				BiggestSeqID:  i + testBatchSize,
				JobUUID:       job.ID,
			}, options)
		if err != nil {
			return err
		}
		i += testBatchSize
	}

	_, err = w.MarathonDB.Model(job).Set("total_tokens = ?", rownsEstimative).Where("id = ?", job.ID).Update()
	if err != nil {
		return err
	}

	batches := i / testBatchSize
	_, err = w.MarathonDB.Model(job).Set("total_batches = ?", batches).Where("id = ?", job.ID).Update()
	if err != nil {
		return err
	}

	return nil
}

// CreateProcessBatchJob creates a new ProcessBatchWorker job
func (w *Worker) CreateProcessBatchJob(jobID string, appName string, users *[]User) (string, error) {
	compressedUsers, err := CompressUsers(users)
	if err != nil {
		return "", err
	}
	producer := w.Manager.Producer()
	return producer.Enqueue(
		"process_batch_worker",
		"Add",
		[]interface{}{jobID, appName, compressedUsers},
	)
}

// CreateResumeJob creates a new ResumeJobWorker job
func (w *Worker) CreateResumeJob(jobID *[]string) (string, error) {
	maxRetries := w.Config.GetInt("workers.resume.maxRetries")
	producer := w.Manager.Producer()
	return producer.EnqueueWithOptions("resume_job_worker", "Add", jobID, goworkers2.EnqueueOptions{
		Retry:      true,
		RetryCount: maxRetries,
	})
}

// ScheduleCreateBatchesJob schedules a new CreateBatchesWorker job
func (w *Worker) ScheduleCreateBatchesJob(jobID *[]string, at int64) (string, error) {
	producer := w.Manager.Producer()

	return producer.EnqueueWithOptions(
		"create_batches_worker",
		"Add",
		jobID,
		goworkers2.EnqueueOptions{
			Retry: true,
			At:    float64(at) / goworkers2.NanoSecondPrecision,
		})
}

// ScheduleCreateBatchesFromFiltersJob schedules a new CreateBatchesWorker job
func (w *Worker) ScheduleCreateBatchesFromFiltersJob(jobID *[]string, at int64) (string, error) {
	producer := w.Manager.Producer()
	return producer.EnqueueWithOptions(
		"create_batches_from_filters_worker",
		"Add",
		jobID,
		goworkers2.EnqueueOptions{
			Retry: true,
			At:    float64(at) / goworkers2.NanoSecondPrecision,
		})
}

// ScheduleProcessBatchJob schedules a new ProcessBatchWorker job
func (w *Worker) ScheduleProcessBatchJob(jobID string, appName string, users *[]User, at int64) (string, error) {
	compressedUsers, err := CompressUsers(users)
	if err != nil {
		return "", err
	}
	producer := w.Manager.Producer()
	return producer.EnqueueWithOptions(
		"process_batch_worker",
		"Add",
		[]interface{}{jobID, appName, compressedUsers},
		goworkers2.EnqueueOptions{
			At: float64(at) / goworkers2.NanoSecondPrecision,
		})
}

// ScheduleJobCompletedJob schedules a new JobCompletedWorker job
func (w *Worker) ScheduleJobCompletedJob(jobID string, at int64) (string, error) {
	maxRetries := w.Config.GetInt("workers.jobCompleted.maxRetries")
	producer := w.Manager.Producer()

	return producer.EnqueueWithOptions(
		"job_completed_worker",
		"Add",
		[]interface{}{jobID},
		goworkers2.EnqueueOptions{
			Retry:      true,
			RetryCount: maxRetries,
			At:         float64(at) / goworkers2.NanoSecondPrecision,
		})
}

// Start starts the worker
func (w *Worker) Start() {
	jobsStatsPort := w.Config.GetInt("workers.statsPort")
	go func() {
		http.HandleFunc("/stats", func(rw http.ResponseWriter, req *http.Request) {

			_, marathonError := w.MarathonDB.Exec("SELECT 1")
			_, pushError := w.MarathonDB.Exec("SELECT 1")
			pong, redisError := w.RedisClient.Ping().Result()

			status := struct {
				MarathonHealthy bool `json:"marathon_db_healthy"`
				PushHealthy     bool `json:"push_db_healthy"`
				RedisHealthy    bool `json:"redis_healthy"`
			}{
				MarathonHealthy: marathonError == nil,
				PushHealthy:     pushError == nil,
				RedisHealthy:    redisError == nil && pong == "PONG",
			}

			if !status.MarathonHealthy || !status.PushHealthy || !status.RedisHealthy {
				rw.WriteHeader(http.StatusServiceUnavailable)
			}
			json.NewEncoder(rw).Encode(status)
		})
		if err := http.ListenAndServe(fmt.Sprint(":", jobsStatsPort), nil); err != nil {
			panic(err)
		}
	}()
	go w.Manager.Run()
}

// SendControlGroupToRedis send a sequency of users ids to redis
func (w *Worker) SendControlGroupToRedis(job *model.Job, ids []string) {
	start := time.Now()
	hash := job.ID.String()
	var args []interface{}
	for _, id := range ids {
		args = append(args, id)
	}
	w.RedisClient.LPush(fmt.Sprintf("%s-CONTROL", hash), args...).Result()
	w.Statsd.Timing("save_control_group", time.Now().Sub(start), job.Labels(), 1)
}

// GetJob get a job from the db
func (w *Worker) GetJob(jobID uuid.UUID) (*model.Job, error) {
	job := model.Job{
		ID: jobID,
	}
	err := job.GetJobInfoAndApp(w.MarathonDB)
	return &job, err
}
