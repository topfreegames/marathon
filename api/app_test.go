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

var _ = Describe("Marathon API Handler", func() {
	BeforeEach(func() {})

	Describe("Create App Handler", func() {
		It("Should create app", func() {
			a := GetDefaultTestApp()
			name := randomdata.FirstName(randomdata.RandomGender)
			group := randomdata.FirstName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"name":           name,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			res := PostJSON(a, "/apps", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["id"]).NotTo(BeNil())
			Expect(result["name"]).To(Equal(payload["name"]))
			Expect(result["appGroup"]).To(Equal(payload["appGroup"]))
			Expect(result["organizationID"]).To(Equal(payload["organizationID"]))

			appIdString := str(result["id"])
			appId, err := uuid.FromString(appIdString)
			Expect(err).NotTo(HaveOccurred())
			dbApp, err := models.GetAppByID(a.Db, appId)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbApp.ID).To(Equal(appId))
		})
		It("Should not create apps with repeated names", func() {
			a := GetDefaultTestApp()

			name1 := randomdata.FirstName(randomdata.RandomGender)
			appGroup1 := randomdata.FirstName(randomdata.RandomGender)
			organizationID1 := uuid.NewV4()

			_, err := models.CreateApp(a.Db, name1, organizationID1, appGroup1)
			Expect(err).To(BeNil())

			name2 := name1
			appGroup2 := randomdata.FirstName(randomdata.RandomGender)
			organizationID2 := uuid.NewV4().String()

			payload := map[string]interface{}{
				"name":           name2,
				"organizationID": organizationID2,
				"appGroup":       appGroup2,
			}

			countBefore, err := models.CountApps(a.Db)
			Expect(err).NotTo(HaveOccurred())

			res := PostJSON(a, "/apps", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())

			countAfter, err := models.CountApps(a.Db)
			Expect(err).NotTo(HaveOccurred())
			Expect(countAfter - countBefore).To(BeZero())
		})
	})
})
