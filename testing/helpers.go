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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	pg "gopkg.in/pg.v5"
	"gopkg.in/pg.v5/orm"
	"gopkg.in/pg.v5/types"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/api"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

// FakeKafkaProducer is a mock producer that implements PushProducer interface
type FakeKafkaProducer struct {
	APNSMessages []string
	GCMMessages  []string
}

// NewFakeKafkaProducer creates a new FakeKafkaProducer
func NewFakeKafkaProducer() *FakeKafkaProducer {
	return &FakeKafkaProducer{
		APNSMessages: []string{},
		GCMMessages:  []string{},
	}
}

// SendAPNSPush for testing
func (f *FakeKafkaProducer) SendAPNSPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64, templateName string) error {
	msg := messages.NewAPNSMessage(
		deviceToken,
		pushExpiry,
		payload,
		messageMetadata,
		pushMetadata,
		templateName,
	)

	if val, ok := pushMetadata["dryRun"]; ok {
		if dryRun, _ := val.(bool); dryRun {
			msg.DeviceToken = extensions.GenerateFakeID(64)
		}
	}

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}

	f.APNSMessages = append(f.APNSMessages, message)

	return nil
}

// SendGCMPush for testing
func (f *FakeKafkaProducer) SendGCMPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64, templateName string) error {
	msg := messages.NewGCMMessage(
		deviceToken,
		payload,
		messageMetadata,
		pushMetadata,
		pushExpiry,
		templateName,
	)

	if val, ok := pushMetadata["dryRun"]; ok {
		if dryRun, _ := val.(bool); dryRun {
			msg.To = extensions.GenerateFakeID(152)
			msg.DryRun = true
		}
	}

	message, err := msg.ToJSON()

	if err != nil {
		return err
	}

	f.GCMMessages = append(f.GCMMessages, message)

	return nil
}

//PGMock should be used for tests that need to connect to PG
type PGMock struct {
	Execs        [][]interface{}
	ExecOnes     [][]interface{}
	Queries      [][]interface{}
	Closed       bool
	RowsAffected int
	RowsReturned int
	Error        error
}

//NewPGMock creates a new instance
func NewPGMock(rowsAffected, rowsReturned int, errOrNil ...error) *PGMock {
	var err error
	if len(errOrNil) == 1 {
		err = errOrNil[0]
	}
	return &PGMock{
		Closed:       false,
		RowsAffected: rowsAffected,
		RowsReturned: rowsReturned,
		Error:        err,
	}
}

func (m *PGMock) getResult() *types.Result {
	return types.NewResult([]byte(fmt.Sprintf(" %d", m.RowsAffected)), m.RowsReturned)
}

//Close records that it is closed
func (m *PGMock) Close() error {
	m.Closed = true

	if m.Error != nil {
		return m.Error
	}

	return nil
}

//Exec stores executed params
func (m *PGMock) Exec(obj interface{}, params ...interface{}) (*types.Result, error) {
	op := []interface{}{
		obj, params,
	}
	m.Execs = append(m.Execs, op)

	if m.Error != nil {
		return nil, m.Error
	}

	result := m.getResult()
	return result, nil
}

//ExecOne stores executed params
func (m *PGMock) ExecOne(obj interface{}, params ...interface{}) (*types.Result, error) {
	op := []interface{}{
		obj, params,
	}
	m.ExecOnes = append(m.ExecOnes, op)

	if m.Error != nil {
		return nil, m.Error
	}

	result := m.getResult()
	return result, nil
}

//Query stores executed params
func (m *PGMock) Query(obj interface{}, query interface{}, params ...interface{}) (*types.Result, error) {
	op := []interface{}{
		obj, query, params,
	}
	m.Execs = append(m.Execs, op)

	if m.Error != nil {
		return nil, m.Error
	}

	result := m.getResult()
	return result, nil
}

//Model mock for testing
func (m *PGMock) Model(params ...interface{}) *orm.Query {
	return nil
}

//Select mock for testing
func (m *PGMock) Select(params interface{}) error {
	return nil
}

//Insert mock for testing
func (m *PGMock) Insert(params ...interface{}) error {
	return nil
}

//Update mock for testing
func (m *PGMock) Update(params interface{}) error {
	return nil
}

//Delete mock for testing
func (m *PGMock) Delete(params interface{}) error {
	return nil
}

// FakeS3 for usage in tests
type FakeS3 struct {
	s3iface.S3API
	FakeStorage *map[string][]byte
}

// MyReaderCloser for usage in tests
type MyReaderCloser struct {
	io.Reader
}

// Close for usage in tests
func (MyReaderCloser) Close() error {
	return nil
}

// RedisReplyToBytes for testing
func RedisReplyToBytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case string:
		return []byte(reply), nil
	case nil:
		return nil, errors.New("nil returned")
	}
	return nil, fmt.Errorf("unexpected type for Bytes, got type %T", reply)
}

// GetObject for usage in tests
func (s *FakeS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if val, ok := (*s.FakeStorage)[fmt.Sprintf("%s/%s", *input.Bucket, *input.Key)]; ok {
		reader := bytes.NewReader(val)
		return &s3.GetObjectOutput{
			Body: MyReaderCloser{reader},
		}, nil
	}
	return nil, fmt.Errorf("NoSuchKey: The specified key does not exist. status code: 404, request id: 000000000000TEST")
}

// NewFakeS3 creates a new FakeS3
func NewFakeS3() *FakeS3 {
	return &FakeS3{
		FakeStorage: &map[string][]byte{},
	}
}

// PutObject for using in tests
func (s *FakeS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	b, err := ioutil.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	(*s.FakeStorage)[fmt.Sprintf("%s/%s", *input.Bucket, *input.Key)] = b
	return &s3.PutObjectOutput{}, nil
}

// ReadLinesFromIOReader for testing
func ReadLinesFromIOReader(reader io.Reader) []string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	scanner := bufio.NewScanner(buf)
	scanner.Split(bufio.ScanLines)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
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
