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

	"github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/model"
	"github.com/valyala/fasttemplate"
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

func (batchWorker *ProcessBatchWorker) parseMessageArray(arr []interface{}) (uuid.UUID, *model.Template, map[string]interface{}, []model.User) {
	// arr is of the following format
	// [jobId, template, context, users]
	// template is a json: { body: json, defaults: json }
	// users is an array of jsons { id: uuid, token: string }

	jobIDStr := arr[0].(string)
	jobID, err := uuid.FromString(jobIDStr)
	checkErr(err)

	templateStr := arr[1].(string)
	var template *model.Template
	err = json.Unmarshal([]byte(templateStr), &template)
	checkErr(err)

	contextStr := arr[2].(string)
	var context map[string]interface{}
	err = json.Unmarshal([]byte(contextStr), &context)
	checkErr(err)

	usersStr := arr[3].(string)
	users := []model.User{}
	err = json.Unmarshal([]byte(usersStr), &users)
	checkErr(err)

	return jobID, template, context, users
}

func (batchWorker *ProcessBatchWorker) buildMessage(template *model.Template, context map[string]interface{}) string {
	body, _ := json.Marshal(template.Body)
	bodyString := string(body)
	t := fasttemplate.New(bodyString, "{{", "}}")

	var substitutions map[string]interface{}

	for k, v := range template.Defaults {
		substitutions[k] = v
	}

	for k, v := range context {
		substitutions[k] = v
	}

	message := t.ExecuteString(substitutions)
	return message
}

func (batchWorker *ProcessBatchWorker) sendToKafka(uuid.UUID, string, string) error {
	return nil
}

// Process processes the messages sent to batch worker queue and send them to kafka
func (batchWorker *ProcessBatchWorker) Process(message *workers.Msg) {
	// l := workers.Logger
	arr, err := message.Args().Array()
	checkErr(err)

	jobID, template, context, users := batchWorker.parseMessageArray(arr)

	message := batchWorker.buildMessage(template, context)

	// TODO: send to kafka
	for _, user := range users {
		batchWorker.sendToKafka(jobID, message, user.Token)
	}
}
