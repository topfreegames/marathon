package api

import "git.topfreegames.com/topfreegames/marathon/models"

func serializeAppsNotifiers(appsNotifiers []models.AppNotifier) []map[string]interface{} {
	serializedApps := make([]map[string]interface{}, len(appsNotifiers))
	for i, appNotifier := range appsNotifiers {
		serializedApps[i] = serializeAppNotifier(&appNotifier)
	}

	return serializedApps
}

func serializeAppNotifier(appNotifier *models.AppNotifier) map[string]interface{} {
	serial := map[string]interface{}{
		"id":         appNotifier.AppID,
		"appName":    appNotifier.AppName,
		"appGroup":   appNotifier.AppGroup,
		"service":    appNotifier.NotifierService,
		"notifierID": appNotifier.NotifierID,
	}

	return serial
}
