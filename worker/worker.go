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

	"github.com/jrallison/go-workers"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Worker is the struct that will configure workers
type Worker struct {
	Debug      bool
	Logger     zap.Logger
	ConfigPath string
	Config     *viper.Viper
}

// NewWorker returns a configured worker
func NewWorker(debug bool, l zap.Logger, configPath string) *Worker {
	worker := &Worker{
		Debug:      debug,
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
	w.configureRedis()
	w.configureWorkers()
}

func (w *Worker) loadConfigurationDefaults() {
	w.Config.SetDefault("workers.redis.server", "localhost:6379")
	w.Config.SetDefault("workers.redis.database", "0")
	w.Config.SetDefault("workers.redis.poolSize", "10")
	w.Config.SetDefault("workers.statsPort", 8081)
	w.Config.SetDefault("workers.concurrency", 10)
	w.Config.SetDefault("database.url", "postgres://localhost:5432/marathon?sslmode=disable")
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

}

func (w *Worker) configureWorkers() {
	p := NewProcessBatchWorker(w.Config, w.Logger)
	c := NewCreateBatchesWorker(w.Config, w.Logger, w)
	f := NewCreateBatchesFromFiltersWorker(w.Config, w.Logger, w)
	createBatchesWorkerConcurrency := w.Config.GetInt("workers.createBatches.concurrency")
	createBatchesFromFiltersWorkerConcurrency := w.Config.GetInt("workers.createBatchesFromFilters.concurrency")
	processBatchWorkerConcurrency := w.Config.GetInt("workers.processBatch.concurrency")
	workers.Process("create_batches_worker", c.Process, createBatchesWorkerConcurrency)
	workers.Process("process_batch_worker", p.Process, processBatchWorkerConcurrency)
	workers.Process("create_batches_from_filters_worker", f.Process, createBatchesFromFiltersWorkerConcurrency)
}

// CreateBatchesJob creates a new CreateBatchesWorker job
func (w *Worker) CreateBatchesJob(jobID *[]string) (string, error) {
	maxRetries := w.Config.GetInt("workers.createBatches.maxRetries")
	return workers.EnqueueWithOptions("create_batches_worker", "Add", jobID, workers.EnqueueOptions{
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
func (w *Worker) CreateProcessBatchJob(jobID string, appName string, users []User) (string, error) {
	return workers.Enqueue("process_batch_worker", "Add", []interface{}{jobID, appName, users})
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
func (w *Worker) ScheduleProcessBatchJob(jobID string, appName string, users []User, at int64) (string, error) {
	return workers.EnqueueWithOptions(
		"process_batch_worker",
		"Add",
		[]interface{}{jobID, appName, users},
		workers.EnqueueOptions{
			At: float64(at) / workers.NanoSecondPrecision,
		})
}

// Start starts the worker
func (w *Worker) Start() {
	jobsStatsPort := w.Config.GetInt("workers.statsPort")
	go workers.StatsServer(jobsStatsPort)
	workers.Run()
}
