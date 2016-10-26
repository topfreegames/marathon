package api

import (
	"encoding/json"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/workers"

	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/labstack/echo"
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
func SendNotifierNotificationHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()
		notifierID := c.Param("notifierID")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "sendNotification"),
			zap.String("notifierID", notifierID),
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
		var notifier *models.Notifier
		err = WithSegment("notifier-retrieve", c, func() error {
			notifier, err = models.GetNotifierByID(application.Db, notifierIDUuid)
			return err
		})
		if err != nil {
			log.E(l, "Could not find notifier.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Got notifier from DB")

		log.D(l, "Get app from DB")
		var app *models.App
		err = WithSegment("retrieve-app", c, func() error {
			app, err = models.GetAppByID(application.Db, notifier.AppID)
			return err
		})
		if err != nil {
			log.E(l, "Could not find app.", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Got app from DB")

		log.D(l, "Parse payload")
		var payload notificationPayload
		err = WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Failed to parse json payload.", func(cm log.CM) {
					cm.Write(
						zap.Error(err),
						zap.Duration("duration", time.Now().Sub(start)),
					)
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Parsed payload", func(cm log.CM) {
			cm.Write(zap.Object("payload", payload))
		})

		log.D(l, "Build filters...")
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
		log.D(l, "Built filters successfully", func(cm log.CM) {
			cm.Write(zap.Object("filters", filters))
		})

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
			Config:     application.Config,
			ConfigPath: application.ConfigPath,
			Logger:     l,
			Notifier:   notifier,
			App:        app,
			Message:    message,
			Filters:    filters,
			Modifiers:  modifiers,
		}
		log.D(l, "Get BatchPGWorker...")
		var worker *workers.BatchPGWorker
		err = WithSegment("batchPGWorker-retrieve", c, func() error {
			worker, err = workers.GetBatchPGWorker(workerConfig)
			return err
		})
		if err != nil {
			log.E(l, "Invalid worker config", func(cm log.CM) {
				cm.Write(zap.Error(err), zap.Duration("duration", time.Now().Sub(start)))
			})
			return FailWith(400, err.Error(), c)
		}
		log.D(l, "Got BatchPGWorker...")

		log.D(l, "Start BatchPGWorker...")
		worker.Start()
		log.D(l, "Started BatchPGWorker...")

		return SucceedWith(map[string]interface{}{
			"id": worker.ID.String(),
		}, c)
	}
}

// GetNotifierNotifications is the handler responsible retrieve a notification status
func GetNotifierNotifications(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
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

		log.I(l, "Get from redis", func(cm log.CM) {
			cm.Write(zap.String("redisKey", redisKey))
		})
		var keys []string
		var err error
		err = WithSegment("redis-keys", c, func() error {
			keys, err = cli.Keys(redisKey).Result()
			return err
		})
		if err != nil {
			if err.Error() != "redis: nil" {
				log.E(l, "Failed to get notification status from redis", func(cm log.CM) {
					cm.Write(
						zap.Error(err),
						zap.Duration("duration", time.Now().Sub(start)),
					)
				})
				return FailWith(400, err.Error(), c)
			}
			log.D(l, "No notifications status from redis", func(cm log.CM) {
				cm.Write(zap.Duration("duration", time.Now().Sub(start)))
			})
			return SucceedWith(map[string]interface{}{"statuses": statuses}, c)
		}
		log.I(l, "Got from redis", func(cm log.CM) {
			cm.Write(
				zap.Object("keys", keys),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

		for i := range keys {
			key := string(keys[i])
			var status string
			err = WithSegment("redis-get", c, func() error {
				status, err = cli.Get(key).Result()
				return err
			})
			if err != nil {
				if err.Error() != "redis: nil" {
					log.E(l, "Failed to get notification status from redis", func(cm log.CM) {
						cm.Write(
							zap.Error(err),
							zap.String("key", key),
							zap.Duration("duration", time.Now().Sub(start)),
						)
					})
					return FailWith(400, err.Error(), c)
				}
				log.D(l, "No notifications status from redis", func(cm log.CM) {
					cm.Write(
						zap.String("key", key),
						zap.Duration("duration", time.Now().Sub(start)),
					)
				})
				return SucceedWith(map[string]interface{}{"statuses": statuses}, c)
			}
			var statusObj map[string]interface{}
			err = json.Unmarshal([]byte(status), &statusObj)
			if err != nil {
				return FailWith(500, err.Error(), c)
			}
			statuses = append(statuses, statusObj)
		}

		return SucceedWith(map[string]interface{}{"statuses": statuses}, c)
	}
}

// GetNotifierNotificationStatusHandler is the handler responsible retrieve a notification status
func GetNotifierNotificationStatusHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()
		notifierID := c.Param("notifierID")
		notificationID := c.Param("notificationID")
		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "getNotificationStatus"),
		)

		cli := application.RedisClient.Client
		redisKey := strings.Join([]string{notifierID, notificationID}, "|")
		log.I(l, "Get from redis", func(cm log.CM) {
			cm.Write(zap.String("redisKey", redisKey))
		})
		var status string
		var err error
		err = WithSegment("redis-get", c, func() error {
			status, err = cli.Get(redisKey).Result()
			return err
		})
		if err != nil {
			log.E(l, "Failed to get notification status from redis", func(cm log.CM) {
				cm.Write(
					zap.Error(err),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
			return FailWith(400, err.Error(), c)
		}
		log.I(l, "Got from redis", func(cm log.CM) {
			cm.Write(
				zap.String("key", notificationID),
				zap.String("value", status),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})
		return SucceedWith(map[string]interface{}{"status": status}, c)
	}
}
