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
	"github.com/topfreegames/marathon/email"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// JobCompletedWorker is the JobCompletedWorker struct
type JobCompletedWorker struct {
	Logger         zap.Logger
	MarathonDB     *extensions.PGClient
	Config         *viper.Viper
	SendgridClient *extensions.SendgridClient
}

// NewJobCompletedWorker gets a new JobCompletedWorker
func NewJobCompletedWorker(config *viper.Viper, logger zap.Logger) *JobCompletedWorker {
	b := &JobCompletedWorker{
		Config: config,
		Logger: logger.With(zap.String("worker", "JobCompletedWorker")),
	}
	b.configure()
	log.D(logger, "Configured JobCompletedWorker successfully.")
	return b
}

func (b *JobCompletedWorker) configureMarathonDatabase() {
	var err error
	b.MarathonDB, err = extensions.NewPGClient("db", b.Config, b.Logger)
	checkErr(b.Logger, err)
}

func (b *JobCompletedWorker) configureSendgrid() {
	apiKey := b.Config.GetString("sendgrid.key")
	if apiKey != "" {
		b.SendgridClient = extensions.NewSendgridClient(b.Config, b.Logger, apiKey)
	}
}

func (b *JobCompletedWorker) configure() {
	b.configureMarathonDatabase()
	b.configureSendgrid()
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
	)
	log.I(l, "starting job_completed_worker")

	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)

	if b.SendgridClient != nil {
		err = email.SendJobCompletedEmail(b.SendgridClient, job, job.App.Name)
		checkErr(l, err)
	}

	log.I(l, "finished job_completed_worker")
}
