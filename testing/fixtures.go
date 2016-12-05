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

	"github.com/jinzhu/gorm"
	"github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/model"
)

func getOpt(options map[string]interface{}, key string, defaultValue interface{}) interface{} {
	val, ok := options[key]
	if !ok {
		val = defaultValue
	}

	return val
}

//GetNewApp with specified optional values
func GetNewApp(db *gorm.DB, options ...map[string]interface{}) *model.App {
	opts := map[string]interface{}{}
	if len(options) == 1 {
		opts = options[0]
	}

	app := &model.App{}
	app.Name = getOpt(opts, "name", uuid.NewV4().String()).(string)
	app.BundleID = getOpt(opts, "bundleId", fmt.Sprintf("com.app.%s", strings.Split(uuid.NewV4().String(), "-")[0])).(string)
	app.CreatedBy = getOpt(opts, "createdBy", "test@test.com").(string)

	err := db.Create(&app).Error
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return app
}

//CreateTestApps for n apps
func CreateTestApps(db *gorm.DB, n int, options ...map[string]interface{}) []*model.App {
	apps := make([]*model.App, n)
	for i := 0; i < n; i++ {
		app := GetNewApp(db, options...)
		apps[i] = app
	}

	return apps
}