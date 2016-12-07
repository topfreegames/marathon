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
	"github.com/uber-go/zap"
)

// ProcessBatchWorker is the ProcessBatchWorker struct
type ProcessBatchWorker struct {
	Config    *viper.Viper
	Logger    zap.Logger
	Kafka     *extensions.KafkaClient
	Zookeeper *extensions.ZookeeperClient
}

// NewProcessBatchWorker gets a new ProcessBatchWorker
func NewProcessBatchWorker(config *viper.Viper, logger zap.Logger) *ProcessBatchWorker {
	zookeeper, err := extensions.NewZookeeperClient(config, logger)
	checkErr(err)
	//Wait 10s at max for a connection
	zookeeper.WaitForConnection(10)
	kafka, err := extensions.NewKafkaClient(zookeeper, config, logger)
	checkErr(err)
	batchWorker := &ProcessBatchWorker{
		Config:    config,
		Logger:    logger,
		Kafka:     kafka,
		Zookeeper: zookeeper,
	}
	return batchWorker
}

func (batchWorker *ProcessBatchWorker) sendToKafka(service, topic string, msg, metadata map[string]interface{}, deviceToken string, expiresAt int64) error {
	pushExpiry := expiresAt / 1000000000 // convert from nanoseconds to seconds
	switch service {
	case "apns":
		_, _, err := batchWorker.Kafka.SendAPNSPush(topic, deviceToken, msg, metadata, pushExpiry)
		if err != nil {
			return err
		}
	case "gcm":
		_, _, err := batchWorker.Kafka.SendGCMPush(topic, deviceToken, msg, metadata, pushExpiry)
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

	parsed, err := ParseProcessBatchWorkerMessageArray(arr)
	checkErr(err)

	msgStr := BuildMessageFromTemplate(parsed.Template, parsed.Context)
	var msg map[string]interface{}
	err = json.Unmarshal([]byte(msgStr), &msg)
	checkErr(err)

	topicTemplate := batchWorker.Config.GetString("workers.topicTemplate")
	topic := BuildTopicName(parsed.AppName, parsed.Service, topicTemplate)
	for _, user := range parsed.Users {
		err = batchWorker.sendToKafka(parsed.Service, topic, msg, parsed.Metadata, user.Token, parsed.ExpiresAt)
		checkErr(err)
	}

	// TODO: this worker should have a db connection to do this
	// err = db.Model(&model.Job{}).Set("completed_batches = completed_batches + 1").Where("id = ?", jobID).Update()
	// checkErr(err)
	// TODO: set completedAt when finished all batches
}
