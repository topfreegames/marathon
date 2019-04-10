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
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	workers "github.com/jrallison/go-workers"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameDbToCsv = "db_to_csv_worker"

// ToCSVMessage store th information needed to retrive part of data from the db
// and save it to the csv
type ToCSVMessage struct {
	Query      string
	Uploader   s3.CreateMultipartUploadOutput
	PartNumber int
	Job        model.Job
	TotalJobs  int
}

// CompletedPart store information related to one completed part
type CompletedPart struct {
	ETag       string
	PartNumber int64
}

type completedParts []*s3.CompletedPart

func (a completedParts) Len() int           { return len(a) }
func (a completedParts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a completedParts) Less(i, j int) bool { return *a[i].PartNumber < *a[j].PartNumber }

// DbToCsvWorker is the CreateBatchesUsingFiltersWorker struct
type DbToCsvWorker struct {
	Logger  zap.Logger
	Workers *Worker
}

// NewDbToCsvWorker gets a new DbToCsvWorker
func NewDbToCsvWorker(workers *Worker) *DbToCsvWorker {
	b := &DbToCsvWorker{
		Logger:  workers.Logger.With(zap.String("worker", "DbToCsvWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Configured DbToCsvWorker successfully")
	return b
}

func (b *DbToCsvWorker) getPageFromDBWith(message *ToCSVMessage) *[]string {
	var users []string
	start := time.Now()
	_, err := b.Workers.PushDB.DB.Query(&users, message.Query)
	b.Workers.Statsd.Timing("get_page_with_filters", time.Now().Sub(start), message.Job.Labels(), 1)
	b.checkErr(&message.Job, err)
	return &users
}

// return if is the last element or not
func (b *DbToCsvWorker) multiPartUploadCompleted(message *ToCSVMessage, etag string) bool {
	hash := message.Job.ID.String()
	var queueLen int64

	data, err := json.Marshal(CompletedPart{
		PartNumber: int64(message.PartNumber),
		ETag:       etag,
	})

	if etag == "" {
		queueLen, err = b.Workers.RedisClient.LLen(hash).Result()
	} else {
		queueLen, err = b.Workers.RedisClient.LPush(hash, data).Result()
	}
	b.checkErr(&message.Job, err)
	return queueLen == int64(message.TotalJobs)
}

func (b *DbToCsvWorker) finishUploads(message *ToCSVMessage) {
	hash := message.Job.ID.String()
	strings, err := b.Workers.RedisClient.LRange(hash, 0, -1).Result()
	b.checkErr(&message.Job, err)

	completeParts := make(completedParts, message.TotalJobs)
	for i, string := range strings {
		var tmpPart CompletedPart
		err := json.Unmarshal([]byte(string), &tmpPart)
		b.checkErr(&message.Job, err)
		completeParts[i] = &s3.CompletedPart{
			ETag:       &tmpPart.ETag,
			PartNumber: &tmpPart.PartNumber,
		}
	}
	err = b.Workers.RedisClient.Del(hash).Err()
	b.checkErr(&message.Job, err)

	// Parts must be sorted in PartNumber order.
	sort.Sort(completeParts)
	b.Logger.Info("completed parts", zap.String("parts", fmt.Sprint(completeParts)))

	err = b.Workers.S3Client.CompleteMultipartUpload(&message.Uploader, completeParts)
	b.checkErr(&message.Job, err)
}

func (b *DbToCsvWorker) createBatchesJob(message *ToCSVMessage) {
	jid, err := b.Workers.CSVSplitJob(message.Job.ID.String())
	b.checkErr(&message.Job, err)
	b.Logger.Info("created create batches job", zap.String("jid", jid))
}

func (b *DbToCsvWorker) uploadPart(users *[]string, message *ToCSVMessage) {
	start := time.Now()
	buffer := bytes.NewBufferString("")

	// if is the first element
	if message.PartNumber == 1 {
		message.Job.TagRunning(b.Workers.MarathonDB, nameDbToCsv, "starting")
		buffer.WriteString("userIds\n")
	}

	for _, user := range *users {
		if IsUserIDValid(user) {
			buffer.WriteString(fmt.Sprintf("%s\n", user))
		}
	}

	tmpPart := int64(message.PartNumber)
	completePart, err := b.Workers.S3Client.UploadPart(buffer, &message.Uploader, tmpPart)
	b.Workers.Statsd.Timing("write_user_page_into_csv", time.Now().Sub(start), message.Job.Labels(), 1)
	b.checkErr(&message.Job, err)

	if b.multiPartUploadCompleted(message, *completePart.ETag) {
		b.finishUploads(message)
		b.createBatchesJob(message)
		message.Job.TagSuccess(b.Workers.MarathonDB, nameDbToCsv, "finished")
	}
}

// Process processes the messages sent to worker queue
func (b *DbToCsvWorker) Process(message *workers.Msg) {
	l := b.Logger.With(
		zap.String("worker", nameDbToCsv),
	)
	l.Info("starting")

	var msg ToCSVMessage
	data := message.Args().ToJson()
	err := json.Unmarshal([]byte(data), &msg)
	checkErr(l, err)

	// check if all uploads parts are finished
	if !b.multiPartUploadCompleted(&msg, "") {
		users := b.getPageFromDBWith(&msg)
		b.uploadPart(users, &msg)
	} else {
		b.finishUploads(&msg)
	}

	l.Info("finished")
}

func (b *DbToCsvWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameDbToCsv, err.Error())
		checkErr(b.Logger, err)
	}
}
