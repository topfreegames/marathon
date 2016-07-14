package api

import (
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

type notifierPayload struct {
	AppID   uuid.UUID
	Service string
}

func validateNotifierPayload(payload interface{}) []string {
	var errors []string
	service := GetAsString("Service", payload)
	_, errUUID := GetAsUUID("OrganizationID", payload)

	if service == "" || len(service) == 0 {
		errors = append(errors, "service is required")
	}
	if errUUID != nil {
		errors = append(errors, "appID is required")
	}
	return errors
}

// CreateNotifierHandler is the handler responsible for creating new organizations
func CreateNotifierHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload notifierPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateNotifierPayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)

		notifier, err := models.CreateNotifier(db, payload.AppID, payload.Service)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}
		application.Logger.Info("Notifier created", zap.Object("notifier", notifier))

		SucceedWith(map[string]interface{}{
			"id": notifier.ID,
		}, c)
	}
}
