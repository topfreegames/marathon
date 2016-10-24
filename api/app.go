package api

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

type appPayload struct {
	AppName        string
	Service        string
	OrganizationID uuid.UUID
	AppGroup       string
}

// CreateAppHandler is the handler responsible for creating new apps
func CreateAppHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "appHandler"),
			zap.String("operation", "createApp"),
		)

		var payload appPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			log.E(l, "Failed to parse json payload.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		// FIXME: This should not work this way. We're ignoring organizationID and appGroup if appName exists
		log.D(l, "Creating app...")
		app, err := models.CreateApp(application.Db, payload.AppName, payload.OrganizationID, payload.AppGroup)
		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"index_apps_on_name\"" {
				app, err = models.GetAppByName(application.Db, payload.AppName)
				if err != nil {
					log.E(l, "Get app failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					return FailWith(400, err.Error(), c)
				}
				log.I(l, "App not created. Already exists", func(cm log.CM) {
					cm.Write(
						zap.String("id", app.ID.String()),
						zap.String("name", app.Name),
						zap.String("group", app.AppGroup),
						zap.String("organization_id", app.OrganizationID.String()),
						zap.Duration("duration", time.Now().Sub(start)),
					)
				})
			} else {
				log.E(l, "Create app failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(400, err.Error(), c)
			}
		} else {
			log.I(l, "App created successfully.", func(cm log.CM) {
				cm.Write(
					zap.String("id", app.ID.String()),
					zap.String("name", app.Name),
					zap.String("group", app.AppGroup),
					zap.String("organization_id", app.OrganizationID.String()),
					zap.Duration("duration", time.Now().Sub(start)),
				)
			})
		}

		log.D(l, "Creating notifier...")
		notifier, err := models.CreateNotifier(application.Db, app.ID, payload.Service)
		if err != nil {
			log.E(l, "Create notifier failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}
		log.I(l, "Notifier created successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("id", notifier.ID.String()),
				zap.String("appID", notifier.AppID.String()),
				zap.String("service", notifier.Service),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

		userTokensTable, err := models.CreateUserTokensTable(application.Db, payload.AppName, payload.Service)
		if err != nil {
			log.E(l, "Create app failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		log.I(l, "UserToken table created successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("id", app.ID.String()),
				zap.String("name", app.Name),
				zap.String("group", app.AppGroup),
				zap.String("organization_id", app.OrganizationID.String()),
				zap.Object("userTokensTableName", userTokensTable.TableName),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

		return SucceedWith(map[string]interface{}{
			"appID":               app.ID,
			"appName":             app.Name,
			"appGroup":            app.AppGroup,
			"organizationID":      app.OrganizationID,
			"notifierService":     notifier.Service,
			"notifierID":          notifier.ID,
			"userTokensTableName": userTokensTable.TableName,
		}, c)
	}
}

// GetAppsHandler is the handler responsible for retrieving a list of apps/services
func GetAppsHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "appHandler"),
			zap.String("operation", "getApps"),
		)

		log.D(l, "Getting app...")
		appNotifiers, err := models.GetAppNotifiers(application.Db)
		if err != nil {
			log.E(l, "Get apps notifiers failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}
		log.I(l, "Apps notifiers retrieved successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("qty", len(appNotifiers)),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

		return SucceedWith(map[string]interface{}{
			"apps": serializeAppsNotifiers(appNotifiers),
		}, c)
	}
}
