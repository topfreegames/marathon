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
	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Worker is the struct that will configure workers
type Worker struct {
	Debug  bool
	Logger zap.Logger
}

// GetWorker returns a configured worker
func GetWorker(debug bool, l zap.Logger) *Worker {
	worker := &Worker{
		Debug:  debug,
		Logger: l,
	}

	worker.configure()
	return worker
}

func (w *Worker) configure() {
	w.loadConfigurationDefaults()
	w.configureRedis()
	w.configureWorkers()
}

func (w *Worker) loadConfigurationDefaults() {
	viper.SetDefault("workers.redis.server", "localhost:6379")
	viper.SetDefault("workers.redis.database", "0")
	viper.SetDefault("workers.redis.poolSize", "10")
	viper.SetDefault("workers.statsPort", 8081)
	viper.SetDefault("workers.concurrency", 10)
	viper.SetDefault("database.url", "postgres://localhost:5432/marathon?sslmode=disable")
}

func (w *Worker) configureRedis() {
	redisServer := viper.GetString("workers.redis.server")
	redisDatabase := viper.GetString("workers.redis.database")
	redisPoolsize := viper.GetString("workers.redis.poolSize")
	// unique process id for this instance of workers (for recovery of inprogress jobs on crash)
	redisProcessID := uuid.NewV4()

	workers.Configure(map[string]string{
		"server":   redisServer,
		"database": redisDatabase,
		"pool":     redisPoolsize,
		"process":  redisProcessID.String(),
	})
}

func (w *Worker) configureWorkers() {
	jobsConcurrency := viper.GetInt("workers.concurrency")
	b := GetBenchmarkWorker(viper.GetString("workers.redis.server"), viper.GetString("workers.redis.database"))
	c := GetCreateBatchesWorker(viper.GetViper())
	//workers.Middleware.Append(&) TODO
	workers.Process("benchmark_worker", b.Process, jobsConcurrency)
	workers.Process("create_batches_worker", c.Process, jobsConcurrency)
}

// CreateBatchesJob creates a new CreateBatchesWorker job
func (w *Worker) CreateBatchesJob(jobID *[]string) (string, error) {
	return workers.Enqueue("create_batches_worker", "Add", jobID)
}

// Start starts the worker
func (w *Worker) Start() {
	jobsStatsPort := viper.GetInt("workers.statsPort")
	go workers.StatsServer(jobsStatsPort)
	workers.Run()
}
