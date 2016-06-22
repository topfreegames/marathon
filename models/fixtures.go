package models

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
)

// OrganizationFactory is responsible for constructing test organization instances
var OrganizationFactory = factory.NewFactory(
	&Organization{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
})

// CreateOrganizationFactory is responsible for constructing test organization instances
func CreateOrganizationFactory(db DB, attrs map[string]interface{}) (*Organization, error) {
	organization := OrganizationFactory.MustCreateWithOption(attrs).(*Organization)
	return organization, nil
}

// AppFactory is responsible for constructing test app instances
var AppFactory = factory.NewFactory(
	&App{},
).Attr("Name", func(args factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("AppGroup", func(args factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
})

// CreateAppFactory is responsible for constructing test app instances
func CreateAppFactory(db DB, attrs map[string]interface{}) (*App, error) {
	if attrs["OrganizationID"] == nil {
		organization := OrganizationFactory.MustCreateWithOption(map[string]interface{}{}).(*Organization)
		insertOrganizationErr := db.Insert(organization)
		if insertOrganizationErr != nil {
			return nil, insertOrganizationErr
		}
		attrs["OrganizationID"] = organization.ID
	}
	app := AppFactory.MustCreateWithOption(attrs).(*App)
	return app, nil
}
