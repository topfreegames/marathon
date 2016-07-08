package models_test

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Describe("Basic Operations", func() {
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
				name := randomdata.SillyName()

				createdOrganization, createdOrganizationErr := models.CreateOrganization(db, name)
				Expect(createdOrganizationErr).To(BeNil())

				dbOrganization, dbOrganizationErr := models.GetOrganizationByID(db, createdOrganization.ID)
				Expect(dbOrganizationErr).To(BeNil())
				Expect(dbOrganization.Name).To(Equal(createdOrganization.Name))
			})
		})
	})
})
