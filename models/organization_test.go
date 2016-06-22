package models

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
)

func TestOrganizationModel(t *testing.T) {
	g := Goblin(t)
	db, err := GetTestDB()
	g.Assert(err).Equal(nil)

	g.Describe("Organization Model", func() {
		g.Describe("Basic Operations", func() {
			g.It("Should create an organization through a factory", func() {
				organization, organizationErr := CreateOrganizationFactory(db, map[string]interface{}{})
				g.Assert(organizationErr).Equal(nil)
				insertOrganizationErr := db.Insert(organization)
				g.Assert(insertOrganizationErr).Equal(nil)

				dbOrganization, dbOrganizationErr := GetOrganizationByID(db, organization.ID)
				g.Assert(dbOrganizationErr).Equal(nil)
				g.Assert(dbOrganization.Name).Equal(organization.Name)
			})

			g.It("Should create an organization", func() {
				name := randomdata.SillyName()

				createdOrganization, createdOrganizationErr := CreateOrganization(db, name)
				g.Assert(createdOrganizationErr).Equal(nil)

				dbOrganization, dbOrganizationErr := GetOrganizationByID(db, createdOrganization.ID)
				g.Assert(dbOrganizationErr).Equal(nil)
				g.Assert(dbOrganization.Name).Equal(createdOrganization.Name)
			})
		})
	})
}
