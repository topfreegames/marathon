package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var _ = Describe("Marathon API Handler", func() {
	var (
		l      zap.Logger
		config *viper.Viper
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

		config = viper.New()
		config.SetConfigFile("./../config/test.yaml")
		config.SetEnvPrefix("marathon")
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()
		err = config.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Notification Handler", func() {
		It("Should create a Notification", func() {
			a := GetDefaultTestApp(config)
			appName := "notifierApp1"
			service := "gcm"
			group := randomdata.FirstName(randomdata.RandomGender)
			payload1 := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			status, body := PostJSON(a, "/apps", payload1)

			Expect(status).To(Equal(http.StatusOK), body)
			var result1 map[string]interface{}
			json.Unmarshal([]byte(body), &result1)
			notifierID := result1["notifierID"]

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
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			err = a.Db.Insert(template)
			Expect(err).NotTo(HaveOccurred())

			payload2 := map[string]interface{}{
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
			url1 := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			status, body = PostJSON(a, url1, payload2)

			Expect(status).To(Equal(http.StatusOK), body)
			var result2 map[string]interface{}
			json.Unmarshal([]byte(body), &result2)
			Expect(result2["success"]).To(BeTrue())
		})

		It("Should not create a Notification when broken json", func() {
			a := GetDefaultTestApp(config)
			appName := "notifierApp2"
			service := "gcm"
			group := randomdata.FirstName(randomdata.RandomGender)
			payload1 := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			status, body := PostJSON(a, "/apps", payload1)

			Expect(status).To(Equal(http.StatusOK), body)
			var result1 map[string]interface{}
			json.Unmarshal([]byte(body), &result1)
			notifierID := result1["notifierID"]

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
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			err = a.Db.Insert(template)
			Expect(err).NotTo(HaveOccurred())

			payload2 := map[string]interface{}{
				"filters": map[string]interface{}{"user_id": userID},
				"message": map[string]interface{}{
					"template": "test_template",
					"params":   map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
					"message":  map[string]interface{}{},
					"metadata": map[string]interface{}{"meta": "data"},
				},
			}
			payloadStr, err := json.Marshal(payload2)
			wrongPayloadString := payloadStr[:len(payloadStr)-1]
			Expect(err).NotTo(HaveOccurred())
			url := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			status, body = PostJSON(a, url, wrongPayloadString)
			Expect(status).To(Equal(http.StatusBadRequest), body)
		})

		It("Should create a Notification and get status", func() {
			a := GetDefaultTestApp(config)
			appName := "notifierApp3"
			service := "gcm"
			group := randomdata.FirstName(randomdata.RandomGender)
			payload1 := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}

			status, body := PostJSON(a, "/apps", payload1)

			Expect(status).To(Equal(http.StatusOK), body)
			var result1 map[string]interface{}
			json.Unmarshal([]byte(body), &result1)
			notifierID := result1["notifierID"]

			createdTable, err := models.CreateUserTokensTable(a.Db, appName, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTable.TableName).To(Equal(models.GetTableName(appName, service)))

			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+03:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err = models.UpsertToken(
				a.Db, appName, service, userID1, token1, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			userID2 := uuid.NewV4().String()
			token2 := uuid.NewV4().String()

			_, err = models.UpsertToken(
				a.Db, appName, service, userID2, token2, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			template := &models.Template{
				Name:     "test_template",
				Locale:   "en",
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			err = a.Db.Insert(template)
			Expect(err).NotTo(HaveOccurred())

			// ===========================================================================================
			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			var consumerConfig = viper.New()
			consumerConfig.SetConfigFile("./../config/test.yaml")
			consumerConfig.SetEnvPrefix("marathon")
			consumerConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			consumerConfig.AutomaticEnv()
			err = consumerConfig.ReadInConfig()
			Expect(err).NotTo(HaveOccurred())
			go consumer.Consumer(l, consumerConfig, appName, service, outChan, doneChan)
			time.Sleep(500 * time.Millisecond)
			// ===========================================================================================

			payload2 := map[string]interface{}{
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
			url1 := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			status, body = PostJSON(a, url1, payload2)

			Expect(status).To(Equal(http.StatusOK), body)
			var result2 map[string]interface{}
			json.Unmarshal([]byte(body), &result2)
			Expect(result2["success"]).To(BeTrue())

			time.Sleep(500 * time.Millisecond)

			url2 := fmt.Sprintf("/notifiers/%s/notifications/%s", notifierID, result2["id"])
			status, body = Get(a, url2)
			Expect(status).To(Equal(http.StatusOK), body)
			var result3 map[string]interface{}
			json.Unmarshal([]byte(body), &result3)
			Expect(result3["success"]).To(BeTrue())

			var resStatus map[string]interface{}
			json.Unmarshal([]byte(result3["status"].(string)), &resStatus)

			var kafkaStatus map[string]interface{}
			json.Unmarshal([]byte(resStatus["kafkaStatus"].(string)), &kafkaStatus)
			Expect(kafkaStatus["initialKafkaOffset"]).To(Equal(float64(0)))
			// Expect(kafkaStatus["currentKafkaOffset"]).To(Equal(float64(2)))    TODO: uncomment this

			var workerStatus map[string]interface{}
			json.Unmarshal([]byte(resStatus["workerStatus"].(string)), &workerStatus)
			Expect(workerStatus["totalPages"]).To(Equal(float64(1)))
			Expect(workerStatus["processedPages"]).To(Equal(float64(1)))
			Expect(workerStatus["totalTokens"]).To(Equal(float64(2)))
			Expect(workerStatus["totalProcessedTokens"]).To(Equal(float64(2)))
		})

		It("Should create a Notifications and get list", func() {
			a := GetDefaultTestApp(config)
			appName := "notifierApp4"
			service := "gcm"
			group := randomdata.FirstName(randomdata.RandomGender)
			payload1 := map[string]interface{}{
				"appName":        appName,
				"service":        service,
				"organizationID": uuid.NewV4().String(),
				"appGroup":       group,
			}
			status, body := PostJSON(a, "/apps", payload1)

			Expect(status).To(Equal(http.StatusOK), body)
			var result1 map[string]interface{}
			json.Unmarshal([]byte(body), &result1)
			notifierID := result1["notifierID"]

			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			locale := uuid.NewV4().String()[:2]
			region := uuid.NewV4().String()[:2]
			tz := "GMT+03:00"
			buildN := uuid.NewV4().String()
			optOut := []string{uuid.NewV4().String(), uuid.NewV4().String()}

			_, err := models.UpsertToken(
				a.Db, appName, service, userID1, token1, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			userID2 := uuid.NewV4().String()
			token2 := uuid.NewV4().String()

			_, err = models.UpsertToken(
				a.Db, appName, service, userID2, token2, locale, region, tz, buildN, optOut,
			)
			Expect(err).NotTo(HaveOccurred())

			template := &models.Template{
				Name:     "test_template",
				Locale:   "en",
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			err = a.Db.Insert(template)
			Expect(err).NotTo(HaveOccurred())

			payload2 := map[string]interface{}{
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
			url1 := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			status, body = PostJSON(a, url1, payload2)

			Expect(status).To(Equal(http.StatusOK), body)
			var result2 map[string]interface{}
			json.Unmarshal([]byte(body), &result2)
			Expect(result2["success"]).To(BeTrue())

			time.Sleep(500 * time.Millisecond)

			url2 := fmt.Sprintf("/notifiers/%s/notifications", notifierID)
			status, body = Get(a, url2)
			Expect(status).To(Equal(http.StatusOK), body)
			var result3 map[string]interface{}
			json.Unmarshal([]byte(body), &result3)
			Expect(result3["success"]).To(BeTrue())
		})
	})
})
