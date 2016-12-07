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
	workers "github.com/jrallison/go-workers"
	"github.com/spf13/viper"
)

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Config *viper.Viper
}

// GetProcessBatchWorker gets a new ProcessBatchWorker
func GetProcessBatchWorker(config *viper.Viper) *ProcessBatchWorker {
	batchWorker := &ProcessBatchWorker{
		Config: config,
	}
	return batchWorker
}

func (batchWorker *ProcessBatchWorker) sendToKafka(appName, service, msg string, metadata map[string]interface{}, deviceToken string) error {
	return nil
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	// l := workers.Logger
	arr, err := message.Args().Array()
	checkErr(err)

	_, appName, service, template, context, metadata, users, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(err)

	msg := BuildMessageFromTemplate(template, context)

	// TODO: send to kafka
	for _, user := range users {
		batchWorker.sendToKafka(appName, service, msg, metadata, user.Token)
	}

	// TODO: increment job completed batches
}
