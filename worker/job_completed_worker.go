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

// Process processes the messages sent to worker queue
func (b *JobCompletedWorker) Process(message *workers.Msg) {
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

	job := &model.Job{
		ID: id,
	}
	err = b.Workers.MarathonDB.DB.Model(job).Where("job.id = ?", job.ID).Select()
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameJobCompleted, "finished")
		checkErr(l, err)
	}
	job.TagRunning(b.Workers.MarathonDB, nameJobCompleted, "starting")
	err = b.Workers.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	b.checkErr(job, err)

	if b.Workers.SendgridClient != nil {
		err = email.SendJobCompletedEmail(b.Workers.SendgridClient, job, job.App.Name)
		b.checkErr(job, err)
	}
	job.TagSuccess(b.Workers.MarathonDB, nameJobCompleted, "finished")
	log.I(l, "finished")
}

func (b *JobCompletedWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameJobCompleted, err.Error())
		checkErr(b.Logger, err)
	}
}
