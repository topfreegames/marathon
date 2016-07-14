package api

import (
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

type organizationPayload struct {
	Name string
}

func validateOrganizationPayload(payload interface{}) []string {
	var errors []string
	name := GetAsString("Name", payload)

	if name == "" || len(name) == 0 {
		errors = append(errors, "name is required")
	}

	return errors
}

// CreateOrganizationHandler is the handler responsible for creating new organizations
func CreateOrganizationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload organizationPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateOrganizationPayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)

		organization, err := models.CreateOrganization(db, payload.Name)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}
		application.Logger.Info("Organization created", zap.Object("organization", organization))

		SucceedWith(map[string]interface{}{
			"id": organization.ID,
		}, c)
	}
}
