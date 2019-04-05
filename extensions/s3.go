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
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/uber-go/zap"
)

// AmazonS3 ...
type AmazonS3 struct {
	client  *s3.S3
	session *session.Session
	logger  zap.Logger
	conf    *viper.Viper
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
		client:  s3.New(sess),
		logger:  logger,
		conf:    conf,
		session: sess,
	}, nil
}

func getInfo(path string) (string, string, error) {
	splittedString := strings.SplitN(path, "/", 2)
	if len(splittedString) < 2 {
		return "", "", fmt.Errorf("Invalid path")
	}
	bucket := splittedString[0]
	objKey := splittedString[1]
	return bucket, objKey, nil
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// GetObject gets an object from s3
func (am *AmazonS3) GetObject(path string) ([]byte, error) {
	downloader := s3manager.NewDownloader(am.session, func(d *s3manager.Downloader) {
		d.Concurrency = 15
	})
	bucket, objKey, err := getInfo(path)
	if err != nil {
		return []byte{}, err
	}

	buffer := &aws.WriteAtBuffer{}
	_, err = downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objKey,
	})
	return buffer.Bytes(), err
}

// PutObject puts an object into s3.
func (am *AmazonS3) PutObject(path string, body *[]byte) (*s3.PutObjectOutput, error) {
	client := s3iface.S3API(am.client)

	bucket, objKey, err := getInfo(path)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(*body)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &objKey,
		Body:   b,
	}
	return client.PutObject(params)
}

// PutObjectRequest return a presigned url for uploading a file to s3
func (am *AmazonS3) PutObjectRequest(path string) (string, error) {
	client := s3iface.S3API(am.client)

	bucket, objKey, err := getInfo(path)
	if err != nil {
		return "", err
	}

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

// InitMultipartUpload build obeject to send multipart
func (am *AmazonS3) InitMultipartUpload(path string) (*s3.CreateMultipartUploadOutput, error) {
	bucket, objKey, err := getInfo(path)
	if err != nil {
		return nil, err
	}

	multiUpInput := &s3.CreateMultipartUploadInput{
		Bucket: &bucket,
		Key:    &objKey,
	}
	return am.client.CreateMultipartUpload(multiUpInput)
}

// UploadPart ...
func (am *AmazonS3) UploadPart(input *bytes.Buffer, multipartUpload *s3.CreateMultipartUploadOutput,
	partNumber int64) (*s3.UploadPartOutput, error) {
	var partNumberTemp int64
	partNumberTemp = int64(partNumber)
	upPartInput := &s3.UploadPartInput{
		Body:       bytes.NewReader(input.Bytes()),
		Bucket:     multipartUpload.Bucket,
		Key:        multipartUpload.Key,
		PartNumber: &partNumberTemp,
		UploadId:   multipartUpload.UploadId,
	}
	return am.client.UploadPart(upPartInput)
}

// CompleteMultipartUpload ...
func (am *AmazonS3) CompleteMultipartUpload(multipartUpload *s3.CreateMultipartUploadOutput, parts []*s3.CompletedPart) error {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   multipartUpload.Bucket,
		Key:      multipartUpload.Key,
		UploadId: multipartUpload.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	}
	_, err := am.client.CompleteMultipartUpload(completeInput)
	return err
}

func byteRange(start, size int64) string {
	return fmt.Sprintf("bytes=%d-%d", start, start+size-1)
}

// DownloadChunk downloads the chunk from s3
func (am *AmazonS3) DownloadChunk(start, size int64, path string) (int, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	client := s3iface.S3API(am.client)

	bucket, objKey, err := getInfo(path)
	if err != nil {
		return 0, nil, err
	}

	in := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objKey,
	}

	// Get the next byte range of data
	in.Range = aws.String(byteRange(start, size))

	resp, err := client.GetObject(in)
	if err != nil {
		return 0, nil, err
	}

	parts := strings.Split(*resp.ContentRange, "/")
	totalSize, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, nil, err
	}

	buf.ReadFrom(resp.Body)

	return totalSize, buf, nil
}
