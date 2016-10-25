package api

import (
	"fmt"
	"net/url"
	"time"

	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/labstack/echo"
	"github.com/minio/minio-go"
	"github.com/uber-go/zap"
)

// UploadHandler handles a file upload
func UploadHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()
		l := application.Logger.With(
			zap.String("source", "uploadHandler"),
			zap.String("operation", "getUploadUrl"),
		)

		s3Bucket := application.Config.GetString("s3.bucket")
		s3Folder := application.Config.GetString("s3.folder")
		s3DaysExpiry := time.Second * 10 * 60 * 60
		s3AccessKeyID := application.Config.GetString("s3.accessKey")
		s3SecretAccessKey := application.Config.GetString("s3.secretAccessKey")
		enableSSL := true

		s3Key := fmt.Sprintf("%s/upload%v.csv", s3Folder, start.Unix())
		s3Client, err := minio.New("s3.amazonaws.com", s3AccessKeyID, s3SecretAccessKey, enableSSL)
		if err != nil {
			log.E(l, "Failed to create S3 client.", func(cm log.CM) {
				cm.Write(
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
			return FailWith(400, err.Error(), c)
		}
		var u *url.URL
		err = WithSegment("s3client-getPresignedUrl", c, func() error {
			u, err = s3Client.PresignedPutObject(s3Bucket, s3Key, s3DaysExpiry)
			return err
		})
		if err != nil {
			log.E(l, "Failed to create presigned PUT policy.", func(cm log.CM) {
				cm.Write(
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
			return FailWith(400, err.Error(), c)
		}
		log.I(l, "Upload handler retrieved url and params successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
		m := make(map[string]interface{})
		// For some reason we need to unescape it
		uURL, err := url.QueryUnescape(u.String())
		if err != nil {
			log.E(l, "Failed to unescape presigned URL.", func(cm log.CM) {
				cm.Write(
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
			return FailWith(500, err.Error(), c)
		}
		m["url"] = uURL
		return SucceedWith(m, c)
	}
}
