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

		It("Should create notifiers for different services and same app", func() {
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

			service2 := randomdata.FirstName(randomdata.RandomGender)[:3]
			payload2 := map[string]interface{}{
				"appName":        appName,
				"service":        service2,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}
			res = PostJSON(a, "/apps", payload2)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["id"]).NotTo(BeNil())
			Expect(result["appName"]).To(Equal(appName))
			Expect(result["appGroup"]).To(Equal(payload["appGroup"]))
			Expect(result["organizationID"]).To(Equal(payload["organizationID"]))

			appIdString = str(result["id"])
			appId, err = uuid.FromString(appIdString)
			Expect(err).NotTo(HaveOccurred())
			dbApp, err = models.GetAppByID(a.Db, appId)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbApp.ID).To(Equal(appId))
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

	// TODO: Improve tests => Check notifier creation
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

			apps := result["apps"].([]interface{})
			for i := range apps {
				app := apps[i].(map[string]interface{})
				Expect(app["id"]).NotTo(BeNil())
				Expect(app["notifierID"]).NotTo(BeNil())
				Expect(app["service"]).NotTo(BeNil())
				Expect(app["appGroup"]).NotTo(BeNil())
			}
		})
	})

})
