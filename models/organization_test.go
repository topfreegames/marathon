package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Models", func() {
	var (
		db models.DB
	)
	BeforeEach(func() {
		_db, dbErr := models.GetTestDB()
		Expect(dbErr).To(BeNil())
		Expect(_db).NotTo(BeNil())
		db = _db
	})

	Describe("Organization", func() {
		Describe("Create Organization", func() {
			It("Should create an organization through a factory", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				Expect(organizationErr).To(BeNil())
				insertOrganizationErr := db.Insert(organization)
				Expect(insertOrganizationErr).To(BeNil())

				dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, organization.ID)
				Expect(dbOrganizationErr).To(BeNil())
				Expect(dbOrganization.Name).To(Equal(organization.Name))
			})

			It("Should create an organization", func() {
				name := uuid.NewV4().String()

				createdOrganization, createdOrganizationErr := models.CreateOrganization(db, name)
				Expect(createdOrganizationErr).To(BeNil())

				dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, createdOrganization.ID)
				Expect(dbOrganizationErr).To(BeNil())
				Expect(dbOrganization.Name).To(Equal(createdOrganization.Name))
			})

			It("Should not create an organization with repeated name", func() {
				name := uuid.NewV4().String()

				_, createdOrganization1Err := models.CreateOrganization(db, name)
				Expect(createdOrganization1Err).To(BeNil())

				_, createdOrganization2Err := models.CreateOrganization(db, name)
				Expect(createdOrganization2Err).NotTo(BeNil())
			})
		})
	})
	Describe("Update Organization", func() {
		It("Should update an organization for an existent id", func() {
			name := uuid.NewV4().String()
			organization, createdOrganizationErr := models.CreateOrganization(db, name)
			Expect(createdOrganizationErr).To(BeNil())

			newName := uuid.NewV4().String()
			updatedOrganization, updatedOrganizationErr := models.UpdateOrganization(db, organization.ID, newName)
			Expect(updatedOrganizationErr).To(BeNil())
			dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, organization.ID)
			Expect(dbOrganizationErr).To(BeNil())

			Expect(updatedOrganization.Name).To(Equal(dbOrganization.Name))
			Expect(updatedOrganization.Name).To(Equal(newName))
		})

		It("Should not update an organization with repeated name", func() {
			name1 := uuid.NewV4().String()
			_, createdOrganization1Err := models.CreateOrganization(db, name1)
			Expect(createdOrganization1Err).To(BeNil())

			name2 := uuid.NewV4().String()
			organization2, createdOrganization2Err := models.CreateOrganization(db, name2)
			Expect(createdOrganization2Err).To(BeNil())

			_, updatedOrganizationErr := models.UpdateOrganization(db, organization2.ID, name1)
			Expect(updatedOrganizationErr).NotTo(BeNil())

			dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, organization2.ID)
			Expect(dbOrganizationErr).To(BeNil())
			Expect(dbOrganization.Name).To(Equal(name2))
		})

		It("Should not update an organization for an unexistent id", func() {
			name := uuid.NewV4().String()
			organization, createdOrganizationErr := models.CreateOrganization(db, name)
			Expect(createdOrganizationErr).To(BeNil())

			newName := uuid.NewV4().String()
			invalidID := uuid.NewV4().String()
			_, updatedOrganizationErr := models.UpdateOrganization(db, invalidID, newName)
			Expect(updatedOrganizationErr).NotTo(BeNil())

			dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, organization.ID)
			Expect(dbOrganizationErr).To(BeNil())
			Expect(dbOrganization.Name).To(Equal(name))
		})
	})

	Describe("Get Organization", func() {
		It("Should retrieve an organization for an existent id", func() {
			name := uuid.NewV4().String()
			organization, createdOrganizationErr := models.CreateOrganization(db, name)
			Expect(createdOrganizationErr).To(BeNil())

			dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, organization.ID)
			Expect(dbOrganizationErr).To(BeNil())
			Expect(dbOrganization.Name).To(Equal(organization.Name))
		})

		It("Should not retrieve an organization for an unexistent id", func() {
			invalidID := uuid.NewV4().String()
			_, dbOrganizationErr := models.GetOrganizationByID(db, invalidID)
			Expect(dbOrganizationErr).NotTo(BeNil())
		})

		It("Should retrieve an organization for an existent name", func() {
			name := uuid.NewV4().String()
			organization, createdOrganizationErr := models.CreateOrganization(db, name)
			Expect(createdOrganizationErr).To(BeNil())

			dbOrganization, dbOrganizationErr := models.GetOrganizationByName(db, organization.Name)
			Expect(dbOrganizationErr).To(BeNil())
			Expect(dbOrganization.Name).To(Equal(organization.Name))
		})

		It("Should not retrieve an organization for an unexistent name", func() {
			invalidName := uuid.NewV4().String()
			_, dbOrganizationErr := models.GetOrganizationByName(db, invalidName)
			Expect(dbOrganizationErr).NotTo(BeNil())
		})
	})
})
