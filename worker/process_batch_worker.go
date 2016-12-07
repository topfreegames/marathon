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
	"encoding/json"
	"fmt"

	workers "github.com/jrallison/go-workers"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
)

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Config      *viper.Viper
	KafkaClient *extensions.KafkaClient
}

// GetProcessBatchWorker gets a new ProcessBatchWorker
func GetProcessBatchWorker(config *viper.Viper) *ProcessBatchWorker {
	batchWorker := &ProcessBatchWorker{
		Config: config,
	}
	return batchWorker
}

func (batchWorker *ProcessBatchWorker) sendToKafka(appName, service, topic string, msg, metadata map[string]interface{}, deviceToken string) error {
	switch service {
	case "apns":
		_, _, err := batchWorker.KafkaClient.SendAPNSPush(topic, deviceToken, msg, metadata, 0) // TODO: use job expireAt instead of 0
		if err != nil {
			return err
		}
	case "gcm":
		_, _, err := batchWorker.KafkaClient.SendGCMPush(topic, deviceToken, msg, metadata, 0) // TODO: use job expireAt instead of 0
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("service should be in ['apns', 'gcm']")
	}
	return nil
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	// l := workers.Logger
	arr, err := message.Args().Array()
	checkErr(err)

	_, appName, service, template, context, metadata, users, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(err)

	msgStr := BuildMessageFromTemplate(template, context)
	var msg map[string]interface{}
	err = json.Unmarshal([]byte(msgStr), &msg)
	checkErr(err)

	topicTemplate := batchWorker.Config.GetString("workers.topicTemplate")
	topic := BuildTopicName(appName, service, topicTemplate)
	for _, user := range users {
		err = batchWorker.sendToKafka(appName, service, topic, msg, metadata, user.Token)
		checkErr(err)
	}

	// TODO: this worker should have a db connection to do this
	// err = db.Model(&model.Job{}).Set("completed_batches = completed_batches + 1").Where("id = ?", jobID).Update()
	// checkErr(err)
	// TODO: set completedAt when finished all batches
}
