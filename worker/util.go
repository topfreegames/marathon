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
var InvalidMessageArray = "array must be of the form [jobId, appName, service, template, context, metadata, users, expireAt]"

// BuildTopicName builds a topic name based in appName, service and a template
func BuildTopicName(appName, service, topicTemplate string) string {
	return fmt.Sprintf(topicTemplate, appName, service)
}

// BatchWorkerMessage is the batch worker message struct
type BatchWorkerMessage struct {
	JobID     uuid.UUID
	AppName   string
	Service   string
	Template  *model.Template
	Context   map[string]interface{}
	Metadata  map[string]interface{}
	Users     []User
	ExpiresAt int64
}

// ParseProcessBatchWorkerMessageArray parses the message array of the process batch worker
func ParseProcessBatchWorkerMessageArray(arr []interface{}) (*BatchWorkerMessage, error) {
	// arr is of the following format
	// [jobId, appName, service, template, context, metadata, users]
	// template is a json: { body: json, defaults: json }
	// users is an array of jsons { id: uuid, token: string }

	if len(arr) != 8 {
		return nil, fmt.Errorf(InvalidMessageArray)
	}

	jobIDStr := arr[0].(string)
	jobID, err := uuid.FromString(jobIDStr)
	if err != nil {
		return nil, err
	}

	templateObj := arr[3].(map[string]interface{})
	tmp, err := json.Marshal(templateObj)
	if err != nil {
		return nil, err
	}
	var template *model.Template
	err = json.Unmarshal([]byte(string(tmp)), &template)
	if err != nil {
		return nil, err
	}

	usersObj := arr[6].([]map[string]interface{})
	tmp, err = json.Marshal(usersObj)
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
		JobID:     jobID,
		AppName:   arr[1].(string),
		Service:   arr[2].(string),
		Template:  template,
		Context:   arr[4].(map[string]interface{}),
		Metadata:  arr[5].(map[string]interface{}),
		Users:     users,
		ExpiresAt: arr[7].(int64),
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
