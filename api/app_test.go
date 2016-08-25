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
			appName := randomdata.FirstName(randomdata.RandomGender)
			service := randomdata.FirstName(randomdata.RandomGender)[:3]
			group := randomdata.FirstName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			res := PostJSON(a, "/apps", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["id"]).NotTo(BeNil())
			Expect(result["appName"]).To(Equal(appName))
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

			appName1 := randomdata.FirstName(randomdata.RandomGender)
			service1 := randomdata.FirstName(randomdata.RandomGender)[:3]
			group1 := randomdata.FirstName(randomdata.RandomGender)
			organizationID1 := uuid.NewV4()

			_, err := models.CreateApp(a.Db, appName1, organizationID1, group1)
			Expect(err).NotTo(HaveOccurred())

			appName2 := appName1
			service2 := service1
			group2 := randomdata.FirstName(randomdata.RandomGender)
			organizationID2 := uuid.NewV4()

			payload := map[string]interface{}{
				"appName":        appName2,
				"service":        service2,
				"organizationID": organizationID2,
				"appGroup":       group2,
			}

			countBefore, err := models.CountApps(a.Db)
			HaveOccurred()
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

		It("Should not create app with broken json", func() {
			a := GetDefaultTestApp()
			appName := randomdata.FirstName(randomdata.RandomGender)
			service := randomdata.FirstName(randomdata.RandomGender)[:3]
			group := randomdata.FirstName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			payloadStr, err := json.Marshal(payload)
			wrongPayloadString := payloadStr[:len(payloadStr)-1]
			Expect(err).NotTo(HaveOccurred())
			req := SendRequest(a, "POST", "/apps")
			res := req.WithBytes([]byte(wrongPayloadString)).Expect()

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("Get Apps Handler", func() {
		It("Should get apps with notifiers", func() {
			a := GetDefaultTestApp()
			appName := randomdata.FirstName(randomdata.RandomGender)
			service := randomdata.FirstName(randomdata.RandomGender)[:3]
			group := randomdata.FirstName(randomdata.RandomGender)
			payload := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			res := PostJSON(a, "/apps", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["id"]).NotTo(BeNil())
			Expect(result["appName"]).To(Equal(appName))
			Expect(result["appGroup"]).To(Equal(payload["appGroup"]))
			Expect(result["organizationID"]).To(Equal(payload["organizationID"]))

			appIdString := str(result["id"])
			appId, err := uuid.FromString(appIdString)
			Expect(err).NotTo(HaveOccurred())
			dbApp, err := models.GetAppByID(a.Db, appId)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbApp.ID).To(Equal(appId))

			req := SendRequest(a, "GET", "/apps")
			res = req.Expect()
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			// var status map[string]interface{}
			// json.Unmarshal([]byte(result["status"].(string)), &status)
			// Expect(status["totalPages"]).To(Equal(float64(1)))
			// Expect(status["processedPages"]).To(Equal(float64(1)))
			// Expect(status["totalTokens"]).To(Equal(float64(2)))
			// Expect(status["processedTokens"]).To(Equal(float64(2)))
		})
	})

})
