package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Marathon API Handler", func() {
	BeforeEach(func() {
		type Table struct {
			TableName string `db:"tablename"`
		}
		_db, err := models.GetTestDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(_db).NotTo(BeNil())

		// Truncate all tables
		var tables []Table
		_, _ = _db.Select(&tables, "SELECT tablename from pg_tables where schemaname='public'")
		var tableNames []string
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName)
		}
		if len(tableNames) > 0 {
			_, err = _db.Exec(fmt.Sprintf("TRUNCATE %s", strings.Join(tableNames, ",")))
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Create Notification Handler", func() {
		It("Should create a Notification", func() {
			a := GetDefaultTestApp()

			app := "app_test_3_1"
			service := "gcm"

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

			template := &models.Template{
				Name:     "test_template",
				Locale:   "en",
				Service:  service,
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "{{param1}}, {{param2}}, {{param3}}"},
			}
			err = a.Db.Insert(template)
			Expect(err).NotTo(HaveOccurred())

			payload := map[string]interface{}{
				"app":      app,
				"service":  service,
				"pageSize": 10,
				"filters":  map[string]interface{}{"user_id": userID},
				"message": map[string]interface{}{
					"template": "test_template",
					"params":   map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
					"message":  map[string]interface{}{},
					"metadata": map[string]interface{}{"meta": "data"},
				},
			}
			res := PostJSON(a, "/apps/appName/users/notifications", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})
	})
})
