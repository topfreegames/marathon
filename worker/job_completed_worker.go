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
	goworkers2 "github.com/digitalocean/go-workers2"
	"io"

	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameJobCompleted = "job_completed_worker"

// JobCompletedWorker is the JobCompletedWorker struct
type JobCompletedWorker struct {
	Workers *Worker
	Logger  zap.Logger
}

// NewJobCompletedWorker gets a new JobCompletedWorker
func NewJobCompletedWorker(workers *Worker) *JobCompletedWorker {
	b := &JobCompletedWorker{
		Logger:  workers.Logger.With(zap.String("worker", "JobCompletedWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Configured JobCompletedWorker successfully.")
	return b
}

func (b *JobCompletedWorker) flushControlGroup(job *model.Job) {
	hash := job.ID.String()
	hash = fmt.Sprintf("%s-CONTROL", hash)
	controlGroup, err := b.Workers.RedisClient.LRange(hash, 0, -1).Result()
	b.checkErr(job, err)

	folder := b.Workers.Config.GetString("s3.controlGroupFolder")
	csvBuffer := &bytes.Buffer{}
	csvWriter := io.Writer(csvBuffer)
	csvWriter.Write([]byte("controlGroupUserIds\n"))
	for _, user := range controlGroup {
		csvWriter.Write([]byte(fmt.Sprintf("%s\n", user)))
	}

	bucket := b.Workers.Config.GetString("s3.bucket")
	writePath := fmt.Sprintf("%s/%s/job-%s.csv", bucket, folder, job.ID.String())
	csvBytes := csvBuffer.Bytes()
	_, err = b.Workers.S3Client.PutObject(writePath, &csvBytes)
	b.checkErr(job, err)
	b.updateJobControlGroupCSVPath(job, writePath)

	err = b.Workers.RedisClient.Del(hash).Err()
	b.checkErr(job, err)
}

func (b *JobCompletedWorker) updateJobControlGroupCSVPath(job *model.Job, csvPath string) {
	job.ControlGroupCSVPath = csvPath
	_, err := b.Workers.MarathonDB.Model(job).Set("control_group_csv_path = ?control_group_csv_path").Update()
	b.checkErr(job, err)
}

// Process processes the messages sent to worker queue
func (b *JobCompletedWorker) Process(message *goworkers2.Msg) error {
	arr, err := message.Args().Array()
	checkErr(b.Logger, err)
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(b.Logger, err)
	l := b.Logger.With(
		zap.String("jobID", id.String()),
		zap.String("worker", nameJobCompleted),
	)
	log.I(l, "starting")

	job, err := b.Workers.GetJob(id)
	checkErr(l, err)

	b.Workers.Statsd.Incr(JobCompletedWorkerStart, job.Labels(), 1)

	job.TagRunning(b.Workers.MarathonDB, nameJobCompleted, "starting")

	if b.Workers.SendgridClient != nil {
		err = email.SendJobCompletedEmail(b.Workers.SendgridClient, job, job.App.Name)
		b.checkErr(job, err)
	}

	job.TagRunning(b.Workers.MarathonDB, nameJobCompleted, "sending control group")
	b.flushControlGroup(job)

	job.TagSuccess(b.Workers.MarathonDB, nameJobCompleted, "finished")
	b.Workers.Statsd.Incr(JobCompletedWorkerCompleted, job.Labels(), 1)

	log.I(l, "finished")

	return nil
}

func (b *JobCompletedWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameJobCompleted, err.Error())
		b.Workers.Statsd.Incr(JobCompletedWorkerError, job.Labels(), 1)

		checkErr(b.Logger, err)
	}
}
