package workers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/minio/minio-go"
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
		db               *models.DB
		l                zap.Logger
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
		modifiers        [][]interface{}
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

		templateName = "integrationtesttemplatename"
		appName = "integrationtestappname"
		service = "apns"
		locale = "PT"
		region = "BR"
		tz = "GMT+03:00"
		optOut = []string{"optout1", "optout2"}
		buildN = "919191"
		templateDefaults = map[string]interface{}{"username": "banduk"}
		templateBody = map[string]interface{}{"alert": "%{username} sent you a message."}
		_, err = models.CreateTemplate(db, templateName, locale, templateDefaults, templateBody)
		Expect(err).NotTo(HaveOccurred())

		_, err = models.CreateUserTokensTable(db, appName, service)
		Expect(err).NotTo(HaveOccurred())

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

		modifiers = [][]interface{}{
			{"ORDER BY", "updated_at ASC"},
			{"LIMIT", 1},
		}
	})

	Describe("Batch csv workers", func() {
		XIt("Send messages to users from csv", func() {
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

			config := viper.New()
			config.SetConfigFile("./../config/test.yaml")
			config.AutomaticEnv()
			config.ReadInConfig()

			s3AccessKeyID := config.GetString("s3.accessKey")
			s3SecretAccessKey := config.GetString("s3.secretAccessKey")
			ssl := true
			s3Client, err := minio.New("s3.amazonaws.com", s3AccessKeyID, s3SecretAccessKey, ssl)
			Expect(err).NotTo(HaveOccurred())
			sampleCsv := fmt.Sprintf("%s\n%s\n", userID1, userID2)
			testKey := "test/files/testcsv.csv"
			_, err = s3Client.PutObject("tfg-push-notifications", testKey, strings.NewReader(sampleCsv), "application/octet-stream")

			// Batch worker that reads from pg and send to kafka
			worker := &BatchCsvWorker{
				ConfigPath: "./../config/test.yaml",
				Message:    message,
				Modifiers:  modifiers,
				Bucket:     "tfg-push-notifications",
				Key:        testKey,
				Notifier:   notifier,
			}
			batchWorker, err := GetBatchCsvWorker(worker)
			Expect(err).NotTo(HaveOccurred())

			// Consume message produced by our pipeline
			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			batchWorkerConfig := *worker.Config
			config.Set("workers.consumer.consumergroup", "consumer-group-test-1")
			config.Set("workers.consumer.topicTemplate", batchWorkerConfig.GetString("workers.producer.topicTemplate"))

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

			tokens := []string{token1, token2}
			for i := 0; i < 2; i++ {
				processedMessage := <-outChan
				processedMessageObj := messages.NewApnsMessage()
				json.Unmarshal([]byte(processedMessage), &processedMessageObj)
				Expect(tokens).To(ContainElement(processedMessageObj.DeviceToken))
				Expect(processedMessageObj.PushExpiry).To(Equal(int64(0)))
			}
			// FIXME: How to test the message?
		})
	})
})
