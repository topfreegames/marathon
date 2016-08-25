package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/Pallinder/go-randomdata"
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
			notifierID := result["notifierID"]

			userID := uuid.NewV4().String()
			token := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+04:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err := models.UpsertToken(
				a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
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

			payload = map[string]interface{}{
				"app":      appName,
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
			url := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			res = PostJSON(a, url, payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("Should not create a Notification when broken json", func() {
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
			notifierID := result["notifierID"]

			createdTable, err := models.CreateUserTokensTable(a.Db, appName, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTable.TableName).To(Equal(models.GetTableName(appName, service)))

			userID := uuid.NewV4().String()
			token := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+04:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err = models.UpsertToken(
				a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
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

			payload = map[string]interface{}{
				"app":      appName,
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
			url := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			req := SendRequest(a, "POST", url)
			res = req.WithBytes([]byte(wrongPayloadString)).Expect()

			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("Should create a Notification and get status", func() {
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
			notifierID := result["notifierID"]

			createdTable, err := models.CreateUserTokensTable(a.Db, appName, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTable.TableName).To(Equal(models.GetTableName(appName, service)))

			userID := uuid.NewV4().String()
			token := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+03:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err = models.UpsertToken(
				a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			userID = uuid.NewV4().String()
			token = uuid.NewV4().String()

			_, err = models.UpsertToken(
				a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
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

			payload = map[string]interface{}{
				"app":      appName,
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
			url := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			res = PostJSON(a, url, payload)

			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			time.Sleep(500 * time.Millisecond)

			url = fmt.Sprintf("/notifiers/%s/notifications/%s", notifierID, result["id"])
			req := SendRequest(a, "GET", url)
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

		// It("Should create a Notifications and get list", func() {
		// 	a := GetDefaultTestApp()
		// 	appName := randomdata.FirstName(randomdata.RandomGender)
		// 	service := randomdata.FirstName(randomdata.RandomGender)[:3]
		// 	group := randomdata.FirstName(randomdata.RandomGender)
		// 	payload := map[string]interface{}{
		// 		"appName":        appName,
		// 		"service":        service,
		// 		"organizationID": uuid.NewV4().String(),
		// 		"appGroup":       group,
		// 	}
		//
		// 	res := PostJSON(a, "/apps", payload)
		//
		// 	Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
		// 	var result map[string]interface{}
		// 	json.Unmarshal([]byte(res.Body().Raw()), &result)
		// 	notifierID := result["notifierID"]
		//
		// 	createdTable, err := models.CreateUserTokensTable(a.Db, appName, service)
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(createdTable.TableName).To(Equal(models.GetTableName(appName, service)))
		//
		// 	userID := uuid.NewV4().String()
		// 	token := uuid.NewV4().String()
		// 	locale := uuid.NewV4().String()[:2]
		// 	region := uuid.NewV4().String()[:2]
		// 	tz := "GMT+03:00"
		// 	buildN := uuid.NewV4().String()
		// 	optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}
		//
		// 	_, err = models.UpsertToken(
		// 		a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
		// 	)
		// 	Expect(err).NotTo(HaveOccurred())
		//
		// 	userID = uuid.NewV4().String()
		// 	token = uuid.NewV4().String()
		//
		// 	_, err = models.UpsertToken(
		// 		a.Db, appName, service, userID, token, locale, region, tz, buildN, optOut,
		// 	)
		// 	Expect(err).NotTo(HaveOccurred())
		//
		// 	template := &models.Template{
		// 		Name:     "test_template",
		// 		Locale:   "en",
		// 		Service:  service,
		// 		Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
		// 		Body:     map[string]interface{}{"alert": "{{param1}}, {{param2}}, {{param3}}"},
		// 	}
		// 	err = a.Db.Insert(template)
		// 	Expect(err).NotTo(HaveOccurred())
		//
		// 	payload = map[string]interface{}{
		// 		"app":      appName,
		// 		"service":  service,
		// 		"pageSize": 10,
		// 		"filters": map[string]interface{}{
		// 			"locale":  locale,
		// 			"region":  region,
		// 			"tz":      tz,
		// 			"build_n": buildN,
		// 			"scope":   "testScope",
		// 		},
		// 		"message": map[string]interface{}{
		// 			"template": "test_template",
		// 			"params":   map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
		// 			"message":  map[string]interface{}{},
		// 			"metadata": map[string]interface{}{"meta": "data"},
		// 		},
		// 	}
		// 	url := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
		// 	res = PostJSON(a, url, payload)
		//
		// 	Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
		// 	json.Unmarshal([]byte(res.Body().Raw()), &result)
		// 	Expect(result["success"]).To(BeTrue())
		//
		// 	time.Sleep(500 * time.Millisecond)
		//
		// 	url = fmt.Sprintf("/notifiers/%s/notifications/%s", notifierID, result["id"])
		// 	req := SendRequest(a, "GET", url)
		// 	res = req.Expect()
		// 	Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
		// 	json.Unmarshal([]byte(res.Body().Raw()), &result)
		// 	Expect(result["success"]).To(BeTrue())
		//
		// 	var status map[string]interface{}
		// 	json.Unmarshal([]byte(result["status"].(string)), &status)
		// 	Expect(status["totalPages"]).To(Equal(float64(1)))
		// 	Expect(status["processedPages"]).To(Equal(float64(1)))
		// 	Expect(status["totalTokens"]).To(Equal(float64(2)))
		// 	Expect(status["processedTokens"]).To(Equal(float64(2)))
		// })
	})
})
