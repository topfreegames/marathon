package api

import (
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/kataras/iris"
)

type userTokenPayload struct {
	App     string
	Service string
	UserID  string
	Token   string
	Locale  string
	Region  string
	Tz      string
	BuildN  string
	OptOut  []string
}

func validateUserTokenPayload(payload interface{}) []string {
	var errors []string
	app := GetAsString("App", payload)
	service := GetAsString("Service", payload)
	userID := GetAsString("UserID", payload)
	token := GetAsString("Token", payload)
	locale := GetAsString("Locale", payload)
	region := GetAsString("Region", payload)
	tz := GetAsString("Tz", payload)
	buildN := GetAsString("BuildN", payload)
	optOut := GetAsArray("OptOut", payload)

	if app == "" || len(app) == 0 {
		errors = append(errors, "app is required")
	}
	if service == "" || len(service) == 0 {
		errors = append(errors, "service is required")
	}
	if userID == "" || len(userID) == 0 {
		errors = append(errors, "userID is required")
	}
	if token == "" || len(token) == 0 {
		errors = append(errors, "token is required")
	}
	if locale == "" || len(locale) == 0 {
		errors = append(errors, "locale is required")
	}
	if region == "" || len(region) == 0 {
		errors = append(errors, "region is required")
	}
	if tz == "" || len(tz) == 0 {
		errors = append(errors, "tz is required")
	}
	if buildN == "" || len(buildN) == 0 {
		errors = append(errors, "buildN is required")
	}
	if optOut == nil || len(optOut) == 0 {
		errors = append(errors, "optOut is required")
	}

	return errors
}

func validateUserTokenTablePayload(payload interface{}) []string {
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

// CreateUserTokenHandler is the handler responsible for creating new userTokens
func CreateUserTokenHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload userTokenPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateUserTokenPayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)
		userToken, err := models.UpsertToken(db, payload.App, payload.Service, payload.UserID, payload.Token,
			payload.Locale, payload.Region, payload.Tz, payload.BuildN, payload.OptOut)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"token":   userToken.Token,
			"user_id": userToken.UserID,
			"locale":  userToken.Locale,
			"region":  userToken.Region,
			"tz":      userToken.Tz,
			"buildn":  userToken.BuildN,
		}, c)
	}
}

// CreateUserTokenTableHandler is the handler responsible for creating new userTokens
func CreateUserTokenTableHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload userTokenPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateUserTokenTablePayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)
		userTokenTable, err := models.CreateUserTokensTable(db, payload.App, payload.Service)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"table_name": userTokenTable.TableName,
		}, c)
	}
}
