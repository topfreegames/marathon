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

package extensions

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

//NewS3 client with the specified configuration
func NewS3(conf *viper.Viper, logger zap.Logger) (s3iface.S3API, error) {
	region := conf.GetString("s3.region")
	accessKey := conf.GetString("s3.accessKey")
	secretAccessKey := conf.GetString("s3.secretAccessKey")
	credentials := credentials.NewStaticCredentials(accessKey, secretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials,
	})
	if err != nil {
		return nil, err
	}
	logger.Debug("configured s3 extensions", zap.String("region", region))
	svc := s3.New(sess)
	s3 := s3iface.S3API(svc)
	return s3, nil
}

//S3GetObject gets an object from s3
func S3GetObject(client s3iface.S3API, path string) (*io.ReadCloser, error) {
	splittedString := strings.SplitN(path, "/", 2)
	if len(splittedString) < 2 {
		return nil, fmt.Errorf("Invalid path")
	}
	bucket := splittedString[0]
	objKey := splittedString[1]
	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objKey,
	}
	resp, err := client.GetObject(params)
	if err != nil {
		return nil, err
	}
	return &resp.Body, nil
}

// S3PutObjectRequest return a presigned url for uploading a file to s3
func S3PutObjectRequest(conf *viper.Viper, client s3iface.S3API, key string) (string, error) {
	bucket := conf.GetString("s3.bucket")
	folder := conf.GetString("s3.folder")
	objKey := fmt.Sprintf("%s/%s", folder, key)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &objKey,
	}
	req, _ := client.PutObjectRequest(params)
	url, err := req.Presign(300 * time.Second)
	if err != nil {
		return "", err
	}
	return url, nil
}
