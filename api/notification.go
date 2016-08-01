package api

import (
	"strings"

	"github.com/kataras/iris"
)

// TODO: Add: userTime, ts, pushExpiry
type notificationPayload struct {
	App      string
	Service  string
	Locale   string
	Region   string
	Tz       string
	BuildN   string
	Scope    string
	Template string
	Params   map[string]interface{}
	Message  map[string]interface{}
	Metadata map[string]interface{}
}

func validateNotificationPayload(payload interface{}) []string {
	var errors []string
	app := GetAsString("App", payload)
	service := GetAsString("Service", payload)

	if app == "" || len(app) == 0 {
		errors = append(errors, "app is required")
	}
	if service == "" || len(service) == 0 {
		errors = append(errors, "service is required")
	}

	// locale := GetAsString("Locale", payload)
	// region := GetAsString("region", payload)
	// tz := GetAsString("tz", payload)
	// buildN := GetAsString("BuildN", payload)
	// scope := GetAsString("Scope", payload)
	// localeDefined := locale == "" || len(locale) == 0
	// regionDefined := region == "" || len(region) == 0
	// tzDefined := tz == "" || len(tz) == 0
	// buildNDefined := buildN == "" || len(buildN) == 0
	// scopeDefined := scope == "" || len(scope) == 0

	return errors
}

// CreateNotificationHandler is the handler responsible for creating new organizations
func CreateNotificationHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload notifierPayload
		_, err := LoadJSONPayload(&payload, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}
		if payloadErrors := validateNotificationPayload(payload); len(payloadErrors) != 0 {
			errorString := strings.Join(payloadErrors[:], ", ")
			FailWith(422, errorString, c)
			return
		}

		// db := GetCtxDB(c)

		// GetUserTokenBatchByFilters
		// notifier, err := models.LaunchNotification(db, payload.AppID, payload.Service)
		// if err != nil {
		// 	FailWith(500, err.Error(), c)
		// 	return
		// }
		// application.Logger.Info("Notifier created", zap.Object("notifier", notifier))
		//
		// SucceedWith(map[string]interface{}{
		// 	"id": notifier.ID,
		// }, c)
	}
}
