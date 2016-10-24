package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/log"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/workers"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

type csvNotificationPayload struct {
	Message message `json:"message"`
	Bucket  string  `json:"bucket"`
	Key     string  `json:"key"`
}

// SendCsvNotificationHandler is the handler responsible for creating new pushes
func SendCsvNotificationHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()
		notifierID := c.Param("notifierID")
		l := application.Logger.With(
			zap.String("source", "csvNotificationHandler"),
			zap.String("operation", "sendNotification"),
		)

		notifierIDUuid, err := uuid.FromString(notifierID)
		if err != nil {
			log.E(l, "Could not convert notifierID into UUID.", func(cm log.CM) {
				cm.Write(
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
			return FailWith(400, err.Error(), c)
		}

		log.D(l, "Get notifier from DB")
		notifier, err := models.GetNotifierByID(application.Db, notifierIDUuid)
		if err != nil {
			log.E(l, "Could not find notifier.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Got notifier from DB")

		log.D(l, "Get app from DB")
		app, err := models.GetAppByID(application.Db, notifier.AppID)
		if err != nil {
			log.E(l, "Could not find app.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Got app from DB")

		log.D(l, "Parse payload")
		var payload csvNotificationPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			log.E(l, "Failed to parse json payload.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Parsed payload", func(cm log.CM) {
			cm.Write(zap.Object("payload", payload))
		})

		modifiers := [][]interface{}{{"LIMIT", 500}}

		message := &messages.InputMessage{
			App:     app.Name,
			Service: notifier.Service,
		}

		if payload.Message.Template != "" {
			message.Template = payload.Message.Template
		}
		if payload.Message.Params != nil {
			message.Params = payload.Message.Params
		}
		if payload.Message.Message != nil {
			message.Message = payload.Message.Message
		}
		if payload.Message.Metadata != nil {
			message.Metadata = payload.Message.Metadata
		}

		workerConfig := &workers.BatchCsvWorker{
			ConfigPath: application.ConfigPath,
			Logger:     l,
			Notifier:   notifier,
			App:        app,
			Message:    message,
			Modifiers:  modifiers,
			Bucket:     payload.Bucket,
			Key:        payload.Key,
		}
		worker, err := workers.GetBatchCsvWorker(workerConfig)
		if err != nil {
			log.E(l, "Invalid worker config,", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}

		worker.Start()

		return SucceedWith(map[string]interface{}{
			"id": worker.ID.String(),
		}, c)
	}
}
