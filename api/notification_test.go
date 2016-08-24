package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

var _ = Describe("Marathon API Handler", func() {
	var (
		l zap.Logger
	)
	BeforeEach(func() {
		type Table struct {
			TableName string `db:"tablename"`
		}
		l = mt.NewMockLogger()
		_db, err := models.GetTestDB(l)
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
				"filters": map[string]interface{}{
					"user_id": userID,
					"locale":  locale,
					"region":  region,
					"tz":      tz,
					"build_n": buildN,
					"scope":   "testScope",
				},
				"message": map[string]interface{}{
					"template": "test_template",
					"params":   map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
					"message":  map[string]interface{}{},
					"metadata": map[string]interface{}{"meta": "data"},
				},
			}
			res := PostJSON(a, "/apps/notifications", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("Should not create a Notification when broken json", func() {
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
			payloadStr, err := json.Marshal(payload)
			wrongPayloadString := payloadStr[:len(payloadStr)-1]
			Expect(err).NotTo(HaveOccurred())
			req := SendRequest(a, "POST", "/apps/notifications")
			res := req.WithBytes([]byte(wrongPayloadString)).Expect()

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("Should create a Notification and get status", func() {
			app := "app_test_3_1"
			service := "gcm"
			a := GetDefaultTestApp()

			createdTable, err := models.CreateUserTokensTable(a.Db, app, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTable.TableName).To(Equal(models.GetTableName(app, service)))

			userID := uuid.NewV4().String()
			token := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+03:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err = models.UpsertToken(
				a.Db, app, service, userID, token, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			userID = uuid.NewV4().String()
			token = uuid.NewV4().String()

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
				"filters": map[string]interface{}{
					"locale":  locale,
					"region":  region,
					"tz":      tz,
					"build_n": buildN,
					"scope":   "testScope",
				},
				"message": map[string]interface{}{
					"template": "test_template",
					"params":   map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
					"message":  map[string]interface{}{},
					"metadata": map[string]interface{}{"meta": "data"},
				},
			}
			res := PostJSON(a, "/apps/notifications", payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			time.Sleep(500 * time.Millisecond)

			req := SendRequest(a, "GET", fmt.Sprintf("/apps/notifications/%s", result["id"]))
			res = req.Expect()
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			var status map[string]interface{}
			json.Unmarshal([]byte(result["status"].(string)), &status)
			Expect(status["totalPages"]).To(Equal(float64(1)))
			Expect(status["processedPages"]).To(Equal(float64(1)))
			Expect(status["totalTokens"]).To(Equal(float64(2)))
			Expect(status["processedTokens"]).To(Equal(float64(2)))
		})
	})
})
