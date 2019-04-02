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
	"sort"
	"time"

	"encoding/json"
	"github.com/aws/aws-sdk-go/service/s3"
	workers "github.com/jrallison/go-workers"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

const nameDbToCsv = "db_to_csv_worker"

// ToCSVMenssage store th information needed to retrive part of data from the db
// and save it to the csv
type ToCSVMenssage struct {
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

func (b *DbToCsvWorker) getPageFromDBWith(message *ToCSVMenssage) *[]User {
	var users []User
	start := time.Now()
	_, err := b.Workers.PushDB.DB.Query(&users, message.Query)
	labels := []string{fmt.Sprintf("game:%s", message.Job.App.Name), fmt.Sprintf("platform:%s", message.Job.Service)}
	b.Workers.Statsd.Timing("get_page_with_filters", time.Now().Sub(start), labels, 1)
	checkErr(b.Logger, err)
	return &users
}

// return if is the last element or not
func (b *DbToCsvWorker) redisSaveCompletedPart(message *ToCSVMenssage, etag string) bool {
	hash := message.Job.ID.String()
	data, err := json.Marshal(CompletedPart{
		PartNumber: int64(message.PartNumber),
		ETag:       etag,
	})
	queueLen, err := b.Workers.RedisClient.LPush(hash, data).Result()
	checkErr(b.Logger, err)

	return queueLen == int64(message.TotalJobs)
}

func (b *DbToCsvWorker) finishUploads(message *ToCSVMenssage) {
	hash := message.Job.ID.String()
	strings, err := b.Workers.RedisClient.LRange(hash, 0, -1).Result()
	checkErr(b.Logger, err)

	completeParts := make(completedParts, message.TotalJobs)
	for i, string := range strings {
		var tmpPart CompletedPart
		err := json.Unmarshal([]byte(string), &tmpPart)
		checkErr(b.Logger, err)
		completeParts[i] = &s3.CompletedPart{
			ETag:       &tmpPart.ETag,
			PartNumber: &tmpPart.PartNumber,
		}
	}
	err = b.Workers.RedisClient.Del(hash).Err()
	checkErr(b.Logger, err)

	// Parts must be sorted in PartNumber order.
	sort.Sort(completeParts)

	err = b.Workers.S3Client.CompleteMultipartUpload(&message.Uploader, completeParts)
	checkErr(b.Logger, err)

}

func (b *DbToCsvWorker) createBatchesJob(message *ToCSVMenssage) {
	jid, err := b.Workers.CreateBatchesJob(&[]string{message.Job.ID.String()})
	checkErr(b.Logger, err)
	b.Logger.Info("created create batches job", zap.String("jid", jid))
}

func (b *DbToCsvWorker) uploadPart(users *[]User, message *ToCSVMenssage) {
	labels := []string{fmt.Sprintf("game:%s", message.Job.App.Name), fmt.Sprintf("platform:%s", message.Job.Service)}
	b.Workers.Statsd.Incr("write_user_page_into_csv", labels, 1)

	buffer := bytes.NewBufferString("")

	// if is the first element
	if message.PartNumber == 1 {
		message.Job.TagRunning(b.Workers.MarathonDB, nameDbToCsv, "starting")
		buffer.WriteString("userIds\n")
	}

	for _, user := range *users {
		buffer.WriteString(fmt.Sprintf("%s\n", user.UserID))
	}

	tmpPart := int64(message.PartNumber)
	completePart, err := b.Workers.S3Client.UploadPart(buffer, &message.Uploader, tmpPart)
	checkErr(b.Logger, err)

	if b.redisSaveCompletedPart(message, *completePart.ETag) {
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

	var msg ToCSVMenssage
	data := message.Args().ToJson()
	err := json.Unmarshal([]byte(data), &msg)
	checkErr(b.Logger, err)

	users := b.getPageFromDBWith(&msg)
	b.uploadPart(users, &msg)

	l.Info("finished")
}

func (b *DbToCsvWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameDbToCsv, err.Error())
		checkErr(b.Logger, err)
	}
}
