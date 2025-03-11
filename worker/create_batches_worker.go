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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	goworkers2 "github.com/digitalocean/go-workers2"

	"github.com/go-pg/pg/v10"

	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	redis "gopkg.in/redis.v5"
)

const nameCreateBatches = "create_batches_worker"
const uuidSizeBytes = 36

// CreateBatchesWorker is the CreateBatchesWorker struct
type CreateBatchesWorker struct {
	Workers *Worker
	Logger  zap.Logger
}

// NewCreateBatchesWorker gets a new CreateBatchesWorker
func NewCreateBatchesWorker(workers *Worker) *CreateBatchesWorker {
	b := &CreateBatchesWorker{
		Logger:  workers.Logger.With(zap.String("worker", "CreateBatchesWorker")),
		Workers: workers,
	}
	b.Logger.Debug("Workers.Configured CreateBatchesWorker successfully.")
	return b
}

// ReadFromCSV reads CSV from S3 and return correspondent array of strings
func (b *CreateBatchesWorker) ReadFromCSV(buffer *[]byte, msg *BatchPart) []string {
	job := &msg.Job
	for i, b := range *buffer {
		if b == 0x0D {
			(*buffer)[i] = 0x0A
		}
	}

	r := csv.NewReader(bytes.NewReader(*buffer))
	lines, err := r.ReadAll()
	b.checkErr(job, err)
	res := []string{}
	for i, line := range lines {
		if i == 0 && msg.Part == 0 {
			continue
		}
		res = append(res, line[0])
	}
	return res
}

func (b *CreateBatchesWorker) updateTotalBatches(totalBatches int, job *model.Job) {
	job.TotalBatches = totalBatches
	// coalesce is necessary since total_batches can be null
	_, err := b.Workers.MarathonDB.Model(job).Set("total_batches = coalesce(total_batches, 0) + ?", totalBatches).Where("id = ?", job.ID).Update()
	b.checkErr(job, err)
}

func (b *CreateBatchesWorker) updateTotalTokens(totalTokens int, job *model.Job) {
	job.TotalTokens = totalTokens
	// coalesce is necessary since total_tokens can be null
	_, err := b.Workers.MarathonDB.Model(job).Set("total_tokens = coalesce(total_tokens, 0) + ?", totalTokens).Where("id = ?", job.ID).Update()
	b.checkErr(job, err)
}

func (b *CreateBatchesWorker) updateCompletedAt(unixTime int64, job *model.Job) {
	_, err := b.Workers.MarathonDB.Model(job).Set("completed_at = ?", unixTime).Where("id = ?", job.ID).Update()
	b.checkErr(job, err)
}

func (b *CreateBatchesWorker) getUserBatchFromPG(userIds *[]string, job *model.Job) *[]User {
	var users []User
	start := time.Now()
	query := fmt.Sprintf("SELECT user_id, token, locale, tz FROM %s WHERE user_id IN (?)", GetPushDBTableName(job.App.Name, job.Service))
	_, err := b.Workers.PushDB.Query(&users, query, pg.In(*userIds))
	b.Workers.Statsd.Timing("get_csv_batch_from_pg", time.Now().Sub(start), job.Labels(), 1)

	b.checkErr(job, err)
	return &users
}

func (b *CreateBatchesWorker) processBatch(ids *[]string, job *model.Job) {
	if len(*ids) == 0 {
		return
	}
	l := b.Logger

	usersFromBatch := b.getUserBatchFromPG(ids, job)
	numUsersFromBatch := len(*usersFromBatch)
	log.I(l, "got users from db", func(cm log.CM) {
		cm.Write(zap.Int("usersInBatch", numUsersFromBatch))
	})

	if numUsersFromBatch != 0 {
		b.updateTotalUsers(job, numUsersFromBatch)
		b.updateTotalBatches(1, job)
		b.updateTotalTokens(numUsersFromBatch, job)
		b.sendBatches(*usersFromBatch, job)
	}
}

func (b *CreateBatchesWorker) sendBatches(users []User, job *model.Job) {
	l := b.Logger
	log.I(l, "sending batch of users to process batches worker", func(cm log.CM) {
		cm.Write(zap.Int("numUsers", len(users)))
	})
	_, err := b.Workers.CreateProcessBatchJob(job.ID.String(), job.App.Name, &users)
	b.checkErr(job, err)
}

func (b *CreateBatchesWorker) updateTotalUsers(job *model.Job, totalUsers int) {
	job.TotalUsers = totalUsers
	_, err := b.Workers.MarathonDB.Model(job).Set("total_users = coalesce(total_users, 0) + ?", totalUsers).Where("id = ?", job.ID).Update()
	b.checkErr(job, err)
}

func (b *CreateBatchesWorker) processIDs(userIds []string, msg *BatchPart) {
	l := b.Logger
	// create a controll group if needed
	controlGroupSize := int(math.Ceil(float64(len(userIds)) * msg.Job.ControlGroup))
	if controlGroupSize > 0 {
		if controlGroupSize >= len(userIds) {
			panic("control group size cannot be higher than number of users")
		}
		log.I(l, "this job has a control group!", func(cm log.CM) {
			cm.Write(
				zap.Int("controlGroupSize", controlGroupSize),
				zap.String("jobID", msg.Job.ID.String()),
			)
		})
		// shuffle slice in place
		for i := range userIds {
			j := rand.Intn(i + 1)
			userIds[i], userIds[j] = userIds[j], userIds[i]
		}
		// grab control group
		controlGroup := userIds[len(userIds)-controlGroupSize:]
		go b.Workers.SendControlGroupToRedis(&msg.Job, controlGroup)

		// cut control group from the slice
		userIds = append(userIds[:len(userIds)-controlGroupSize], userIds[len(userIds):]...)
		log.I(l, "control group cut from the users", func(cm log.CM) {
			cm.Write(
				zap.Int("usersRemaining", len(userIds)),
			)
		})
	}

	// pull from db and send to kafta
	b.processBatch(&userIds, &msg.Job)
}

// get the list of ids and send to redis the splited ids
func (b *CreateBatchesWorker) getIDs(buffer *bytes.Buffer, msg *BatchPart) []string {
	bBystes := buffer.Bytes()
	lines := b.ReadFromCSV(&bBystes, msg)
	totalLines := len(lines)
	start := 0
	end := totalLines

	// is not the first part
	if msg.Part != 0 && totalLines > 0 {
		str := fmt.Sprintf("%s-INI-%d", msg.Job.ID, msg.Part)
		b.Workers.RedisClient.Set(str, lines[0], 90*24*time.Hour)
		start++
	}

	// is not the last part
	if msg.Part != msg.TotalParts-1 {
		str := fmt.Sprintf("%s-END-%d", msg.Job.ID, msg.Part)
		b.Workers.RedisClient.Set(str, lines[totalLines-1], 90*24*time.Hour)
		end--
	}

	return lines[start:end]
}

func (b *CreateBatchesWorker) getSplitedIds(totalParts int, job *model.Job) []string {
	var ids []string
	for i := 0; i < totalParts-1; i++ {
		begin := fmt.Sprintf("%s-INI-%d", job.ID, i+1)
		end := fmt.Sprintf("%s-END-%d", job.ID, i)

		beginStr, err := b.Workers.RedisClient.Get(begin).Result()
		if err == redis.Nil {
			continue
		}
		b.checkErr(job, err)
		endStr, err := b.Workers.RedisClient.Get(end).Result()
		if err == redis.Nil {
			continue
		}
		b.checkErr(job, err)

		ids = append(ids, endStr+beginStr)
		ids = append(ids, beginStr)
		ids = append(ids, endStr)

		b.Workers.RedisClient.Del(begin)
		b.Workers.RedisClient.Del(end)
	}
	return ids
}

func (b *CreateBatchesWorker) setAsComplete(part int, job *model.Job) int {
	hash := job.ID.String()
	count, err := b.Workers.RedisClient.LPush(hash, part).Result()
	b.checkErr(job, err)
	return int(count)
}

// Process processes the messages sent to batch worker queue
func (b *CreateBatchesWorker) Process(message *goworkers2.Msg) error {

	var msg BatchPart
	data := message.Args().ToJson()
	err := json.Unmarshal([]byte(data), &msg)
	checkErr(b.Logger, err)

	l := b.Logger.With(
		zap.String("worker", nameCreateBatches),
		zap.Int("part", msg.Part),
		zap.Int("totalParts", msg.TotalParts),
	)

	b.Workers.Statsd.Incr(CreateBatchesWorkerStart, msg.Job.Labels(), 1)

	err = b.Workers.MarathonDB.Model(&msg.Job).Column("job.status", "App").Where("job.id = ?", msg.Job.ID).Select()
	b.checkErr(&msg.Job, err)

	if msg.Job.Status == stoppedJobStatus {
		l.Info("stopped job")
		b.Workers.Statsd.Incr(CreateBatchesWorkerCompleted, msg.Job.Labels(), 1)
		return nil
	}
	l.Info("starting")

	// if is the first element
	if msg.Part == 0 {
		msg.Job.TagRunning(b.Workers.MarathonDB, nameCreateBatches, "starting")
	}

	start := time.Now()
	_, buffer, err := b.Workers.S3Client.DownloadChunk(int64(msg.Start), int64(msg.Size), msg.Job.CSVPath)
	labels := msg.Job.Labels()
	labels = append(labels, fmt.Sprintf("error:%t", err != nil))
	b.Workers.Statsd.Timing(GetCsvFromS3Timing, time.Now().Sub(start), labels, 1)
	b.checkErr(&msg.Job, err)

	ids := b.getIDs(buffer, &msg)

	// pull from db, send to control and send to kafka
	b.processIDs(ids, &msg)

	completedParts := b.setAsComplete(msg.Part, &msg.Job)

	if completedParts == msg.TotalParts {
		ids = b.getSplitedIds(msg.TotalParts, &msg.Job)

		b.processIDs(ids, &msg)

		if msg.Job.TotalUsers == 0 {
			_, err := b.Workers.MarathonDB.Model(&msg.Job).Set("status = 'stopped', updated_at = ?, completed_at = ?", time.Now().UnixNano(), time.Now().UnixNano()).Where("id = ?", msg.Job.ID).Update()
			b.checkErr(&msg.Job, err)
			//b.updateCompletedAt(time.Now().UnixNano(), &msg.Job)
			msg.Job.TagError(b.Workers.MarathonDB, nameCreateBatches, "the job has finished without finding any valid user ids")
			b.Workers.Statsd.Incr(CreateBatchesWorkerError, msg.Job.Labels(), 1)
		} else {
			msg.Job.TagSuccess(b.Workers.MarathonDB, nameCreateBatches, "finished")
			b.Workers.Statsd.Incr(CreateBatchesWorkerCompleted, msg.Job.Labels(), 1)
		}

		// TODO: schedule a job to run after send all messages. This job will check
		// for errors and delete waste if a error happen
	} else {
		str := fmt.Sprintf("complete part %d of %d", completedParts, msg.TotalParts)
		msg.Job.TagRunning(b.Workers.MarathonDB, nameCreateBatches, str)
	}
	ids = nil

	l.Info("finished")

	return nil
}

func (b *CreateBatchesWorker) checkErr(job *model.Job, err error) {
	if err != nil {
		job.TagError(b.Workers.MarathonDB, nameCreateBatches, err.Error())
		b.Workers.Statsd.Incr(CreateBatchesWorkerError, job.Labels(), 1)

		checkErr(b.Logger, err)
	}
}
