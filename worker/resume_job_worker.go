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

	"gopkg.in/redis.v5"

	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

// ResumeJobWorker is the CreateBatchesUsingFiltersWorker struct
type ResumeJobWorker struct {
	Logger  zap.Logger
	Workers *Worker
}

// NewResumeJobWorker gets a new ResumeJobWorker
func NewResumeJobWorker(workers *Worker) *ResumeJobWorker {
	b := &ResumeJobWorker{
		Logger:  workers.Logger.With(zap.String("worker", "ResumeJobWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Configured ResumeJobWorker successfully.")
	return b
}

// Process processes the messages sent to worker queue
func (b *ResumeJobWorker) Process(message *workers.Msg) {
	arr, err := message.Args().Array()
	checkErr(b.Logger, err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(b.Logger, err)
	l := b.Logger.With(
		zap.String("jobID", id.String()),
	)
	log.I(l, "starting resume_job_worker")

	job, err := b.Workers.GetJob(id)
	checkErr(l, err)
	if job.Status == stoppedJobStatus {
		l.Info("stopped job resume_job_worker")
		err := b.Workers.RedisClient.Del(fmt.Sprintf("%s-pausedjobs", jobID.(string))).Err()
		if err != nil && err != redis.Nil {
			checkErr(b.Logger, err)
		}
		return
	}

	for {
		batchInfo, err := b.Workers.RedisClient.RPop(fmt.Sprintf("%s-pausedjobs", jobID.(string))).Result()
		if err != nil && err == redis.Nil {
			break
		}
		checkErr(b.Logger, err)
		pausedJobArgs, err := workers.NewMsg(batchInfo)
		checkErr(b.Logger, err)
		pausedJobArr, err := pausedJobArgs.Args().Array()
		checkErr(b.Logger, err)
		parsed, err := ParseProcessBatchWorkerMessageArray(pausedJobArr)
		checkErr(b.Logger, err)
		_, err = b.Workers.CreateProcessBatchJob(parsed.JobID.String(), parsed.AppName, &parsed.Users)
		checkErr(l, err)
	}

	log.I(b.Logger, "finished resume_job_worker")
}
