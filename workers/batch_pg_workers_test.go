package workers_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/workers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var _ = Describe("Models", func() {
	type Table struct {
		TableName string `db:"tablename"`
	}
	var (
		db               models.DB
		err              error
		appName          string
		service          string
		templateName     string
		locale           string
		templateDefaults map[string]interface{}
		templateBody     map[string]interface{}
		region           string
		tz               string
		optOut           []string
		buildN           string
		pushExpiry       int64
		msg              map[string]interface{}
		metadata         map[string]interface{}
		message          *messages.InputMessage
		byteMessage      []byte
		filters          [][]interface{}
		modifiers        [][]interface{}
	)

	BeforeEach(func() {
		_db, err := models.GetTestDB()
		Expect(err).To(BeNil())
		Expect(_db).NotTo(BeNil())
		db = _db

		// Truncate all tables
		var tables []Table
		_, _ = db.Select(&tables, "SELECT tablename from pg_tables where schemaname='public'")
		var tableNames []string
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName)
		}
		if len(tableNames) > 0 {
			_, err := db.Exec(fmt.Sprintf("TRUNCATE %s", strings.Join(tableNames, ",")))
			Expect(err).To(BeNil())
		}

		templateName = "integrationtesttemplatename"
		appName = "integrationtestappname"
		service = "apns"
		locale = "PT"
		region = "BR"
		tz = "GMT+03:00"
		optOut = []string{"optout1", "optout2"}
		buildN = "919191"
		templateDefaults = map[string]interface{}{"username": "banduk"}
		templateBody = map[string]interface{}{"alert": "{{username}} sent you a message."}
		_, err = models.CreateTemplate(db, templateName, service, locale, templateDefaults, templateBody)
		Expect(err).To(BeNil())

		_, err = models.CreateUserTokensTable(db, appName, service)
		Expect(err).To(BeNil())

		pushExpiry = int64(0)
		msg = map[string]interface{}{"aps": "hey"}
		metadata = map[string]interface{}{"meta": "data"}

		message = &messages.InputMessage{
			App:        appName,
			Service:    service,
			PushExpiry: pushExpiry,
			Locale:     locale,
			Message:    msg,
			Metadata:   metadata,
		}
		byteMessage, err = json.Marshal(message)
		Expect(err).To(BeNil())

		filters = [][]interface{}{
			{"locale", locale},
		}
		modifiers = [][]interface{}{
			{"ORDER BY", "updated_at ASC"},
			{"LIMIT", 1},
		}
	})

	Describe("Batch pg workers", func() {
		It("Send messages for segmented of users", func() {
			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID1, token1, locale, region, tz, buildN, optOut)
			Expect(err).To(BeNil())

			// Continuous worker that process messages produced
			continuousWorker := workers.GetContinuousWorker("./../config/test.yaml")
			// Batch worker that reads from pg and sent to continuous worker
			batchWorker := workers.GetBatchPGWorker("./../config/test.yaml")
			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			consumerConfig := *continuousWorker.Config
			var config = viper.New()
			config.SetDefault("workers.consumer.brokers", consumerConfig.GetStringSlice("workers.consumer.brokers"))
			config.SetDefault("workers.consumer.consumergroup", "consumer-group-test-1")
			config.SetDefault("workers.consumer.topics", consumerConfig.GetStringSlice("workers.consumer.topics"))

			go consumer.Consumer(config, "workers", outChan, doneChan)

			continuousWorker.StartWorker()
			Expect(continuousWorker).NotTo(BeNil())

			batchWorker.StartWorker(string(byteMessage), filters, modifiers)
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
			Expect(len(outChan)).To(Equal(1))

			processedMessage := <-outChan
			processedMessageObj := messages.NewInputMessage()
			json.Unmarshal([]byte(processedMessage), &processedMessageObj)
			Expect(processedMessageObj.App).To(Equal(appName))
			Expect(processedMessageObj.Service).To(Equal(service))
			Expect(processedMessageObj.Token).To(Equal(token1))
			Expect(processedMessageObj.PushExpiry).To(Equal(pushExpiry))
			Expect(processedMessageObj.Locale).To(Equal(locale))
			Expect(processedMessageObj.Message).To(Equal(msg))
			Expect(processedMessageObj.Metadata).To(Equal(metadata))
		})

		It("Send messages for segmented of users", func() {
			userID1 := uuid.NewV4().String()
			token1 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID1, token1, locale, region, tz, buildN, optOut)
			Expect(err).To(BeNil())

			userID2 := uuid.NewV4().String()
			token2 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID2, token2, locale, region, tz, buildN, optOut)
			Expect(err).To(BeNil())

			userID3 := uuid.NewV4().String()
			token3 := uuid.NewV4().String()
			_, err = models.UpsertToken(db, appName, service, userID3, token3, locale, region, tz, buildN, optOut)
			Expect(err).To(BeNil())

			tokens := []string{token1, token2, token3}

			// Continuous worker that process messages produced
			continuousWorker := workers.GetContinuousWorker("./../config/test.yaml")
			// Batch worker that reads from pg and sent to continuous worker
			batchWorker := workers.GetBatchPGWorker("./../config/test.yaml")
			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			consumerConfig := *continuousWorker.Config
			var config = viper.New()
			config.SetDefault("workers.consumer.brokers", consumerConfig.GetStringSlice("workers.consumer.brokers"))
			config.SetDefault("workers.consumer.consumergroup", "consumer-group-test-2")
			config.SetDefault("workers.consumer.topics", consumerConfig.GetStringSlice("workers.consumer.topics"))
			go consumer.Consumer(config, "workers", outChan, doneChan)

			continuousWorker.StartWorker()
			Expect(continuousWorker).NotTo(BeNil())

			batchWorker.StartWorker(string(byteMessage), filters, modifiers)
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

			processedMessage1 := <-outChan
			processedMessage1 = <-outChan // Discard first message (from other test)
			processedMessage2 := <-outChan
			processedMessage3 := <-outChan

			processedMessageObj1 := messages.NewInputMessage()
			json.Unmarshal([]byte(processedMessage1), &processedMessageObj1)

			processedMessageObj2 := messages.NewInputMessage()
			json.Unmarshal([]byte(processedMessage2), &processedMessageObj2)

			processedMessageObj3 := messages.NewInputMessage()
			json.Unmarshal([]byte(processedMessage3), &processedMessageObj3)

			processedMessageObjs := []*messages.InputMessage{
				processedMessageObj1,
				processedMessageObj2,
				processedMessageObj3,
			}

			for _, processedMessageObj := range processedMessageObjs {
				Expect(processedMessageObj.App).To(Equal(appName))
				Expect(processedMessageObj.Service).To(Equal(service))
				Expect(processedMessageObj.PushExpiry).To(Equal(pushExpiry))
				Expect(processedMessageObj.Locale).To(Equal(locale))
				Expect(processedMessageObj.Message).To(Equal(msg))
				Expect(processedMessageObj.Metadata).To(Equal(metadata))
				Expect(contains(tokens, processedMessageObj.Token)).To(BeTrue())
			}
		})
	})
})
