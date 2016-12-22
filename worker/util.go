/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to
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
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/pg.v5"
	"gopkg.in/redis.v5"

	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	"github.com/valyala/fasttemplate"
)

// User is the struct that will keep users before sending them to send batches worker
type User struct {
	UserID string `json:"user_id" sql:"user_id"`
	Token  string `json:"token" sql:"token"`
	Locale string `json:"locale" sql:"locale"`
	Region string `json:"region" sql:"region"`
	Tz     string `json:"tz" sql:"tz"`
}

// Batch is a struct that helps tracking processes pages
type Batch struct {
	UserIds *[]string
	PageID  int
}

// DBPage is a struct that helps create batches from filters jobs
type DBPage struct {
	Page   int
	Offset int
}

func isPageProcessed(page int, jobID uuid.UUID, redisClient *redis.Client, l zap.Logger) bool {
	res, err := redisClient.SIsMember(fmt.Sprintf("%s-processedpages", jobID.String()), page).Result()
	checkErr(l, err)
	return res
}

// GetTimeOffsetFromUTCInSeconds returns the offset in seconds from UTC for tz
func GetTimeOffsetFromUTCInSeconds(tz string, l zap.Logger) (int, error) {
	r := regexp.MustCompile(`([\+|\-])(\d{2}).*(\d{2})`)
	matches := r.FindStringSubmatch(tz)
	if len(matches) < 4 {
		return 0, nil
	}
	hours, err := strconv.Atoi(matches[2])
	checkErr(l, err)
	minutes, err := strconv.Atoi(matches[3])
	checkErr(l, err)
	offsetInSeconds := (hours*60 + minutes) * 60
	if matches[1] == "+" {
		offsetInSeconds *= -1
	}
	return offsetInSeconds, err
}

func checkIsReexecution(jobID uuid.UUID, redisClient *redis.Client, l zap.Logger) bool {
	res, err := redisClient.Exists(fmt.Sprintf("%s-processedpages", jobID.String())).Result()
	checkErr(l, err)
	return res
}

func markProcessedPage(page int, jobID uuid.UUID, redisClient *redis.Client) {
	redisClient.SAdd(fmt.Sprintf("%s-processedpages", jobID.String()), page)
}

func updateTotalBatches(totalBatches int, job *model.Job, db *pg.DB, l zap.Logger) {
	job.TotalBatches = totalBatches
	_, err := db.Model(job).Column("total_batches").Update()
	checkErr(l, err)
}

// SplitUsersInBucketsByTZ splits users in buckets by tz
func SplitUsersInBucketsByTZ(users *[]User) map[string][]User {
	bucketsByTZ := map[string][]User{}
	for _, user := range *users {
		userTz := user.Tz
		if len(userTz) == 0 {
			userTz = "-0500"
		}
		if res, ok := bucketsByTZ[userTz]; ok {
			bucketsByTZ[userTz] = append(res, user)
		} else {
			bucketsByTZ[userTz] = []User{user}
		}
	}
	return bucketsByTZ
}

func checkErr(l zap.Logger, err error) {
	if err != nil {
		log.P(l, "Worker panic.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
}

// GetWhereClauseFromFilters returns a string cointaining the where clause to use in the query
func GetWhereClauseFromFilters(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	queryFilters := []string{}
	for key, val := range filters {
		strVal := val.(string)
		if strings.Contains(strVal, ",") {
			filterArray := []string{}
			vals := strings.Split(strVal, ",")
			for _, fVal := range vals {
				filterArray = append(filterArray, fmt.Sprintf("\"%s\"='%s'", key, fVal))
			}
			queryFilters = append(queryFilters, fmt.Sprintf("(%s)", strings.Join(filterArray, " OR ")))
		} else {
			queryFilters = append(queryFilters, fmt.Sprintf("\"%s\"='%s'", key, val))
		}
	}
	return strings.Join(queryFilters, " AND ")
}

// GetPushDBTableName get the table name using appName and service
func GetPushDBTableName(appName, service string) string {
	return fmt.Sprintf("%s_%s", appName, service)
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
func BuildMessageFromTemplate(template *model.Template, context map[string]interface{}) (string, error) {
	body, err := json.Marshal(template.Body)
	if err != nil {
		return "", err
	}
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
	return message, nil
}
