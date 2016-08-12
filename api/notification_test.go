package api_test

import (
	"encoding/json"
	"net/http"

	"git.topfreegames.com/topfreegames/marathon/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Marathon API Handler", func() {
	BeforeEach(func() {})

	Describe("Create Notification Handler", func() {
		It("Should create a Notification", func() {
			a := GetDefaultTestApp()

			app := "app_test_3_1"
			service := "apns"

			createdTable, err := models.CreateUserTokensTable(a.Db, app, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

			userID := uuid.NewV4().String()
			token := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+04:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err = models.UpsertToken(
				a.Db, app, service, userID, token, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"app":      "app",
				"service":  "service",
				"pageSize": 10,
				"filters":  map[string]interface{}{"user_id": userID},
			}
			res := PostJSON(a, "/apps/appName/users/notifications", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})
	})
})
