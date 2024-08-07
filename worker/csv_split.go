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

/*** Description ***

 */

package worker

import (
	"encoding/json"
	goworkers2 "github.com/digitalocean/go-workers2"
	"math"

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
func (b *CSVSplitWorker) Process(message *goworkers2.Msg) error {
	var id uuid.UUID
	csvSizeLimit := b.getCSVSizeLimitBytes()
	err := json.Unmarshal([]byte(message.Args().ToJson()), &id)
	checkErr(b.Logger, err)

	isReexecution := checkIsReexecution(id, b.Workers.RedisClient, b.Logger)
	l := b.Logger.With(
		zap.String("jobID", id.String()),
		zap.Bool("isReexecution", isReexecution),
		zap.String("worker", nameSCVSplit),
	)
	log.I(l, "starting")

	job, err := b.Workers.GetJob(id)
	checkErr(l, err)
	l = l.With(
		zap.String("appID", job.AppID.String()),
	)
	l.Debug("job found")

	job.TagRunning(b.Workers.MarathonDB, nameSCVSplit, "starting")
	b.Workers.Statsd.Incr(CsvSplitWorkerStart, job.Labels(), 1)

	if job.Status == stoppedJobStatus {
		l.Info("stopped job")
		b.Workers.Statsd.Incr(CsvSplitWorkerCompleted, job.Labels(), 1)
		return nil
	}

	// get file information
	totalSize, _, err := b.Workers.S3Client.DownloadChunk(0, 1, job.CSVPath)
	b.checkErr(job, err)

	start := 0
	totalParts := int(math.Ceil(float64(totalSize) / csvSizeLimit))

	for i := 0; i < totalParts; i++ {
		size := totalSize - start
		if size > int(math.Ceil(csvSizeLimit)) {
			size = int(math.Ceil(csvSizeLimit))
		}
		_, err := b.Workers.CreateBatchesJob(&BatchPart{
			Start:      start,
			Size:       size,
			TotalParts: totalParts,
			TotalSize:  totalSize,
			Part:       i,
			Job:        *job,
		})
		b.checkErr(job, err)
		start += size
		b.Workers.Statsd.Incr("csv_job_part", job.Labels(), 1)
	}

	job.TagSuccess(b.Workers.MarathonDB, nameSCVSplit, "finished")
	b.Workers.Statsd.Incr(CsvSplitWorkerCompleted, job.Labels(), 1)
	l.Info("finished")

	return nil
}

func (b *CSVSplitWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameSCVSplit, err.Error())
		b.Workers.Statsd.Incr(CsvSplitWorkerError, job.Labels(), 1)

		checkErr(b.Logger, err)
	}
}

func (b *CSVSplitWorker) getCSVSizeLimitBytes() float64 {
	b.Workers.Config.SetDefault("workers.csvSplitWorker.csvSizeLimitMB", 10)
	csvSizeLimitMB := b.Workers.Config.GetFloat64("workers.csvSplitWorker.csvSizeLimitMB")

	return csvSizeLimitMB * 1024 * 1024
}
