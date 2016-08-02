package workers

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// BatchPGWorker contains all batch pg worker configs and channels
type BatchPGWorker struct {
	PgToKafkaChan chan *messages.InputMessage
	DoneChan      chan struct{}
	Config        *viper.Viper
	Logger        zap.Logger
	Db            models.DB
	ConfigPath    string
}

type parseError struct {
	Message string
}

func (e parseError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

func (worker *BatchPGWorker) createChannels() {
	worker.PgToKafkaChan = make(chan *messages.InputMessage,
		worker.Config.GetInt("batch_pg_workers.modules.pgtokafkachansize"))
}

func (worker *BatchPGWorker) connectDatabase() {
	host := worker.Config.GetString("postgres.host")
	user := worker.Config.GetString("postgres.user")
	dbName := worker.Config.GetString("postgres.dbname")
	password := worker.Config.GetString("postgres.password")
	port := worker.Config.GetInt("postgres.port")
	sslMode := worker.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(host, user, port, sslMode, dbName, password)

	if err != nil {
		worker.Logger.Error(
			"Could not connect to postgres...",
			zap.String("host", host),
			zap.Int("port", port),
			zap.String("user", user),
			zap.String("dbName", dbName),
			zap.String("error", err.Error()),
		)
		panic(err)
	}
	worker.Db = db
}

// TODO: Set all default configs
func (worker *BatchPGWorker) setConfigurationDefaults() {
	worker.Config.SetDefault("healthcheck.workingtext", "working")
	worker.Config.SetDefault("postgres.host", "localhost")
	worker.Config.SetDefault("postgres.user", "khan")
	worker.Config.SetDefault("postgres.dbname", "khan")
	worker.Config.SetDefault("postgres.port", 5432)
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("batch_pg_workers.modules.logger.level", "error")
	worker.Config.SetDefault("batch_pg_workers.modules.producers", 1)
	worker.Config.SetDefault("batch_pg_workers.modules.pgtokafkachansize", 10)
	worker.Config.SetDefault("batch_pg_workers.postgres.defaults.pushexpiry", 0)
	worker.Config.SetDefault("batch_pg_workers.postgres.defaults.modifiers.limit", 1000)
	worker.Config.SetDefault("batch_pg_workers.postgres.defaults.modifiers.order", "updated_at ASC")
}

func (worker *BatchPGWorker) loadConfiguration() {
	worker.Config.SetConfigFile(worker.ConfigPath)
	worker.Config.SetEnvPrefix("marathon")
	worker.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	worker.Config.AutomaticEnv()

	if err := worker.Config.ReadInConfig(); err == nil {
		worker.Logger.Info("Loaded config file.", zap.String("configFile", worker.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", worker.ConfigPath))
	}
}

// Configure configures the worker
func (worker *BatchPGWorker) Configure() {
	worker.setConfigurationDefaults()
	worker.Logger = ConfigureLogger(Log{Level: "warn"}, worker.Config)
	worker.loadConfiguration()
	worker.Logger = ConfigureLogger(Log{}, worker.Config)
	worker.createChannels()
	worker.connectDatabase()
}

// StartWorker starts the workers according to the configuration and returns the workers object
func (worker *BatchPGWorker) StartWorker(message string, filters [][]interface{}, modifiers [][]interface{}) {
	// Run modules
	for i := 0; i < worker.Config.GetInt("workers.modules.producers"); i++ {
		go worker.producer(worker.Config, "batch_pg_workers", worker.PgToKafkaChan, worker.DoneChan)
	}
	go worker.pgReader(message, filters, modifiers, worker.PgToKafkaChan)
}

// Close stops the modules of the instance
func (worker BatchPGWorker) Close() {
	worker.Logger.Warn("Stopping workers")
	// close(worker.PgToKafkaChan)
	// close(worker.DoneChan)
	worker.Logger.Warn("Stopped workers")
}

// GetBatchPGWorker returns a new worker
func GetBatchPGWorker(configPath string) *BatchPGWorker {
	worker := &BatchPGWorker{
		Config:     viper.New(),
		ConfigPath: configPath,
	}
	worker.Configure()
	return worker
}

func buildTopicName(app string, service string) string {
	return fmt.Sprintf("%s_%s", app, service)
}

// PGReader reads from pg in batches and sends the built messages to kafka
func (worker *BatchPGWorker) pgReader(message string, filters [][]interface{},
	modifiers [][]interface{}, outChan chan<- *messages.InputMessage) {

	msgObj, err := worker.parse(message)
	if err != nil {
		worker.Logger.Fatal(
			"Could not parse message",
			zap.String("message", message),
			zap.Error(err),
		)
	}

	limit := -1
	for _, modifier := range modifiers {
		if modifier[0] == "LIMIT" {
			limit = modifier[1].(int)
		}
	}
	if limit <= 0 {
		worker.Logger.Fatal("Limit should be greater than 0", zap.Int("limit", limit))
	}

	userTokensCount, err := models.CountUserTokensByFilters(worker.Db, msgObj.App, msgObj.Service, filters, modifiers)
	if err != nil {
		worker.Logger.Fatal(
			"Error while counting tokens",
			zap.String("app", msgObj.App),
			zap.String("service", msgObj.Service),
			zap.String("filters", fmt.Sprintf("%+v", filters)),
			zap.String("modifiers", fmt.Sprintf("%+v", modifiers)),
			zap.Error(err),
		)
	}

	if userTokensCount < int64(0) {
		worker.Logger.Fatal(
			"Tokens size lower than 0",
			zap.String("app", msgObj.App),
			zap.String("service", msgObj.Service),
			zap.String("filters", fmt.Sprintf("%+v", filters)),
			zap.String("modifiers", fmt.Sprintf("%+v", modifiers)),
			zap.Error(err),
		)
	}

	userTokens, err := models.GetUserTokensBatchByFilters(worker.Db, msgObj.App, msgObj.Service, filters, modifiers)
	if err != nil {
		worker.Logger.Fatal(
			"Error while getting users",
			zap.String("app", msgObj.App),
			zap.String("service", msgObj.Service),
			zap.String("filters", fmt.Sprintf("%+v", filters)),
			zap.String("modifiers", fmt.Sprintf("%+v", modifiers)),
			zap.Error(err),
		)
	}

	processesTokens := int64(0)
	for processesTokens < userTokensCount {
		for _, userToken := range userTokens {
			msgObj.Token = userToken.Token
			msgObj.Locale = userToken.Locale
			outChan <- msgObj
		}
		processesTokens += int64(len(userTokens))
	}
}

// Parse parses the message string received, it expects to be in JSON format,
// as defined by the RequestMessage struct
func (worker *BatchPGWorker) parse(msg string) (*messages.InputMessage, error) {
	msgObj := messages.NewInputMessage()
	err := json.Unmarshal([]byte(msg), msgObj)
	if err != nil {
		errStr := fmt.Sprintf("Error parsing JSON: %+v, the message was: %s", err, msg)
		e := parseError{errStr}
		return msgObj, e
	}

	e := parseError{}
	// All fields should be set, except by either Template & Params or Message
	if msgObj.App == "" || msgObj.Service == "" {
		errStr := fmt.Sprintf(
			"One of the mandatory fields is missing app=%s, service=%s", msgObj.App, msgObj.Service)
		worker.Logger.Error(
			errStr,
			zap.String("app", msgObj.App),
			zap.String("service", msgObj.Service),
		)
		e = parseError{errStr}
	}
	// Either Template & Params should be defined or Message should be defined
	// Not both at the same time
	if msgObj.Template != "" && msgObj.Params == nil {
		errStr := "Template defined, but not Params"
		worker.Logger.Error(
			errStr,
			zap.String("template", msgObj.Template),
			zap.String("params", fmt.Sprintf("%+v", msgObj.Params)),
		)
		e = parseError{errStr}
	}
	if msgObj.Template != "" && (msgObj.Message != nil && len(msgObj.Message) > 0) {
		errStr := "Both Template and Message defined"
		worker.Logger.Error(
			errStr,
			zap.String("template", msgObj.Template),
			zap.String("message", fmt.Sprintf("%+v", msgObj.Message)),
		)
		e = parseError{errStr}
	}
	if msgObj.Template == "" && (msgObj.Message == nil || len(msgObj.Message) == 0) {
		errStr := "Either Template or Message should be defined"
		worker.Logger.Error(
			errStr,
			zap.String("template", msgObj.Template),
			zap.String("message", fmt.Sprintf("%+v", msgObj.Message)),
		)
		e = parseError{errStr}
	}
	if msgObj.PushExpiry < 0 {
		errStr := "PushExpiry should be above 0"
		worker.Logger.Error(errStr, zap.Int64("pushexpiry", msgObj.PushExpiry))
		e = parseError{errStr}
	}
	if e.Message != "" {
		return nil, e
	}

	worker.Logger.Debug("Decoded message", zap.String("msg", msg))
	return msgObj, nil
}

func (worker *BatchPGWorker) producer(config *viper.Viper, configRoot string, inChan <-chan *messages.InputMessage, doneChan <-chan struct{}) {
	saramaConfig := sarama.NewConfig()
	topic := config.GetStringSlice("workers.consumer.topics")[0]
	brokers := config.GetStringSlice("workers.consumer.brokers")
	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		worker.Logger.Error(
			"Failed to start kafka producer",
			zap.Error(err),
		)
		return
	}
	defer producer.Close()

	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		case msg := <-inChan:
			byteMsg, err := json.Marshal(msg)
			if err != nil {
				worker.Logger.Error("Error marshalling message", zap.Error(err))
			}
			saramaMessage := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(byteMsg),
			}

			_, _, err = producer.SendMessage(saramaMessage)
			if err != nil {
				worker.Logger.Error(
					"Error sending message",
					zap.Error(err),
				)
			} else {
				worker.Logger.Info(
					"Sent message",
					zap.String("topic", topic),
				)
			}
		}
	}
}
