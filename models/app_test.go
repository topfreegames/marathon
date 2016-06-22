package models

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
)

func TestAppModel(t *testing.T) {
	g := Goblin(t)
	db, dbErr := GetTestDB()
	g.Assert(dbErr).Equal(nil)

	g.Describe("App Model", func() {
		g.Describe("Basic Operations", func() {
			g.It("Should create an app through a factory", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr).Equal(nil)
				insertAppErr := db.Insert(app)
				g.Assert(insertAppErr).Equal(nil)

				dbApp, dbAppErr := GetAppByID(db, app.ID)
				g.Assert(dbAppErr).Equal(nil)
				g.Assert(dbApp.Name).Equal(app.Name)
				g.Assert(dbApp.AppGroup).Equal(app.AppGroup)
				g.Assert(dbApp.OrganizationID).Equal(app.OrganizationID)
			})

			g.It("Should create an app", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				g.Assert(organizationErr).Equal(nil)
				insertOrganizationErr := db.Insert(organization)
				g.Assert(insertOrganizationErr).Equal(nil)

				name := randomdata.SillyName()
				appGroup := randomdata.SillyName()
				organizationID := organization.ID

				createdApp, createdAppErr := CreateApp(db, name, organizationID, appGroup)
				g.Assert(createdAppErr).Equal(nil)
				dbApp, dbAppErr := GetAppByID(db, createdApp.ID)
				g.Assert(dbAppErr).Equal(nil)
				g.Assert(dbApp.Name).Equal(createdApp.Name)
				g.Assert(dbApp.AppGroup).Equal(createdApp.AppGroup)
				g.Assert(dbApp.OrganizationID).Equal(createdApp.OrganizationID)
			})
		})
	})
}
