package workers_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"git.topfreegames.com/topfreegames/marathon/util"
	"git.topfreegames.com/topfreegames/marathon/workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var _ = Describe("Models", func() {
	type Table struct {
		TableName string `db:"tablename"`
	}
	var (
		db *models.DB
		l  zap.Logger
	)

	BeforeEach(func() {
		l = mt.NewMockLogger()
		_db, err := models.GetTestDB(l)
		Expect(err).NotTo(HaveOccurred())
		Expect(_db).NotTo(BeNil())
		db = _db

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

	Describe("Batch pg workers", func() {
		XIt("Send messages for segmented of users", func() {
			appName := "integrationtestappname1"
			templateName := "integrationtesttemplatename1"
			service := "gcm"
			locale := "PT"
			region := "BR"
			tz := "GMT+03:00"
			optOut := []string{"optout1", "optout2"}
			buildN := "919191"
			templateDefaults := map[string]interface{}{"username": "banduk"}
			templateBody := map[string]interface{}{"alert": "{{username}} sent you a message."}
			_, err := models.CreateTemplate(db, templateName, service, locale, templateDefaults, templateBody)
			Expect(err).NotTo(HaveOccurred())

			_, err = models.CreateUserTokensTable(db, appName, service)
			Expect(err).NotTo(HaveOccurred())

			pushExpiry := int64(0)
			msg := map[string]interface{}{"aps": "hey"}
			metadata := map[string]interface{}{"meta": "data"}

			message := &messages.InputMessage{
				App:        appName,
				Service:    service,
				PushExpiry: pushExpiry,
				Locale:     locale,
				Message:    msg,
				Metadata:   metadata,
			}

			filters := [][]interface{}{
				{"locale", locale},
			}
			modifiers := [][]interface{}{
				{"ORDER BY", "updated_at ASC"},
				{"LIMIT", 1},
			}
			appGroup := uuid.NewV4().String()
			organizationID := uuid.NewV4()

			app, createdAppErr := models.CreateApp(db, appName, organizationID, appGroup)
			Expect(createdAppErr).To(BeNil())
			appID := app.ID

			notifier, createdNotifier1Err := models.CreateNotifier(db, appID, service)
			Expect(createdNotifier1Err).To(BeNil())

			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID1, token1, locale, region, tz, buildN, optOut)
			Expect(err).NotTo(HaveOccurred())

			// Batch worker that reads from pg and send to kafka
			worker := &workers.BatchPGWorker{
				ConfigPath: "./../config/test.yaml",
				Message:    message,
				Filters:    filters,
				Modifiers:  modifiers,
				Notifier:   notifier,
				App:        app,
			}
			batchWorker, err := workers.GetBatchPGWorker(worker)
			Expect(err).NotTo(HaveOccurred())

			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			var config = viper.New()
			config.SetConfigFile("./../config/test.yaml")
			config.SetEnvPrefix("marathon")
			config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			config.AutomaticEnv()
			err = config.ReadInConfig()
			Expect(err).NotTo(HaveOccurred())

			config.Set("workers.consumer.topicTemplate", config.GetString("workers.producer.topicTemplate"))

			go consumer.Consumer(l, config, appName, service, outChan, doneChan)

			batchWorker.Start()
			Expect(batchWorker).NotTo(BeNil())

			timeElapsed := time.Duration(0)
			timeStepMillis := 500 * time.Millisecond
			for len(outChan) < 1 {
				time.Sleep(timeStepMillis)
				if timeElapsed > 5000*time.Millisecond {
					Fail("Timeout while waiting out channel from consumer")
				}
				timeElapsed += timeStepMillis
			}
			Expect(len(outChan)).To(Equal(2))

			processedMessage := <-outChan
			processedMessage = <-outChan
			processedMessageObj := messages.NewApnsMessage()
			json.Unmarshal([]byte(processedMessage), &processedMessageObj)
			Expect(processedMessageObj.DeviceToken).To(Equal(token1))
			Expect(processedMessageObj.PushExpiry).To(Equal(int64(0)))
			// FIXME: How to test the message?
		})

		XIt("Send messages for segmented of users", func() {
			appName := "integrationtestappname2"
			templateName := "integrationtesttemplatename2"
			service := "gcm"
			locale := "PT"
			region := "BR"
			tz := "GMT+03:00"
			optOut := []string{"optout1", "optout2"}
			buildN := "919191"
			templateDefaults := map[string]interface{}{"username": "banduk"}
			templateBody := map[string]interface{}{"alert": "{{username}} sent you a message."}
			_, err := models.CreateTemplate(db, templateName, service, locale, templateDefaults, templateBody)
			Expect(err).NotTo(HaveOccurred())

			_, err = models.CreateUserTokensTable(db, appName, service)
			Expect(err).NotTo(HaveOccurred())

			pushExpiry := int64(0)
			msg := map[string]interface{}{"aps": "hey"}
			metadata := map[string]interface{}{"meta": "data"}

			message := &messages.InputMessage{
				App:        appName,
				Service:    service,
				PushExpiry: pushExpiry,
				Locale:     locale,
				Message:    msg,
				Metadata:   metadata,
			}

			filters := [][]interface{}{
				{"locale", locale},
			}
			modifiers := [][]interface{}{
				{"ORDER BY", "updated_at ASC"},
				{"LIMIT", 1},
			}
			appGroup := uuid.NewV4().String()
			organizationID := uuid.NewV4()

			app, createdAppErr := models.CreateApp(db, appName, organizationID, appGroup)
			Expect(createdAppErr).To(BeNil())

			appID := app.ID

			notifier, createdNotifier1Err := models.CreateNotifier(db, appID, service)
			Expect(createdNotifier1Err).To(BeNil())

			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID1, token1, locale, region, tz, buildN, optOut)
			Expect(err).NotTo(HaveOccurred())

			userID2 := uuid.NewV4().String()
			token2 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID2, token2, locale, region, tz, buildN, optOut)
			Expect(err).NotTo(HaveOccurred())

			userID3 := uuid.NewV4().String()
			token3 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID3, token3, locale, region, tz, buildN, optOut)
			Expect(err).NotTo(HaveOccurred())

			tokens := []string{token1, token2, token3}

			// Batch worker that reads from pg and sent to continuous worker
			worker := &workers.BatchPGWorker{
				ConfigPath: "./../config/test.yaml",
				Message:    message,
				Filters:    filters,
				Modifiers:  modifiers,
				Notifier:   notifier,
				App:        app,
			}
			batchWorker, err := workers.GetBatchPGWorker(worker)
			Expect(err).NotTo(HaveOccurred())

			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			var config = viper.New()
			config.SetConfigFile("./../config/test.yaml")
			config.SetEnvPrefix("marathon")
			config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			config.AutomaticEnv()
			err = config.ReadInConfig()
			Expect(err).NotTo(HaveOccurred())

			config.Set("workers.consumer.topicTemplate", config.GetString("workers.producer.topicTemplate"))

			go consumer.Consumer(l, config, appName, service, outChan, doneChan)

			batchWorker.Start()
			Expect(batchWorker).NotTo(BeNil())

			timeElapsed := time.Duration(0)
			timeStepMillis := 500 * time.Millisecond
			for len(outChan) < 3 {
				time.Sleep(timeStepMillis)
				if timeElapsed > 5000*time.Millisecond {
					Fail("Timeout while waiting out channel from consumer")
				}
				timeElapsed += timeStepMillis
			}
			Expect(len(outChan)).To(Equal(4))

			processedMessage1 := <-outChan // Discard first message (from other test)
			processedMessage1 = <-outChan
			processedMessage2 := <-outChan
			processedMessage3 := <-outChan

			processedMessageObj1 := messages.NewApnsMessage()
			json.Unmarshal([]byte(processedMessage1), &processedMessageObj1)

			processedMessageObj2 := messages.NewApnsMessage()
			json.Unmarshal([]byte(processedMessage2), &processedMessageObj2)

			processedMessageObj3 := messages.NewApnsMessage()
			json.Unmarshal([]byte(processedMessage3), &processedMessageObj3)

			processedMessageObjs := []*messages.ApnsMessage{
				processedMessageObj1,
				processedMessageObj2,
				processedMessageObj3,
			}

			for _, processedMessageObj := range processedMessageObjs {
				Expect(util.SliceContains(tokens, processedMessageObj.DeviceToken)).To(BeTrue())
				Expect(processedMessageObj.PushExpiry).To(Equal(int64(0)))
				// FIXME: How to test the message?
			}
		})
	})
})
