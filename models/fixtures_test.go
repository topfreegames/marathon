package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/bluele/factory-go/factory"
	"github.com/satori/go.uuid"
)

// AppFactory is responsible for constructing test app instances
var AppFactory = factory.NewFactory(
	&models.App{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("AppGroup", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
})

// CreateAppFactory is responsible for constructing test app instances
func CreateAppFactory(db models.DB, attrs map[string]interface{}) (*models.App, error) {
	if attrs["OrganizationID"] == nil {
		attrs["OrganizationID"] = uuid.NewV4()
	}
	app := AppFactory.MustCreateWithOption(attrs).(*models.App)
	return app, nil
}

// NotifierFactory is responsible for constructing test notifier instances
var NotifierFactory = factory.NewFactory(
	&models.Notifier{},
).Attr("Service", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String()[:4], nil
})

// CreateNotifierFactory is responsible for constructing test notifier instances
func CreateNotifierFactory(db models.DB, attrs map[string]interface{}) (*models.Notifier, error) {
	if attrs["AppID"] == nil {
		app, appErr := CreateAppFactory(db, map[string]interface{}{})
		if appErr != nil {
			return nil, appErr
		}
		insertAppErr := db.Insert(app)
		if insertAppErr != nil {
			return nil, insertAppErr
		}
		attrs["AppID"] = app.ID
	}
	notifier := NotifierFactory.MustCreateWithOption(attrs).(*models.Notifier)
	return notifier, nil
}

var TemplateFactory = factory.NewFactory(
	&models.Template{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Service", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String()[:4], nil
}).Attr("Locale", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String()[:2], nil
}).Attr("Defaults", func(args factory.Args) (interface{}, error) {
	return map[string]interface{}{"username": "banduk"}, nil
}).Attr("Body", func(args factory.Args) (interface{}, error) {
	return map[string]interface{}{"alert": "{{username}} sent you a message."}, nil
})

// CreateTemplateFactory is responsible for constructing test template instances
func CreateTemplateFactory(db models.DB, attrs map[string]interface{}) (*models.Template, error) {
	if attrs["AppID"] == nil {
		app, appErr := CreateAppFactory(db, map[string]interface{}{})
		if appErr != nil {
			return nil, appErr
		}
		insertAppErr := db.Insert(app)
		if insertAppErr != nil {
			return nil, insertAppErr
		}
		attrs["AppID"] = app.ID
	}
	template := TemplateFactory.MustCreateWithOption(attrs).(*models.Template)
	return template, nil
}

var UserTokenFactory = factory.NewFactory(
	&models.UserToken{},
).Attr("Token", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("UserID", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Locale", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String()[:2], nil
}).Attr("Region", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String()[:2], nil
}).Attr("Tz", func(args factory.Args) (interface{}, error) {
	return "GMT+03:00", nil
}).Attr("BuildN", func(args factory.Args) (interface{}, error) {
	return "30000", nil
}).Attr("OptOut", func(args factory.Args) (interface{}, error) {
	return []string{uuid.NewV4().String(), uuid.NewV4().String()}, nil
})

func CreateUserTokenFactory(db models.DB, attrs map[string]interface{}) (*models.UserToken, error) {
	userToken := UserTokenFactory.MustCreateWithOption(attrs).(*models.UserToken)
	return userToken, nil
}
