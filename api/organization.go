package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

type organizationPayload struct {
	Name string
}

// CreateOrganizationHandler is the handler responsible for creating new organizations
func CreateOrganizationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "organizationHandler"),
			zap.String("operation", "createOrganization"),
		)

		var payload organizationPayload
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

		l.Debug("Creating organization...")
		organization, err := models.CreateOrganization(db, payload.Name)
		if err != nil {
			l.Error("Create organization failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		l.Info(
			"Organization created successfully.",
			zap.String("organizationId", organization.ID.String()),
			zap.String("organizationName", organization.Name),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"name": organization.Name,
			"id":   organization.ID,
		}, c)
	}
}
