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

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

// GetUploadURL handles a file upload
func (a *Application) GetUploadURL(c echo.Context) error {
	start := time.Now()
	l := a.Logger.With(
		zap.String("source", "uploadHandler"),
		zap.String("operation", "getUploadUrl"),
	)
	folder := a.Config.GetString("s3.folder")
	bucket := a.Config.GetString("s3.bucket")
	u, err := a.S3Client.PutObjectRequest(fmt.Sprintf("%s/%s/job-%v-%s.csv", bucket, folder, start.Unix(), uuid.NewV4().String()))
	if err != nil {
		log.E(l, "Failed to create presigned PUT policy.", func(cm log.CM) {
			cm.Write(
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	log.I(l, "Upload handler retrieved url and params successfully.", func(cm log.CM) {
		cm.Write(zap.Duration("duration", time.Now().Sub(start)))
	})
	m := make(map[string]interface{})
	// For some reason we need to unescape it
	uURL, err := url.QueryUnescape(u)
	if err != nil {
		log.E(l, "Failed to unescape presigned URL.", func(cm log.CM) {
			cm.Write(
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	m["url"] = uURL
	log.D(l, "Retrieved upload URL succesfully.", func(cm log.CM) {
		cm.Write(zap.Object("URLInfo", m))
	})
	return c.JSON(http.StatusOK, m)
}
