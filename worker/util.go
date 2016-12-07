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

// ParseProcessBatchWorkerMessageArray parses the message array of the process batch worker
func ParseProcessBatchWorkerMessageArray(arr []interface{}) (uuid.UUID, *model.Template, map[string]interface{}, []model.User, error) {
	// arr is of the following format
	// [jobId, template, context, users]
	// template is a json: { body: json, defaults: json }
	// users is an array of jsons { id: uuid, token: string }

	if len(arr) != 4 {
		return uuid.UUID{}, nil, nil, nil, fmt.Errorf("array must be of the form [jobId, template, context, users]")
	}

	jobIDStr := arr[0].(string)
	jobID, err := uuid.FromString(jobIDStr)
	if err != nil {
		return uuid.UUID{}, nil, nil, nil, err
	}

	templateStr := arr[1].(string)
	var template *model.Template
	err = json.Unmarshal([]byte(templateStr), &template)
	if err != nil {
		return uuid.UUID{}, nil, nil, nil, err
	}

	if len(template.Defaults) == 0 {
		return uuid.UUID{}, nil, nil, nil, fmt.Errorf("template must have defaults")
	}

	if len(template.Body) == 0 {
		return uuid.UUID{}, nil, nil, nil, fmt.Errorf("template must have body")
	}

	contextStr := arr[2].(string)
	var context map[string]interface{}
	err = json.Unmarshal([]byte(contextStr), &context)
	if err != nil {
		return uuid.UUID{}, nil, nil, nil, err
	}

	usersStr := arr[3].(string)
	users := []model.User{}
	err = json.Unmarshal([]byte(usersStr), &users)
	if err != nil {
		return uuid.UUID{}, nil, nil, nil, err
	}

	if len(users) == 0 {
		return uuid.UUID{}, nil, nil, nil, fmt.Errorf("there must be at least one user")
	}

	return jobID, template, context, users, nil
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
