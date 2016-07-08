package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
)

// OrganizationFactory is responsible for constructing test organization instances
var OrganizationFactory = factory.NewFactory(
	&models.Organization{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
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
	return randomdata.SillyName(), nil
}).Attr("AppGroup", func(args factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
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
