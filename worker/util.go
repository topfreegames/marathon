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
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	pg "gopkg.in/pg.v5"
	"gopkg.in/redis.v5"

	raven "github.com/getsentry/raven-go"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
	"github.com/valyala/fasttemplate"
)

const stoppedJobStatus = "stopped"

// User is the struct that will keep users before sending them to send batches worker
type User struct {
	CreatedAt pg.NullTime `json:"created_at" sql:"created_at"`
	UserID    string      `json:"user_id" sql:"user_id"`
	Token     string      `json:"token" sql:"token"`
	Locale    string      `json:"locale" sql:"locale"`
	Region    string      `json:"region" sql:"region"`
	Tz        string      `json:"tz" sql:"tz"`
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

// SentBatches is a struct that helps tracking sent batches
type SentBatches struct {
	NumBatches int
	TotalUsers int
}

// IsUserIDValid tests whether a userID is valid or not
func IsUserIDValid(userID string) bool {
	forbiddenChars := []string{
		"\"",
		",",
		"'",
	}
	for _, c := range forbiddenChars {
		if strings.Contains(userID, c) {
			return false
		}
	}
	return true
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

// SplitUsersInBucketsByTZ splits users in buckets by tz
func SplitUsersInBucketsByTZ(users *[]User) map[string]*[]User {
	bucketsByTZ := map[string]*[]User{}
	for _, user := range *users {
		userTz := user.Tz
		if len(userTz) == 0 {
			userTz = "-0500"
		}
		if res, ok := bucketsByTZ[userTz]; ok {
			users := append(*res, user)
			bucketsByTZ[userTz] = &users
		} else {
			bucketsByTZ[userTz] = &[]User{user}
		}
	}
	return bucketsByTZ
}

func checkErr(l zap.Logger, err error) {
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
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
		operator := "="
		connector := " OR "
		if strings.Contains(key, "NOT") {
			key = strings.Trim(key, "NOT")
			operator = "!="
			connector = " AND "
		}
		strVal := val.(string)
		if strings.Contains(strVal, ",") {
			filterArray := []string{}
			vals := strings.Split(strVal, ",")
			for _, fVal := range vals {
				filterArray = append(filterArray, fmt.Sprintf("\"%s\"%s'%s'", key, operator, fVal))
			}
			queryFilters = append(queryFilters, fmt.Sprintf("(%s)", strings.Join(filterArray, connector)))
		} else {
			queryFilters = append(queryFilters, fmt.Sprintf("\"%s\"%s'%s'", key, operator, val))
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
func BuildMessageFromTemplate(template model.Template, context map[string]interface{}) (string, error) {
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

// RandomElementFromSlice gets a random element from a slice
func RandomElementFromSlice(elements []string) string {
	rand.Seed(time.Now().Unix())
	element := elements[rand.Intn(len(elements))]
	return element
}

//SendCircuitBreakJobEmail builds a circuit break job email message and sends it with sendgrid
func SendCircuitBreakJobEmail(sendgridClient *extensions.SendgridClient, job *model.Job, appName string, expireAt int64) error {
	subject := "Push job entered circuit break state"

	var platform string
	if job.Service == "apns" {
		platform = "iOS"
	} else if job.Service == "gcm" {
		platform = "Android"
	} else {
		platform = fmt.Sprintf("Unknown platform for service %s", job.Service)
	}

	expireAtDate := fmt.Sprintf("(%s)", time.Unix(0, expireAt).UTC().Format(time.RFC1123))

	message := fmt.Sprintf(`
Hello, your push job status has changed to circuit break.

App: %s
Template: %s
Platform: %s
JobID: %s
CreatedBy: %s

This job will be removed from the paused queue on %s. After this date the job will no longer be available.
Please fix the issues causing the circuit break and resume or stop it before then.
`, appName, job.TemplateName, platform, job.ID, job.CreatedBy, expireAtDate)
	return sendgridClient.SendgridSendEmail(job.CreatedBy, subject, message)
}
