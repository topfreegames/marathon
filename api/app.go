package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

type appPayload struct {
	Name           string
	OrganizationID uuid.UUID
	AppGroup       string
}

// CreateAppHandler is the handler responsible for creating new apps
func CreateAppHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "appHandler"),
			zap.String("operation", "createApp"),
		)

		var payload appPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error("Failed to parse json payload.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Creating app...")
		app, err := models.CreateApp(db, payload.Name, payload.OrganizationID, payload.AppGroup)
		if err != nil {
			l.Error("Create app failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		l.Info(
			"App created successfully.",
			zap.String("id", app.ID.String()),
			zap.String("name", app.Name),
			zap.String("app_group", app.AppGroup),
			zap.String("organization_id", app.OrganizationID.String()),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"id":              app.ID,
			"name":            app.Name,
			"app_group":       app.AppGroup,
			"organization_id": app.OrganizationID,
		}, c)
	}
}
