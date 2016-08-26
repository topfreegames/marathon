package api

import (
	"encoding/json"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/workers"

	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
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
	Filters filter  `json:"filters"`
	Message message `json:"message"`
}

// SendNotifierNotificationHandler is the handler responsible for creating new apps
func SendNotifierNotificationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		notifierID := c.Param("notifierID")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "sendNotification"),
			zap.String("notifierID", notifierID),
		)

		notifierIDUuid, err := uuid.FromString(notifierID)
		if err != nil {
			l.Error(
				"Could not convert notifierID into UUID.",
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
			FailWith(400, err.Error(), c)
			return
		}

		l.Debug("Get notifier from DB")
		notifier, err := models.GetNotifierByID(application.Db, notifierIDUuid)
		if err != nil {
			l.Error("Could not find notifier.", zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			FailWith(400, err.Error(), c)
			return
		}
		l.Debug("Got notifier from DB")

		l.Debug("Get app from DB")
		app, err := models.GetAppByID(application.Db, notifier.AppID)
		if err != nil {
			l.Error("Could not find app.", zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			FailWith(400, err.Error(), c)
			return
		}
		l.Debug("Got app from DB")

		l.Debug("Parse payload")
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
		l.Debug("Parsed payload", zap.Object("payload", payload))

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

		// TODO: Set in config
		modifiers := [][]interface{}{{"LIMIT", 1000}}

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

		workerConfig := &workers.BatchPGWorker{
			ConfigPath: application.ConfigPath,
			Logger:     l,
			Notifier:   notifier,
			Message:    message,
			Filters:    filters,
			Modifiers:  modifiers,
		}
		l.Debug("Get BatchPGWorker...")
		worker, err := workers.GetBatchPGWorker(workerConfig)
		if err != nil {
			l.Error("Invalid worker config,", zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			FailWith(400, err.Error(), c)
			return
		}
		l.Debug("Got BatchPGWorker...")

		l.Debug("Start BatchPGWorker...")
		worker.Start()
		l.Debug("Started BatchPGWorker...")

		SucceedWith(map[string]interface{}{
			"id": worker.ID.String(),
		}, c)
	}
}

// GetNotifierNotifications is the handler responsible retrieve a notification status
func GetNotifierNotifications(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		notifierID := c.Param("notifierID")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "getNotifications"),
			zap.String("notifierID", notifierID),
		)

		cli := application.RedisClient.Client
		redisKey := strings.Join([]string{notifierID, "*"}, "|")
		statuses := []map[string]interface{}{}

		l.Info("Get from redis", zap.String("redisKey", redisKey))
		keys, err := cli.Keys(redisKey).Result()
		if err != nil {
			if err.Error() != "redis: nil" {
				l.Error(
					"Failed to get notification status from redis",
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
				FailWith(400, err.Error(), c)
				return
			}
			l.Debug(
				"No notifications status from redis",
				zap.Duration("duration", time.Now().Sub(start)),
			)
			SucceedWith(map[string]interface{}{"statuses": statuses}, c)
			return
		}
		l.Info(
			"Got from redis",
			zap.Object("keys", keys),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		for i := range keys {
			key := string(keys[i])
			status, err := cli.Get(key).Result()
			if err != nil {
				if err.Error() != "redis: nil" {
					l.Error(
						"Failed to get notification status from redis",
						zap.Error(err),
						zap.String("key", key),
						zap.Duration("duration", time.Now().Sub(start)),
					)
					FailWith(400, err.Error(), c)
					return
				}
				l.Debug(
					"No notifications status from redis",
					zap.String("key", key),
					zap.Duration("duration", time.Now().Sub(start)),
				)
				SucceedWith(map[string]interface{}{"statuses": statuses}, c)
				return
			}
			var statusObj map[string]interface{}
			err = json.Unmarshal([]byte(status), &statusObj)
			statuses = append(statuses, statusObj)
		}

		SucceedWith(map[string]interface{}{"statuses": statuses}, c)
	}
}

// GetNotifierNotificationStatusHandler is the handler responsible retrieve a notification status
func GetNotifierNotificationStatusHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		notifierID := c.Param("notifierID")
		notificationID := c.Param("notificationID")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "getNotificationStatus"),
		)

		cli := application.RedisClient.Client
		redisKey := strings.Join([]string{notifierID, notificationID}, "|")
		l.Info("Get from redis", zap.String("redisKey", redisKey))
		status, err := cli.Get(redisKey).Result()
		if err != nil {
			l.Error(
				"Failed to get notification status from redis",
				zap.Error(err),
				zap.Duration("duration", time.Now().Sub(start)),
			)
			FailWith(400, err.Error(), c)
			return
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
