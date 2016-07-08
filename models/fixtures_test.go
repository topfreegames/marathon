package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/bluele/factory-go/factory"
	"github.com/satori/go.uuid"
)

// OrganizationFactory is responsible for constructing test organization instances
var OrganizationFactory = factory.NewFactory(
	&models.Organization{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
})

// CreateOrganizationFactory is responsible for constructing test organization instances
func CreateOrganizationFactory(db models.DB, attrs map[string]interface{}) (*models.Organization, error) {
	organization := OrganizationFactory.MustCreateWithOption(attrs).(*models.Organization)
	return organization, nil
}

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
		organization := OrganizationFactory.MustCreateWithOption(map[string]interface{}{}).(*models.Organization)
		insertOrganizationErr := db.Insert(organization)
		if insertOrganizationErr != nil {
			return nil, insertOrganizationErr
		}
		attrs["OrganizationID"] = organization.ID
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
