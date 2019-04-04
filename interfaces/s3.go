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

package interfaces

import (
	"bytes"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3 represents the contract for a amazom S3
type S3 interface {
	InitMultipartUpload(path string) (*s3.CreateMultipartUploadOutput, error)
	PutObject(path string, body *[]byte) (*s3.PutObjectOutput, error)
	GetObject(path string) ([]byte, error)
	UploadPart(input *bytes.Buffer, multipartUpload *s3.CreateMultipartUploadOutput,
		partNumber int64) (*s3.UploadPartOutput, error)
	PutObjectRequest(path string) (string, error)
	CompleteMultipartUpload(multipartUpload *s3.CreateMultipartUploadOutput, parts []*s3.CompletedPart) error
	DownloadChunk(start, size int64, path string) (*s3.GetObjectOutput, error)
}
