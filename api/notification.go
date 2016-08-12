package api

import (
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/util"
	"git.topfreegames.com/topfreegames/marathon/workers"

	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

// TODO: Add: userTime, ts, pushExpiry
type notificationPayload struct {
	App      string
	Service  string
	PageSize int
	Filters  map[string]interface{}
	Template string
	Params   map[string]interface{}
	Message  map[string]interface{}
	Metadata map[string]interface{}
}

// SendNotificationHandler is the handler responsible for creating new apps
func SendNotificationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		// start := time.Now()
		// appName := c.Param("appName")

		l := application.Logger.With(
			zap.String("source", "notificationHandler"),
			zap.String("operation", "sendNotification"),
		)

		var payload notificationPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error("Failed to parse json payload.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		allowedFilters := []string{"locale", "region", "tz", "build_n", "scope"}
		l.Debug("Allowed filters", zap.Object("allowedFilters", allowedFilters))

		filters := [][]interface{}{}
		for k, v := range payload.Filters {
			if util.SliceContains(allowedFilters, k) {
				filters = append(filters, []interface{}{k, v})
			}
		}
		l.Debug("Built filters", zap.Object("filters", filters))

		// TODO: Should we accept as parameters ?
		modifiers := [][]interface{}{{"LIMIT", payload.PageSize}}

		workerConfig := &workers.BatchPGWorker{
			ConfigPath: application.ConfigPath,
			Logger:     l,
		}
		worker, err := workers.GetBatchPGWorker(workerConfig)
		if err != nil {
			l.Error("Invalid worker config,", zap.Error(err))
			FailWith(400, err.Error(), c)
		}

		message := &messages.InputMessage{
			App:      payload.App,
			Service:  payload.Service,
			Template: payload.Template,
			Params:   payload.Params,
			Message:  payload.Message,
			Metadata: payload.Metadata,
		}
		worker.StartWorker(message, filters, modifiers)
		SucceedWith(map[string]interface{}{}, c)
	}
}
