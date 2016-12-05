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
	"encoding/csv"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
)

// CreateBatchesWorker is the CreateBatchesWorker struct
type CreateBatchesWorker struct {
	MarathonDBURL string
	PushDBURL     string
	MarathonDB    *gorm.DB
	PushDB        *gorm.DB
}

// GetCreateBatchesWorker gets a new CreateBatchesWorker
func GetCreateBatchesWorker(marathonDBURL, pushDBURL string) *CreateBatchesWorker {
	b := &CreateBatchesWorker{
		MarathonDBURL: marathonDBURL,
		PushDBURL:     pushDBURL,
	}
	b.configure()
	return b
}

func (b *CreateBatchesWorker) configure() {
	b.configureDatabases()
}

func (b *CreateBatchesWorker) configureDatabases() {
	db, err := gorm.Open("postgres", b.MarathonDBURL)
	if err != nil {
		panic(err)
	}
	b.MarathonDB = db
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (b *CreateBatchesWorker) readRemoteCSV(csvURL string) []string {
	res := []string{}
	resp, err := http.Get(csvURL)
	checkErr(err)
	defer resp.Body.Close()
	r := csv.NewReader(resp.Body)
	lines, err := r.ReadAll()
	checkErr(err)
	for i, line := range lines {
		if i == 0 {
			// skip header line
			continue
		}
		res = append(res, line[0])
	}
	return res
}

func (b *CreateBatchesWorker) getTokensUsingCSV(csvURL string) []string {
	tokens := []string{}
	userIds := b.readRemoteCSV(csvURL)
	for _, id := range userIds {
		workers.Logger.Printf("id readen %s", id)
	}
	return tokens
}

// Process processes the messages sent to batch worker queue
func (b *CreateBatchesWorker) Process(message *workers.Msg) {
	l := workers.Logger
	arr, err := message.Args().Array()
	if err != nil {
		checkErr(err)
	}
	jobID := arr[0]
	id, err := uuid.FromString(jobID.(string))
	checkErr(err)
	job := &model.Job{
		ID: id,
	}
	err = b.MarathonDB.Preload("Template.App").Preload("App").Where(job).First(&job).Error
	checkErr(err)
	var pushIds []string
	if len(job.CsvURL) > 0 {
		pushIds = b.getTokensUsingCSV(job.CsvURL)
		// Load userIds from csv
		// Download the csv from s3
		// reads user_ids
		// get in batches from pn
	} else {
		// Find the ids based on filters
	}
	l.Printf("successfully got job from db %s", pushIds)
}
