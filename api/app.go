package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

type appPayload struct {
	AppName        string
	Service        string
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

		l.Debug("Creating app...")
		app, err := models.CreateApp(application.Db, payload.AppName, payload.OrganizationID, payload.AppGroup)
		if err != nil {
			l.Error("Create app failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}
		l.Info(
			"App created successfully.",
			zap.String("id", app.ID.String()),
			zap.String("name", app.Name),
			zap.String("group", app.AppGroup),
			zap.String("organization_id", app.OrganizationID.String()),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		l.Debug("Creating notifier...")
		notifier, err := models.CreateNotifier(application.Db, app.ID, payload.Service)
		if err != nil {
			l.Error("Create notifier failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}
		l.Info(
			"Notifier created successfully.",
			zap.String("id", notifier.ID.String()),
			zap.String("appID", notifier.AppID.String()),
			zap.String("service", notifier.Service),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		userTokensTable, err := models.CreateUserTokensTable(application.Db, payload.AppName, payload.Service)
		if err != nil {
			l.Error("Create app failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		l.Info(
			"UserToken table created successfully.",
			zap.String("id", app.ID.String()),
			zap.String("name", app.Name),
			zap.String("group", app.AppGroup),
			zap.String("organization_id", app.OrganizationID.String()),
			zap.Object("userTokensTableName", userTokensTable.TableName),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"id":                  app.ID,
			"appName":             app.Name,
			"appGroup":            app.AppGroup,
			"organizationID":      app.OrganizationID,
			"service":             notifier.Service,
			"notifierID":          notifier.ID,
			"userTokensTableName": userTokensTable.TableName,
		}, c)
	}
}

// GetAppsHandler is the handler responsible for retrieaving a list of apps/services
func GetAppsHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "appHandler"),
			zap.String("operation", "getApps"),
		)

		l.Debug("Getting app...")
		appNotifiers, err := models.GetAppNotifiers(application.Db)
		if err != nil {
			l.Error("Get apps notifiers failed.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}
		l.Info(
			"Apps notifiers retrieved successfully.",
			zap.Int("qty", len(appNotifiers)),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"apps": serializeAppsNotifiers(appNotifiers),
		}, c)
	}
}
