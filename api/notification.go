package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/workers"

	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

type filter struct {
	UserID string `json:"user_id"`
	Locale string `json:"locale"`
	Region string `json:"region"`
	Tz     string `json:"tz"`
	BuildN string `json:"build_n"`
	Scope  string `json:"scope"`
}

type message struct {
	Template string                 `json:"template"`
	Params   map[string]interface{} `json:"params"`
	Message  map[string]interface{} `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

type notificationPayload struct {
	App      string  `json:"app"`
	Service  string  `json:"service"`
	PageSize int     `json:"pageSize"`
	Filters  filter  `json:"filters"`
	Message  message `json:"message"`
}

// SendNotificationHandler is the handler responsible for creating new apps
func SendNotificationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "sendNotification"),
		)

		var payload notificationPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error(
				"Failed to parse json payload.",
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
			FailWith(400, err.Error(), c)
			return
		}

		filters := [][]interface{}{}
		if payload.Filters.UserID != "" {
			filters = append(filters, []interface{}{"user_id", payload.Filters.UserID})
		}
		if payload.Filters.Locale != "" {
			filters = append(filters, []interface{}{"locale", payload.Filters.Locale})
		}
		if payload.Filters.Region != "" {
			filters = append(filters, []interface{}{"region", payload.Filters.Region})
		}
		if payload.Filters.Tz != "" {
			filters = append(filters, []interface{}{"tz", payload.Filters.Tz})
		}
		if payload.Filters.BuildN != "" {
			filters = append(filters, []interface{}{"build_n", payload.Filters.BuildN})
		}
		if payload.Filters.Scope != "" {
			filters = append(filters, []interface{}{"scope", payload.Filters.Scope})
		}

		// TODO: Should we accept as parameters ?
		modifiers := [][]interface{}{{"LIMIT", payload.PageSize}}

		workerConfig := &workers.BatchPGWorker{
			ConfigPath: application.ConfigPath,
			Logger:     l,
		}
		worker, err := workers.GetBatchPGWorker(workerConfig)
		if err != nil {
			l.Error(
				"Invalid worker config,",
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
			FailWith(400, err.Error(), c)
		}

		message := &messages.InputMessage{
			App:     payload.App,
			Service: payload.Service,
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

		worker.StartWorker(message, filters, modifiers)

		SucceedWith(map[string]interface{}{
			"id": worker.ID.String(),
		}, c)
	}
}

// GetNotificationStatusHandler is the handler responsible retrieve a notification status
func GetNotificationStatusHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		notificationID := c.Param("notificationId")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "getNotificationStatus"),
		)

		cli := application.RedisClient.Client
		l.Info("Get from redis", zap.String("key", notificationID))
		status, err := cli.Get(notificationID).Result()
		if err != nil {
			l.Panic(
				"Failed to get notification status from redis",
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		}
		l.Info(
			"Got from redis",
			zap.String("key", notificationID),
			zap.String("value", status),
			zap.Duration("duration", time.Now().Sub(start)),
		)
		SucceedWith(map[string]interface{}{"status": status}, c)
	}
}
