package models

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	"github.com/satori/go.uuid"
)

func TestAppModel(t *testing.T) {
	g := Goblin(t)
	db, dbErr := GetTestDB()
	g.Assert(dbErr).Equal(nil)

	g.Describe("App Model", func() {
		g.Describe("Create app", func() {
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

			g.It("Should not create an app with repeated name", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				g.Assert(organizationErr).Equal(nil)
				insertOrganizationErr := db.Insert(organization)
				g.Assert(insertOrganizationErr).Equal(nil)

				name1 := randomdata.SillyName()
				appGroup1 := randomdata.SillyName()
				organizationID1 := organization.ID

				createdApp, createdAppErr := CreateApp(db, name1, organizationID1, appGroup1)
				g.Assert(createdAppErr).Equal(nil)
				_, dbAppErr1 := GetAppByID(db, createdApp.ID)
				g.Assert(dbAppErr1).Equal(nil)

				name2 := name1
				appGroup2 := randomdata.SillyName()
				organizationID2 := organization.ID

				_, createdAppErr2 := CreateApp(db, name2, organizationID2, appGroup2)
				g.Assert(createdAppErr2 != nil).IsTrue()
			})
		})

		g.Describe("Update app", func() {
			g.It("Should update an app for an existent id", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr).Equal(nil)
				insertAppErr := db.Insert(app)
				g.Assert(insertAppErr).Equal(nil)

				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				g.Assert(organizationErr).Equal(nil)
				insertOrganizationErr := db.Insert(organization)
				g.Assert(insertOrganizationErr).Equal(nil)

				newName := randomdata.SillyName()
				newOrganizationID := organization.ID
				newAppGroup := randomdata.SillyName()

				updatedApp, updatedAppErr := UpdateApp(db, app.ID, newName, newOrganizationID, newAppGroup)
				g.Assert(updatedAppErr).Equal(nil)
				dbApp, dbAppErr := GetAppByID(db, app.ID)
				g.Assert(dbAppErr).Equal(nil)

				g.Assert(updatedApp.Name).Equal(dbApp.Name)
				g.Assert(updatedApp.AppGroup).Equal(dbApp.AppGroup)
				g.Assert(updatedApp.OrganizationID).Equal(dbApp.OrganizationID)
			})

			g.It("Should not update an app with repeated name", func() {
				app1, appErr1 := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr1).Equal(nil)
				insertAppErr1 := db.Insert(app1)
				g.Assert(insertAppErr1).Equal(nil)

				app2, appErr2 := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr2).Equal(nil)
				insertAppErr2 := db.Insert(app2)
				g.Assert(insertAppErr2).Equal(nil)

				newName := app1.Name
				newOrganizationID := app2.OrganizationID
				newAppGroup := app2.AppGroup

				_, updatedAppErr := UpdateApp(db, app2.ID, newName, newOrganizationID, newAppGroup)
				g.Assert(updatedAppErr != nil).IsTrue()
			})

			g.It("Should not update an app for an unexistent id", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				g.Assert(organizationErr).Equal(nil)
				insertOrganizationErr := db.Insert(organization)
				g.Assert(insertOrganizationErr).Equal(nil)

				newName := randomdata.SillyName()
				newOrganizationID := organization.ID
				newAppGroup := randomdata.SillyName()

				invalidID := uuid.NewV4().String()
				_, updatedAppErr := UpdateApp(db, invalidID, newName, newOrganizationID, newAppGroup)
				g.Assert(updatedAppErr != nil).IsTrue()
			})
		})

		g.Describe("Get app", func() {
			g.It("Should retrieve an app for an existent id", func() {
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

			g.It("Should not retrieve an app for an unexistent id", func() {
				invalidID := uuid.NewV4().String()
				_, dbAppErr := GetAppByID(db, invalidID)
				g.Assert(dbAppErr != nil).IsTrue()
			})

			g.It("Should retrieve an app for an existent name", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr).Equal(nil)
				insertAppErr := db.Insert(app)
				g.Assert(insertAppErr).Equal(nil)

				dbApp, dbAppErr := GetAppByName(db, app.Name)
				g.Assert(dbAppErr).Equal(nil)
				g.Assert(dbApp.Name).Equal(app.Name)
				g.Assert(dbApp.AppGroup).Equal(app.AppGroup)
				g.Assert(dbApp.OrganizationID).Equal(app.OrganizationID)
			})

			g.It("Should not retrieve an app for an unexistent name", func() {
				invalidName := randomdata.SillyName()
				_, dbAppErr := GetAppByName(db, invalidName)
				g.Assert(dbAppErr != nil).IsTrue()
			})

			g.It("Should retrieve all apps for an existent group", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				g.Assert(appErr).Equal(nil)
				insertAppErr := db.Insert(app)
				g.Assert(insertAppErr).Equal(nil)

				dbApps, dbAppsErr := GetAppsByGroup(db, app.AppGroup)
				g.Assert(dbAppsErr).Equal(nil)
				for eachApp := range dbApps {
					g.Assert(dbApps[eachApp].Name).Equal(app.Name)
					g.Assert(dbApps[eachApp].AppGroup).Equal(app.AppGroup)
					g.Assert(dbApps[eachApp].OrganizationID).Equal(app.OrganizationID)
				}
			})

			g.It("Should not retrieve an app for an unexistent group", func() {
				invalidGroup := randomdata.SillyName()
				_, dbAppsErr := GetAppsByGroup(db, invalidGroup)
				g.Assert(dbAppsErr != nil).IsTrue()
			})
		})
	})
}
