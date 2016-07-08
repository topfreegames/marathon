package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Models", func() {
	Describe("Notifier", func() {
		var (
			db models.DB
		)
		BeforeEach(func() {
			_db, dbErr := models.GetTestDB()
			Expect(dbErr).To(BeNil())
			Expect(_db).NotTo(BeNil())
			db = _db
		})

		Describe("Create notifier", func() {
			It("Should create a notifier through a factory", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				dbNotifier, dbNotifierErr := models.GetNotifierByID(db, notifier.ID)
				Expect(dbNotifierErr).To(BeNil())
				Expect(dbNotifier.Service).To(Equal(notifier.Service))
				Expect(dbNotifier.AppID).To(Equal(notifier.AppID))
			})

			It("Should create a notifier", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				service := uuid.NewV4().String()[:4]
				appID := app.ID

				createdNotifier, createdNotifierErr := models.CreateNotifier(db, appID, service)
				Expect(createdNotifierErr).To(BeNil())
				dbNotifier, dbNotifierErr := models.GetNotifierByID(db, createdNotifier.ID)
				Expect(dbNotifierErr).To(BeNil())
				Expect(dbNotifier.Service).To(Equal(createdNotifier.Service))
				Expect(dbNotifier.AppID).To(Equal(createdNotifier.AppID))
			})

			It("Should not create a notifier with repeated service,app", func() {
				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())

				service := uuid.NewV4().String()[:4]
				appID := app.ID

				_, createdNotifier1Err := models.CreateNotifier(db, appID, service)
				Expect(createdNotifier1Err).To(BeNil())

				_, createdNotifier2Err := models.CreateNotifier(db, appID, service)
				Expect(createdNotifier2Err).NotTo(BeNil())
			})
		})

		Describe("Update app", func() {
			It("Should update a notifier for an existent id", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())
				service := uuid.NewV4().String()[:4]
				appID := app.ID

				updatedNotifier, updatedNotifierErr := models.UpdateNotifier(db, notifier.ID, appID, service)
				Expect(updatedNotifierErr).To(BeNil())
				dbNotifier, dbNotifierErr := models.GetNotifierByID(db, notifier.ID)
				Expect(dbNotifierErr).To(BeNil())

				Expect(updatedNotifier.Service).To(Equal(dbNotifier.Service))
				Expect(updatedNotifier.AppID).To(Equal(dbNotifier.AppID))
			})

			It("Should not update a notifier with repeated service,app", func() {
				notifier1, notifierErr1 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr1).To(BeNil())
				insertNotifierErr1 := db.Insert(notifier1)
				Expect(insertNotifierErr1).To(BeNil())

				notifier2, notifierErr2 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr2).To(BeNil())
				insertNotifierErr2 := db.Insert(notifier2)
				Expect(insertNotifierErr2).To(BeNil())

				_, updatedNotifierErr := models.UpdateNotifier(db, notifier1.ID, notifier2.AppID, notifier2.Service)
				Expect(updatedNotifierErr).NotTo(BeNil())
			})

			It("Should not update a notifier for an unexistent id", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				app, appErr := CreateAppFactory(db, map[string]interface{}{})
				Expect(appErr).To(BeNil())
				insertAppErr := db.Insert(app)
				Expect(insertAppErr).To(BeNil())
				service := uuid.NewV4().String()[:4]
				appID := app.ID
				invalidID := uuid.NewV4().String()

				_, updatedNotifierErr := models.UpdateNotifier(db, invalidID, appID, service)
				Expect(updatedNotifierErr).NotTo(BeNil())
			})

			It("Should not update a notifier for an unexistent app", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				invalidID := uuid.NewV4().String()

				_, updatedNotifierErr := models.UpdateNotifier(db, notifier.ID, invalidID, notifier.Service)
				Expect(updatedNotifierErr).NotTo(BeNil())
			})
		})

		Describe("Get notifier", func() {
			It("Should retrieve a notifier for an existent id", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				dbNotifier, dbNotifierErr := models.GetNotifierByID(db, notifier.ID)
				Expect(dbNotifierErr).To(BeNil())
				Expect(dbNotifier.Service).To(Equal(notifier.Service))
				Expect(dbNotifier.AppID).To(Equal(notifier.AppID))
			})

			It("Should not retrieve a notifier for an unexistent id", func() {
				invalidID := uuid.NewV4().String()
				_, dbNotifierErr := models.GetNotifierByID(db, invalidID)
				Expect(dbNotifierErr).NotTo(BeNil())
			})

			It("Should retrieve all notifiers for app", func() {
				notifier1, notifierErr1 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr1).To(BeNil())
				insertNotifierErr1 := db.Insert(notifier1)
				Expect(insertNotifierErr1).To(BeNil())

				notifier2Attrs := map[string]interface{}{"AppID": notifier1.AppID}
				notifier2, notifierErr2 := CreateNotifierFactory(db, notifier2Attrs)
				Expect(notifierErr2).To(BeNil())
				insertNotifierErr2 := db.Insert(notifier2)
				Expect(insertNotifierErr2).To(BeNil())

				dbNotifiers, dbNotifiersErr := models.GetNotifiersByApp(db, notifier1.AppID)
				Expect(dbNotifiersErr).To(BeNil())
				Expect(len(dbNotifiers)).To(Equal(2))
				for _, dbNotifier := range dbNotifiers {
					Expect(dbNotifier.AppID).To(Equal(notifier2.AppID))
				}
			})

			It("Should retrieve empty notifiers for app when none matching", func() {
				notifier1, notifierErr1 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr1).To(BeNil())
				insertNotifierErr1 := db.Insert(notifier1)
				Expect(insertNotifierErr1).To(BeNil())

				invalidID := uuid.NewV4().String()

				_, dbNotifiersErr := models.GetNotifiersByApp(db, invalidID)
				Expect(dbNotifiersErr).NotTo(BeNil())
			})

			It("Should retrieve all notifiers for service", func() {
				notifiers := []*models.Notifier{}

				notifier1, notifierErr1 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr1).To(BeNil())
				insertNotifierErr1 := db.Insert(notifier1)
				Expect(insertNotifierErr1).To(BeNil())
				notifiers = append(notifiers, notifier1)

				notifier2Attrs := map[string]interface{}{"Service": notifier1.Service}
				notifier2, notifierErr2 := CreateNotifierFactory(db, notifier2Attrs)
				Expect(notifierErr2).To(BeNil())
				insertNotifierErr2 := db.Insert(notifier2)
				Expect(insertNotifierErr2).To(BeNil())
				notifiers = append(notifiers, notifier2)

				dbNotifiers, dbNotifiersErr := models.GetNotifiersByService(db, notifier1.Service)
				Expect(dbNotifiersErr).To(BeNil())
				Expect(len(dbNotifiers)).To(Equal(2))
				for index, dbNotifier := range dbNotifiers {
					Expect(dbNotifier.AppID).To(Equal(notifiers[index].AppID))
				}
			})

			It("Should retrieve empty notifiers for service when none matching", func() {
				notifier1, notifierErr1 := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr1).To(BeNil())
				insertNotifierErr1 := db.Insert(notifier1)
				Expect(insertNotifierErr1).To(BeNil())

				invalidService := uuid.NewV4().String()[:4]

				_, dbNotifiersErr := models.GetNotifiersByService(db, invalidService)
				Expect(dbNotifiersErr).NotTo(BeNil())
			})

			It("Should retrieve the notifier for app,service", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				dbNotifier, dbNotifiersErr := models.GetNotifierByAppAndService(db, notifier.AppID, notifier.Service)
				Expect(dbNotifiersErr).To(BeNil())
				Expect(dbNotifier.Service).To(Equal(notifier.Service))
				Expect(dbNotifier.AppID).To(Equal(notifier.AppID))
			})

			It("Should not retrieve the notifier for app,service when wrong app", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				invalidApp := uuid.NewV4().String()
				_, dbNotifiersErr := models.GetNotifierByAppAndService(db, invalidApp, notifier.Service)
				Expect(dbNotifiersErr).NotTo(BeNil())
			})

			It("Should not retrieve the notifier for app,service when wrong service", func() {
				notifier, notifierErr := CreateNotifierFactory(db, map[string]interface{}{})
				Expect(notifierErr).To(BeNil())
				insertNotifierErr := db.Insert(notifier)
				Expect(insertNotifierErr).To(BeNil())

				invalidService := uuid.NewV4().String()[:4]
				_, dbNotifiersErr := models.GetNotifierByAppAndService(db, notifier.AppID, invalidService)
				Expect(dbNotifiersErr).NotTo(BeNil())
			})
		})
	})
})
