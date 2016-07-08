package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Models", func() {
	Describe("App", func() {
		var (
			db models.DB
		)
		BeforeEach(func() {
			_db, dbErr := models.GetTestDB()
			Expect(dbErr).To(BeNil())
			Expect(_db).NotTo(BeNil())
			db = _db
		})

		Describe("Create app", func() {
			It("Should create an app through a factory", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				dbApp, dbAppErr := models.GetAppByID(db, app.ID)
				Expect(dbAppErr).To(BeNil())
				Expect(dbApp.Name).To(Equal(app.Name))
				Expect(dbApp.AppGroup).To(Equal(app.AppGroup))
				Expect(dbApp.OrganizationID).To(Equal(app.OrganizationID))
			})

			It("Should create an app", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				Expect(organizationErr).To(BeNil())
				insertOrganizationErr := db.Insert(organization)
				Expect(insertOrganizationErr).To(BeNil())

				name := uuid.NewV4().String()
				appGroup := uuid.NewV4().String()
				organizationID := organization.ID

				createdApp, createdAppErr := models.CreateApp(db, name, organizationID, appGroup)
				Expect(createdAppErr).To(BeNil())
				dbApp, dbAppErr := models.GetAppByID(db, createdApp.ID)
				Expect(dbAppErr).To(BeNil())
				Expect(dbApp.Name).To(Equal(createdApp.Name))
				Expect(dbApp.AppGroup).To(Equal(createdApp.AppGroup))
				Expect(dbApp.OrganizationID).To(Equal(createdApp.OrganizationID))
			})

			It("Should not create an app with repeated name", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				Expect(organizationErr).To(BeNil())
				insertOrganizationErr := db.Insert(organization)
				Expect(insertOrganizationErr).To(BeNil())

				name1 := uuid.NewV4().String()
				appGroup1 := uuid.NewV4().String()
				organizationID1 := organization.ID

				createdApp, createdAppErr := models.CreateApp(db, name1, organizationID1, appGroup1)
				Expect(createdAppErr).To(BeNil())
				_, dbAppErr1 := models.GetAppByID(db, createdApp.ID)
				Expect(dbAppErr1).To(BeNil())

				name2 := name1
				appGroup2 := uuid.NewV4().String()
				organizationID2 := organization.ID

				_, createdAppErr2 := models.CreateApp(db, name2, organizationID2, appGroup2)
				Expect(createdAppErr2).NotTo(BeNil())
			})

			It("Should not create an app with invelid organization", func() {
				name := uuid.NewV4().String()
				appGroup := uuid.NewV4().String()
				invalidID := uuid.NewV4().String()
				_, createdAppErr := models.CreateApp(db, name, invalidID, appGroup)
				Expect(createdAppErr).NotTo(BeNil())
			})
		})

		Describe("Update app", func() {
			It("Should update an app for an existent id", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				Expect(organizationErr).To(BeNil())
				insertOrganizationErr := db.Insert(organization)
				Expect(insertOrganizationErr).To(BeNil())

				newName := uuid.NewV4().String()
				newOrganizationID := organization.ID
				newAppGroup := uuid.NewV4().String()

				updatedApp, updatedAppErr := models.UpdateApp(db, app.ID, newName, newOrganizationID, newAppGroup)
				Expect(updatedAppErr).To(BeNil())
				dbApp, dbAppErr := models.GetAppByID(db, app.ID)
				Expect(dbAppErr).To(BeNil())

				Expect(updatedApp.Name).To(Equal(dbApp.Name))
				Expect(updatedApp.AppGroup).To(Equal(dbApp.AppGroup))
				Expect(updatedApp.OrganizationID).To(Equal(dbApp.OrganizationID))
			})

			It("Should not update an app with repeated name", func() {
				app1, appErr1 := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr1).To(BeNil())
				insertAppErr1 := db.Insert(app1)
				Expect(insertAppErr1).To(BeNil())

				app2, appErr2 := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr2).To(BeNil())
				insertAppErr2 := db.Insert(app2)
				Expect(insertAppErr2).To(BeNil())

				newName := app1.Name
				newOrganizationID := app2.OrganizationID
				newAppGroup := app2.AppGroup

				_, updatedAppErr := models.UpdateApp(db, app2.ID, newName, newOrganizationID, newAppGroup)
				Expect(updatedAppErr).NotTo(BeNil())
			})

			It("Should not update an app for an unexistent id", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				Expect(organizationErr).To(BeNil())
				insertOrganizationErr := db.Insert(organization)
				Expect(insertOrganizationErr).To(BeNil())

				newName := uuid.NewV4().String()
				newOrganizationID := organization.ID
				newAppGroup := uuid.NewV4().String()

				invalidID := uuid.NewV4().String()
				_, updatedAppErr := models.UpdateApp(db, invalidID, newName, newOrganizationID, newAppGroup)
				Expect(updatedAppErr).NotTo(BeNil())
			})
		})

		Describe("Get app", func() {
			It("Should retrieve an app for an existent id", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				dbApp, dbAppErr := models.GetAppByID(db, app.ID)
				Expect(dbAppErr).To(BeNil())
				Expect(dbApp.Name).To(Equal(app.Name))
				Expect(dbApp.AppGroup).To(Equal(app.AppGroup))
				Expect(dbApp.OrganizationID).To(Equal(app.OrganizationID))
			})

			It("Should not retrieve an app for an unexistent id", func() {
				invalidID := uuid.NewV4().String()
				_, dbAppErr := models.GetAppByID(db, invalidID)
				Expect(dbAppErr).NotTo(BeNil())
			})

			It("Should retrieve an app for an existent name", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				dbApp, dbAppErr := models.GetAppByName(db, app.Name)
				Expect(dbAppErr).To(BeNil())
				Expect(dbApp.Name).To(Equal(app.Name))
				Expect(dbApp.AppGroup).To(Equal(app.AppGroup))
				Expect(dbApp.OrganizationID).To(Equal(app.OrganizationID))
			})

			It("Should not retrieve an app for an unexistent name", func() {
				invalidName := uuid.NewV4().String()
				_, dbAppErr := models.GetAppByName(db, invalidName)
				Expect(dbAppErr).NotTo(BeNil())
			})

			It("Should retrieve all apps for an existent group", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				dbApps, dbAppsErr := models.GetAppsByGroup(db, app.AppGroup)
				Expect(dbAppsErr).To(BeNil())
				for eachApp := range dbApps {
					Expect(dbApps[eachApp].Name).To(Equal(app.Name))
					Expect(dbApps[eachApp].AppGroup).To(Equal(app.AppGroup))
					Expect(dbApps[eachApp].OrganizationID).To(Equal(app.OrganizationID))
				}
			})

			It("Should not retrieve an app for an unexistent group", func() {
				invalidGroup := uuid.NewV4().String()
				_, dbAppsErr := models.GetAppsByGroup(db, invalidGroup)
				Expect(dbAppsErr).NotTo(BeNil())
			})
		})
	})
})
