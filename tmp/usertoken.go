package controllers

import (
	"reflect"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/api"
	"github.com/kataras/iris"
)

type appPayload struct {
	App    string
	Locale string
	Region string
	Tz     string
	BuildN string
	OptOut []string
}

type appPayload struct {
	Token  string
	Locale string
	Region string
	Tz     string
	BuildN string
	OptOut []string
}

func getAsInt(field string, payload interface{}) int {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(int)
}

func getAsJSON(field string, payload interface{}) map[string]interface{} {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(map[string]interface{})
}

func getAsArray(field string, payload interface{}) []string {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.([]string)
}

func getAsString(field string, payload interface{}) string {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(string)
}

func validateAppPayload(payload interface{}) []string {
	var errors []string
	token := getAsString("Token", payload)
	locale := getAsString("Locale", payload)
	region := getAsString("Region", payload)
	tz := getAsString("Tz", payload)
	buildN := getAsString("BuildN", payload)

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
	return errors
}

// CreateAppHandler is the handler responsible for creating new apps
func CreateAppHandler(application *api.Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload appPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateGamePayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		db := GetCtxDB(c)

		scope := db.NewScope(models.Token)
		scope.Search.tableName =
			db.CreateTable(models.Token)
		// game, err := models.CreateGame(
		// 	db,
		// 	payload.Token,
		// 	payload.Locale,
		// 	payload.Region,
		// 	payload.Tz,
		// 	payload.BuildN,
		// 	payload.OptOut,
		// )

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": game.PublicID,
		}, c)
	}
}
