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

package testing

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/topfreegames/marathon/model"
)

func getOpt(options map[string]interface{}, key string, defaultValue interface{}) interface{} {
	val, ok := options[key]
	if !ok {
		val = defaultValue
	}

	return val
}

//CreateTestApp with specified optional values
func CreateTestApp(db interfaces.DB, options ...map[string]interface{}) *model.App {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	app := &model.App{}
	app.ID = getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)
	app.Name = getOpt(opts, "name", "testapp").(string)
	app.BundleID = getOpt(opts, "bundleId", fmt.Sprintf("com.app.%s", strings.Split(uuid.NewV4().String(), "-")[0])).(string)
	app.CreatedBy = getOpt(opts, "createdBy", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)

	err := db.Insert(&app)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return app
}

//CreateTestApps for n apps
func CreateTestApps(db interfaces.DB, n int, options ...map[string]interface{}) []*model.App {
	apps := make([]*model.App, n)
	for i := 0; i < n; i++ {
		app := CreateTestApp(db, options...)
		apps[i] = app
	}

	return apps
}

//CreateTestUser with specified optional values
func CreateTestUser(db interfaces.DB, options ...map[string]interface{}) *model.User {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	user := &model.User{}
	user.ID = getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)
	user.Email = getOpt(opts, "email", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)
	user.IsAdmin = getOpt(opts, "isAdmin", true).(bool)
	user.AllowedApps = getOpt(opts, "allowedApps", []uuid.UUID{uuid.NewV4()}).([]uuid.UUID)
	user.CreatedBy = getOpt(opts, "createdBy", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)

	err := db.Insert(&user)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return user
}

//CreateTestUsers for n users
func CreateTestUsers(db interfaces.DB, n int, options ...map[string]interface{}) []*model.User {
	users := make([]*model.User, n)
	for i := 0; i < n; i++ {
		user := CreateTestUser(db, options...)
		users[i] = user
	}

	return users
}

//GetAppPayload with specified optional values
func GetAppPayload(options ...map[string]interface{}) map[string]interface{} {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}
	id := getOpt(opts, "id", uuid.NewV4())
	name := getOpt(opts, "name", uuid.NewV4().String()).(string)
	bundleID := getOpt(opts, "bundleId", fmt.Sprintf("com.app.%s", strings.Split(uuid.NewV4().String(), "-")[0])).(string)

	app := map[string]interface{}{
		"id":       id,
		"name":     name,
		"bundleId": bundleID,
	}
	return app
}

//GetUserPayload with specified optional values
func GetUserPayload(options ...map[string]interface{}) map[string]interface{} {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}
	email := getOpt(opts, "email", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)
	isAdmin := getOpt(opts, "isAdmin", true).(bool)
	allowedApps := getOpt(opts, "allowedApps", []uuid.UUID{uuid.NewV4()}).([]uuid.UUID)

	user := map[string]interface{}{
		"email":       email,
		"isAdmin":     isAdmin,
		"allowedApps": allowedApps,
	}
	return user
}

//CreateTestTemplate with specified optional values
func CreateTestTemplate(db interfaces.DB, appID uuid.UUID, options ...map[string]interface{}) *model.Template {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	defaults := getOpt(opts, "defaults", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})
	body := getOpt(opts, "body", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})

	template := &model.Template{}
	template.AppID = appID
	template.Defaults = defaults
	template.Body = body
	template.ID = getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)
	template.Name = getOpt(opts, "name", uuid.NewV4().String()).(string)
	template.Locale = getOpt(opts, "locale", strings.Split(uuid.NewV4().String(), "-")[0]).(string)
	template.CreatedBy = getOpt(opts, "createdBy", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)

	err := db.Insert(&template)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return template
}

//CreateTestTemplates for n apps
func CreateTestTemplates(db interfaces.DB, appID uuid.UUID, n int, options ...map[string]interface{}) []*model.Template {
	templates := make([]*model.Template, n)
	for i := 0; i < n; i++ {
		template := CreateTestTemplate(db, appID, options...)
		templates[i] = template
	}

	return templates
}

//GetTemplatePayload with specified optional values
func GetTemplatePayload(options ...map[string]interface{}) map[string]interface{} {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	name := getOpt(opts, "name", uuid.NewV4().String()).(string)
	id := getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)
	locale := getOpt(opts, "locale", strings.Split(uuid.NewV4().String(), "-")[0]).(string)

	defaults := getOpt(opts, "defaults", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})
	body := getOpt(opts, "body", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})

	template := map[string]interface{}{
		"name":     name,
		"locale":   locale,
		"defaults": defaults,
		"body":     body,
		"id":       id,
	}
	return template
}

//GetTemplatePayloads with specified optional values
func GetTemplatePayloads(amount int, options ...map[string]interface{}) []map[string]interface{} {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
		opts["id"] = getOpt(opts, "id", uuid.NewV4())
	}

	templates := make([]map[string]interface{}, amount)
	for i := 0; i < amount; i++ {
		templates[i] = GetTemplatePayload(opts)
	}
	return templates
}

//CreateTestJob with specified optional values
func CreateTestJob(db interfaces.DB, appID uuid.UUID, templateName string, options ...map[string]interface{}) *model.Job {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	filters := getOpt(opts, "filters", map[string]interface{}{"locale": strings.Split(uuid.NewV4().String(), "-")[0]}).(map[string]interface{})
	context := getOpt(opts, "context", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})
	metadata := getOpt(opts, "metadata", map[string]interface{}{"meta": uuid.NewV4().String()}).(map[string]interface{})

	job := &model.Job{}
	job.AppID = appID
	job.TemplateName = templateName
	job.Filters = filters
	job.Metadata = metadata
	job.Context = context
	job.ControlGroup = getOpt(opts, "controlGroup", 0.0).(float64)
	job.Localized = getOpt(opts, "localized", false).(bool)
	job.ID = getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)
	job.Service = getOpt(opts, "service", "apns").(string)
	job.CSVPath = getOpt(opts, "csvPath", "").(string)
	job.PastTimeStrategy = getOpt(opts, "pastTimeStrategy", "").(string)
	job.ExpiresAt = getOpt(opts, "expiresAt", time.Now().Add(time.Hour).UnixNano()).(int64)
	job.CreatedBy = getOpt(opts, "createdBy", fmt.Sprintf("%s@test.com", strings.Split(uuid.NewV4().String(), "-")[0])).(string)
	job.StartsAt = getOpt(opts, "startsAt", time.Now().Add(time.Hour).UnixNano()).(int64)

	err := db.Insert(&job)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return job
}

//CreateTestJobs for n apps
func CreateTestJobs(db interfaces.DB, appID uuid.UUID, templateName string, n int, options ...map[string]interface{}) []*model.Job {
	jobs := make([]*model.Job, n)
	for i := 0; i < n; i++ {
		job := CreateTestJob(db, appID, templateName, options...)
		jobs[i] = job
	}

	return jobs
}

//GetJobPayload with specified optional values
func GetJobPayload(options ...map[string]interface{}) map[string]interface{} {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	filters := getOpt(opts, "filters", map[string]interface{}{"locale": strings.Split(uuid.NewV4().String(), "-")[0]}).(map[string]interface{})
	context := getOpt(opts, "context", map[string]interface{}{"value": uuid.NewV4().String()}).(map[string]interface{})
	metadata := getOpt(opts, "filters", map[string]interface{}{"meta": uuid.NewV4().String()}).(map[string]interface{})

	controlGroup := getOpt(opts, "controlGroup", 0.0).(float64)
	controlGroupCsvPath := getOpt(opts, "controlGroupCsvPath", "").(string)
	service := getOpt(opts, "service", "apns").(string)
	csvURL := getOpt(opts, "csvPath", "").(string)
	expiresAt := getOpt(opts, "expiresAt", time.Now().Add(time.Hour).UnixNano()).(int64)
	startsAt := getOpt(opts, "startsAt", time.Now().Add(time.Hour).UnixNano()).(int64)
	id := getOpt(opts, "id", uuid.NewV4()).(uuid.UUID)

	job := map[string]interface{}{
		"filters":             filters,
		"context":             context,
		"metadata":            metadata,
		"service":             service,
		"csvPath":             csvURL,
		"controlGroupCsvPath": controlGroupCsvPath,
		"controlGroup":        controlGroup,
		"expiresAt":           expiresAt,
		"startsAt":            startsAt,
		"id":                  id,
	}
	return job
}
