package api

import (
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
)

type appPayload struct {
	Name           string
	OrganizationID uuid.UUID
	AppGroup       string
}

func validateAppPayload(payload interface{}) []string {
	var errors []string
	name := GetAsString("Name", payload)
	_, errUUID := GetAsUUID("OrganizationID", payload)
	appGroup := GetAsString("AppGroup", payload)

	if name == "" || len(name) == 0 {
		errors = append(errors, "name is required")
	}
	if errUUID != nil {
		errors = append(errors, "organizationID is required")
	}
	if appGroup == "" || len(appGroup) == 0 {
		errors = append(errors, "appGroup is required")
	}
	return errors
}

// CreateAppHandler is the handler responsible for creating new apps
func CreateAppHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload appPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateAppPayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)
		app, err := models.CreateApp(db, payload.Name, payload.OrganizationID, payload.AppGroup)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"name":            app.Name,
			"app_group":       app.AppGroup,
			"organization_id": app.OrganizationID,
		}, c)
	}
}
