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

	"gopkg.in/pg.v5"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/api"
	"github.com/uber-go/zap"
)

// FakeS3 for usage in tests
type FakeS3 struct {
	s3iface.S3API
}

// MyReaderCloser for usage in tests
type MyReaderCloser struct {
	io.Reader
}

// Close for usage in tests
func (MyReaderCloser) Close() error {
	return nil
}

// GetObject for usage in tests
func (s *FakeS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	validObjs := map[string]s3.GetObjectOutput{
		"test/jobs/obj1.csv": s3.GetObjectOutput{
			Body: MyReaderCloser{bytes.NewBufferString(
				`userids
9e558649-9c23-469d-a11c-59b05813e3d5
57be9009-e616-42c6-9cfe-505508ede2d0
a8e8d2d5-f178-4d90-9b31-683ad3aae920
5c3033c0-24ad-487a-a80d-68432464c8de
4223171e-c665-4612-9edd-485f229240bf
2df5bb01-15d1-4569-bc56-49fa0a33c4c3
67b872de-8ae4-4763-aef8-7c87a7f928a7
3f8732a1-8642-4f22-8d77-a9688dd6a5ae
21854bbf-ea7e-43e3-8f79-9ab2c121b941
843a61f8-45b3-44f9-9ab7-8becb2765653
`)},
		},
		"test/jobs/obj2.csv": s3.GetObjectOutput{
			Body: MyReaderCloser{bytes.NewBufferString(`userids`)},
		},
	}
	if val, ok := validObjs[*input.Key]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("NoSuchKey: The specified key does not exist. status code: 404, request id: 000000000000TEST")
}

//GetTestDB for usage in tests
func GetTestDB(a *api.Application) *pg.DB {
	host := a.Config.GetString("db.host")
	user := a.Config.GetString("db.user")
	pass := a.Config.GetString("db.pass")
	database := a.Config.GetString("db.database")
	port := a.Config.GetInt("db.port")
	poolSize := a.Config.GetInt("db.poolSize")
	maxRetries := a.Config.GetInt("db.maxRetries")
	db := pg.Connect(&pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	})
	return db
}

// GetFaultyTestDB returns an ill-configured database for usage in tests
func GetFaultyTestDB(a *api.Application) *pg.DB {
	host := a.Config.GetString("faultyDb.host")
	user := a.Config.GetString("faultyDb.user")
	pass := a.Config.GetString("faultyDb.pass")
	database := a.Config.GetString("faultyDb.database")
	port := a.Config.GetInt("faultyDb.port")
	poolSize := a.Config.GetInt("faultyDb.poolSize")
	maxRetries := a.Config.GetInt("faultyDb.maxRetries")
	faultyDb := pg.Connect(&pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   database,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
	})

	return faultyDb
}

//GetConfPath for the environment we're in
func GetConfPath() string {
	conf := "../config/test.yaml"

	// TODO: return CI conf when necessary
	// conf = "../config/ci.yaml"
	return conf
}

//GetConf returns a viper config for the environment we're in
func GetConf() *viper.Viper {
	confPath := GetConfPath()
	config := viper.New()
	config.SetConfigFile(confPath)
	config.ReadInConfig()
	return config
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
