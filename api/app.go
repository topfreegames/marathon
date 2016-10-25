package api

import (
	"time"

	gorp "gopkg.in/gorp.v1"

	"git.topfreegames.com/topfreegames/marathon/models"

	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/labstack/echo"
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
func CreateAppHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		start := time.Now()

		l := application.Logger.With(
			zap.String("source", "appHandler"),
			zap.String("operation", "createApp"),
		)

		var payload appPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Failed to parse json payload.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})

		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		// FIXME: This should not work this way. We're ignoring organizationID and appGroup if appName exists
		log.D(l, "Creating app...")
		var app *models.App
		err = WithSegment("app-create", c, func() error {
			app, err = models.CreateApp(application.Db, payload.AppName, payload.OrganizationID, payload.AppGroup)
			return err
		})

		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"index_apps_on_name\"" {
				err = WithSegment("app-retrive-by-name", c, func() error {
					app, err = models.GetAppByName(application.Db, payload.AppName)
					return err
				})
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
		}

		log.I(l, "App created successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("id", app.ID.String()),
				zap.String("name", app.Name),
				zap.String("group", app.AppGroup),
				zap.String("organization_id", app.OrganizationID.String()),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

		log.D(l, "Creating notifier...")
		var notifier *models.Notifier
		err = WithSegment("notifier-create", c, func() error {
			notifier, err = models.CreateNotifier(application.Db, app.ID, payload.Service)
			return err
		})
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

		var userTokensTable *gorp.TableMap
		err = WithSegment("userTokensTable-create", c, func() error {
			userTokensTable, err = models.CreateUserTokensTable(application.Db, payload.AppName, payload.Service)
			return err
		})
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
		var appNotifiers []models.AppNotifier
		var err error
		err = WithSegment("appNotifiers-retrieve", c, func() error {
			appNotifiers, err = models.GetAppNotifiers(application.Db)
			return err
		})
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

		return WithSegment("response-serialize", c, func() error {
			return SucceedWith(map[string]interface{}{
				"apps": serializeAppsNotifiers(appNotifiers),
			}, c)
		})
	}
}
