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
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// BatchPart  hold the information of a batch
type BatchPart struct {
	Start      int
	Size       int
	TotalParts int
	TotalSize  int
	Part       int
	Job        model.Job
}

const nameSCVSplit = "csv_split_worker"

const partSize = 10 * 1024 * 1024

// const partSize = 1024 * 5

// CSVSplitWorker is the CSVSplitWorker struct
type CSVSplitWorker struct {
	Workers *Worker
	Logger  zap.Logger
}

// NewCSVSplitWorker gets a new CSVSplitWorker
func NewCSVSplitWorker(workers *Worker) *CSVSplitWorker {
	b := &CSVSplitWorker{
		Logger:  workers.Logger.With(zap.String("worker", "CSVSplitWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Workers.Configured CSVSplitWorker successfully.")
	return b
}

// Process processes the messages sent to batch worker queue
func (b *CSVSplitWorker) Process(message *workers.Msg) {
	var id uuid.UUID
	err := json.Unmarshal([]byte(message.Args().ToJson()), &id)
	checkErr(b.Logger, err)

	isReexecution := checkIsReexecution(id, b.Workers.RedisClient, b.Logger)
	l := b.Logger.With(
		zap.String("jobID", id.String()),
		zap.Bool("isReexecution", isReexecution),
		zap.String("worker", nameSCVSplit),
	)
	log.I(l, "starting")
	job := &model.Job{
		ID: id,
	}
	job.TagRunning(b.Workers.MarathonDB, nameSCVSplit, "starting")
	err = b.Workers.MarathonDB.DB.Model(job).Column("job.*", "App").Where("job.id = ?", job.ID).Select()
	checkErr(l, err)
	if job.Status == stoppedJobStatus {
		l.Info("stopped job")
		return
	}

	// get file information
	resp, err := b.Workers.S3Client.DownloadChunk(0, 7, job.CSVPath)
	checkErr(l, err)
	parts := strings.Split(*resp.ContentRange, "/")
	totalSize, err := strconv.Atoi(parts[1])
	checkErr(l, err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	firtsBytes := buf.String()

	if strings.Compare("userIds", firtsBytes) != 0 {
		checkErr(l, errors.New("Invalid CSV - it does not contain column"))
	}

	start := 0
	totalParts := int(math.Ceil(float64(totalSize) / float64(partSize)))
	labels := []string{fmt.Sprintf("game:%s", job.App.Name), fmt.Sprintf("platform:%s", job.Service)}

	for i := 0; i < totalParts; i++ {
		size := totalSize - start
		if size > partSize {
			size = partSize
		}
		_, err := b.Workers.CreateBatchesJob(&BatchPart{
			Start:      start,
			Size:       size,
			TotalParts: totalParts,
			TotalSize:  totalSize,
			Part:       i,
			Job:        *job,
		})
		checkErr(l, err)
		start += size
		b.Workers.Statsd.Incr("csv_job_part", labels, 1)
	}
	job.TagSuccess(b.Workers.MarathonDB, nameSCVSplit, "finished")
}

func (b *CSVSplitWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameSCVSplit, err.Error())
		checkErr(b.Logger, err)
	}
}
