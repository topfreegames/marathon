package api

//
// import (
// 	"fmt"
// 	"time"
//
// 	"github.com/kataras/iris"
// 	"github.com/minio/minio-go"
// 	"github.com/uber-go/zap"
// )
//
// func UploadHandler(application *Application) func(c *iris.Context) {
// 	return func(c *iris.Context) {
// 		start := time.Now()
// 		l := application.Logger.With(
// 			zap.String("source", "uploadHandler"),
// 			zap.String("operation", "getUploadUrl"),
// 		)
//
// 		s3Bucket := application.Config.GetString("s3.bucket")
// 		s3Folder := application.Config.GetString("s3.folder")
// 		s3DaysExpiry := application.Config.GetInt("s3.daysExpiry")
// 		s3AccessKeyID := application.Config.GetString("s3.accessKey")
// 		s3SecretAccessKey := application.Config.GetString("s3.secretAccessKey")
//
// 		enableSSL := true
// 		policy := minio.NewPostPolicy()
// 		policy.SetKey(fmt.Sprintf("%s/upload-%v.csv", s3Folder, start))
// 		policy.SetBucket(s3Bucket)
// 		policy.SetExpires(start.UTC().AddDate(0, 0, s3DaysExpiry))
// 		s3Client, err := minio.New("s3.amazonaws.com", s3AccessKeyID, s3SecretAccessKey, enableSSL)
// 		if err != nil {
// 			l.Error(
// 				"Failed to create S3 client.",
// 				zap.Error(err),
// 				zap.Duration("duration", time.Now().Sub(start)),
// 			)
// 			FailWith(400, err.Error(), c)
// 			return
// 		}
//
// 		url, params, err := s3Client.PresignedPostPolicy(policy)
// 		if err != nil {
// 			l.Error(
// 				"Failed to create presigned POST policy.",
// 				zap.Error(err),
// 				zap.Duration("duration", time.Now().Sub(start)),
// 			)
// 			FailWith(400, err.Error(), c)
// 			return
// 		}
// 		params["url"] = url.String()
//
// 		l.Info(
// 			"Upload handler retrieved url and params successfully.",
// 			zap.Duration("duration", time.Now().Sub(start)),
// 		)
// 		m := make(map[string]interface{})
// 		for k, v := range params {
// 			m[k] = v
// 		}
//
// 		SucceedWith(m, c)
// 	}
// }
