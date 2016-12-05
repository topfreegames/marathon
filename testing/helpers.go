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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/api"
	"github.com/uber-go/zap"
)

//GetTestDB for usage in tests
func GetTestDB() *gorm.DB {
	dbURL := viper.GetString("database.url")
	db, err := gorm.Open("postgres", dbURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return db
}

// GetFaultyTestDB returns an ill-configured database for usage in tests
func GetFaultyTestDB() *gorm.DB {
	faultyDbURL := viper.GetString("database.faulty")
	faultyDb, _ := gorm.Open("postgres", faultyDbURL)
	return faultyDb
}

//GetConfPath for the environment we're in
func GetConfPath() string {
	conf := "../config/test.yaml"

	// TODO: return CI conf when necessary
	// conf = "../config/ci.yaml"
	return conf
}

//GetDefaultTestApp returns a new marathon API Application bound to 0.0.0.0:8833 for test
func GetDefaultTestApp(logger zap.Logger) *api.Application {
	return api.GetApplication("0.0.0.0", 8833, false, logger, GetConfPath())
}

//Get from server
func Get(app *api.Application, url, auth string) (int, string) {
	return doRequest(app, "GET", url, "", auth)
}

//Post to server
func Post(app *api.Application, url, body string, auth string) (int, string) {
	return doRequest(app, "POST", url, body, auth)
}

//Put to server
func Put(app *api.Application, url, body string, auth string) (int, string) {
	return doRequest(app, "PUT", url, body, auth)
}

//Delete from server
func Delete(app *api.Application, url, auth string) (int, string) {
	return doRequest(app, "DELETE", url, "", auth)
}

func doRequest(app *api.Application, method, url, body, auth string) (int, string) {
	ts := httptest.NewServer(app.API)
	defer ts.Close()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", ts.URL, url), reader)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	if auth != "" {
		req.Header.Add("x-forwarded-email", auth)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	if res != nil {
		defer res.Body.Close()
	}

	b, err := ioutil.ReadAll(res.Body)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return res.StatusCode, string(b)
}

//ResetStdout back to os.Stdout
var ResetStdout func()

//ReadStdout value
var ReadStdout func() string

//MockStdout to read it's value later
func MockStdout() {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	os.Stdout = w

	ReadStdout = func() string {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		r.Close()
		return buf.String()
	}

	ResetStdout = func() {
		w.Close()
		os.Stdout = stdout
	}
}

//TestBuffer is a mock buffer
type TestBuffer struct {
	bytes.Buffer
}

//Sync does nothing
func (b *TestBuffer) Sync() error {
	return nil
}

//Lines returns all lines of log
func (b *TestBuffer) Lines() []string {
	output := strings.Split(b.String(), "\n")
	return output[:len(output)-1]
}

//Stripped removes new lines
func (b *TestBuffer) Stripped() string {
	return strings.TrimRight(b.String(), "\n")
}
