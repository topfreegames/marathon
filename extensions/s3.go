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
	"bytes"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/uber-go/zap"
)

// AmazonS3 ...
type AmazonS3 struct {
	client *s3.S3
	logger zap.Logger
	conf   *viper.Viper
}

//NewS3 client with the specified configuration
func NewS3(conf *viper.Viper, logger zap.Logger) (interfaces.S3, error) {
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
	return &AmazonS3{
		client: s3.New(sess),
		logger: logger,
		conf:   conf,
	}, nil
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// GetObject gets an object from s3
//
func (am *AmazonS3) GetObject(path string) ([]byte, error) {
	client := s3iface.S3API(am.client)
	bucket := am.conf.GetString("s3.bucket")

	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &path,
	}
	resp, err := client.GetObject(params)
	if err != nil {
		return nil, err
	}
	return streamToByte(resp.Body), nil
}

// PutObject puts an object into s3.
func (am *AmazonS3) PutObject(path string, body *[]byte) (*s3.PutObjectOutput, error) {
	client := s3iface.S3API(am.client)
	bucket := am.conf.GetString("s3.bucket")
	//b := aws.ReadSeekCloser(*body)
	b := bytes.NewReader(*body)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   b,
	}
	return client.PutObject(params)
}

// PutObjectRequest return a presigned url for uploading a file to s3
func (am *AmazonS3) PutObjectRequest(path string) (string, error) {
	client := s3iface.S3API(am.client)
	bucket := am.conf.GetString("s3.bucket")
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
	}
	req, _ := client.PutObjectRequest(params)
	url, err := req.Presign(300 * time.Second)
	if err != nil {
		return "", err
	}
	return url, nil
}

// InitMultipartUpload build obeject to send multipart
func (am *AmazonS3) InitMultipartUpload(path string) (*s3.CreateMultipartUploadOutput, error) {
	bucket := am.conf.GetString("s3.bucket")

	multiUpInput := &s3.CreateMultipartUploadInput{
		Bucket: &bucket,
		Key:    &path,
	}
	return am.client.CreateMultipartUpload(multiUpInput)
}

// UploadPart ...
func (am *AmazonS3) UploadPart(stream io.Reader, multipartUpload *s3.CreateMultipartUploadOutput,
	partNumber int64) (*s3.UploadPartOutput, error) {
	bucket := am.conf.GetString("s3.bucket")

	bytess := streamToByte(stream)
	var partNumberTemp, lenTemp int64
	partNumber = int64(partNumber)
	lenTemp = int64(len(bytess))
	upPartInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(bytess),
		Bucket:        &bucket,
		Key:           multipartUpload.Key,
		PartNumber:    &partNumberTemp,
		UploadId:      multipartUpload.UploadId,
		ContentLength: &lenTemp,
	}
	return am.client.UploadPart(upPartInput)
}
