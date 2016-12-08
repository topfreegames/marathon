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

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
	"github.com/valyala/fasttemplate"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// InvalidMessageArray is the string returned when the message array of the process batch worker is not valid
var InvalidMessageArray = "array must be of the form [jobId, appName, users]"

// BuildTopicName builds a topic name based in appName, service and a template
func BuildTopicName(appName, service, topicTemplate string) string {
	return fmt.Sprintf(topicTemplate, appName, service)
}

// BatchWorkerMessage is the batch worker message struct
type BatchWorkerMessage struct {
	JobID   uuid.UUID
	AppName string
	Users   []User
}

// ParseProcessBatchWorkerMessageArray parses the message array of the process batch worker
func ParseProcessBatchWorkerMessageArray(arr []interface{}) (*BatchWorkerMessage, error) {
	// arr is of the following format
	// [jobId, appName, users]
	// users is an array of jsons { user_id: uuid, token: string, locale: string }

	if len(arr) != 3 {
		return nil, fmt.Errorf(InvalidMessageArray)
	}

	jobIDStr := arr[0].(string)
	jobID, err := uuid.FromString(jobIDStr)
	if err != nil {
		return nil, err
	}

	usersObj := arr[2].([]interface{})
	tmp, err := json.Marshal(usersObj)
	if err != nil {
		return nil, err
	}
	users := []User{}
	err = json.Unmarshal([]byte(string(tmp)), &users)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("there must be at least one user")
	}

	message := &BatchWorkerMessage{
		JobID:   jobID,
		AppName: arr[1].(string),
		Users:   users,
	}

	return message, nil
}

// BuildMessageFromTemplate build a message using a template and the context
func BuildMessageFromTemplate(template *model.Template, context map[string]interface{}) string {
	body, err := json.Marshal(template.Body)
	checkErr(err)
	bodyString := string(body)
	t := fasttemplate.New(bodyString, "{{", "}}")

	substitutions := make(map[string]interface{})
	for k, v := range template.Defaults {
		substitutions[k] = v
	}
	for k, v := range context {
		substitutions[k] = v
	}
	message := t.ExecuteString(substitutions)
	return message
}
