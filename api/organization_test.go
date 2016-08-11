package api_test

import (
	"encoding/json"
	"net/http"

	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Clan API Handler", func() {
	BeforeEach(func() {})

	Describe("Create Clan Handler", func() {
		It("Should create clan", func() {
			payload := map[string]interface{}{
				"name": randomdata.FirstName(randomdata.RandomGender),
			}

			a := GetDefaultTestApp()
			res := PostJSON(a, "/organizations", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["id"]).NotTo(BeNil())
			Expect(result["name"]).To(Equal(payload["name"]))

			organizationIdString := str(result["id"])
			organizationId, err := uuid.FromString(organizationIdString)
			Expect(err).NotTo(HaveOccurred())
			dbOrganization, err := models.GetOrganizationByID(a.Db, organizationId)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbOrganization.ID).To(Equal(organizationId))
		})

		It("Should not create clan if missing name", func() {
			payload := map[string]interface{}{}

			a := GetDefaultTestApp()

			countBefore, err := models.CountOrganizations(a.Db)
			Expect(err).NotTo(HaveOccurred())

			res := PostJSON(a, "/organizations", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())

			countAfter, err := models.CountOrganizations(a.Db)
			Expect(err).NotTo(HaveOccurred())
			Expect(countAfter - countBefore).To(BeZero())
		})

		It("Should not create clan with repeated name", func() {
			a := GetDefaultTestApp()
			name := randomdata.FirstName(randomdata.RandomGender)

			_, createdOrganization1Err := models.CreateOrganization(a.Db, name)
			Expect(createdOrganization1Err).To(BeNil())

			payload := map[string]interface{}{"name": name}

			countBefore, err := models.CountOrganizations(a.Db)
			Expect(err).NotTo(HaveOccurred())

			res := PostJSON(a, "/organizations", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())

			countAfter, err := models.CountOrganizations(a.Db)
			Expect(err).NotTo(HaveOccurred())
			Expect(countAfter - countBefore).To(BeZero())
		})
	})
})
