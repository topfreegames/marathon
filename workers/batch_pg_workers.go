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
func (worker *BatchPGWorker) StartWorker(message *messages.InputMessage, filters [][]interface{}, modifiers [][]interface{}) {
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
func GetBatchPGWorker(worker *BatchPGWorker) (*BatchPGWorker, error) {
	if worker.ConfigPath == "" {
		errStr := "Invalid worker config"
		worker.Logger.Error(errStr, zap.Object("worker", worker))
		e := parseError{errStr}
		return nil, e
	}
	if worker.Config == nil {
		worker.Config = viper.New()
	}
	worker.Configure()
	return worker, nil
}

func buildTopicName(app string, service string) string {
	return fmt.Sprintf("%s_%s", app, service)
}

// PGReader reads from pg in batches and sends the built messages to kafka
func (worker *BatchPGWorker) pgReader(message *messages.InputMessage, filters [][]interface{},
	modifiers [][]interface{}, outChan chan<- *messages.InputMessage) (chan<- *messages.InputMessage, error) {

	limit := -1
	for _, modifier := range modifiers {
		if modifier[0] == "LIMIT" {
			limit = modifier[1].(int)
		}
	}
	if limit <= 0 {
		worker.Logger.Fatal("Limit should be greater than 0", zap.Int("limit", limit))
	}

	userTokensCount, err := models.CountUserTokensByFilters(worker.Db, message.App, message.Service, filters, modifiers)
	if err != nil {
		worker.Logger.Fatal(
			"Error while counting tokens",
			zap.String("app", message.App),
			zap.String("service", message.Service),
			zap.Object("filters", filters),
			zap.Object("modifiers", modifiers),
			zap.Error(err),
		)
	}

	if userTokensCount < int64(0) {
		worker.Logger.Fatal(
			"Tokens size lower than 0",
			zap.String("app", message.App),
			zap.String("service", message.Service),
			zap.Object("filters", filters),
			zap.Object("modifiers", modifiers),
			zap.Error(err),
		)
	}

	pages := userTokensCount / int64(limit)
	workerModifiers := [2][]interface{}{ // TODO: Should we order ? {"ORDER BY", "updated_at ASC"}
		{"LIMIT", limit},
	}
	for page := int64(0); page < pages; page++ {
		workerModifiers[1] = []interface{}{"OFFSET", page}
		userTokens, err := models.GetUserTokensBatchByFilters(
			worker.Db, message.App, message.Service, filters, modifiers,
		)
		if err != nil {
			worker.Logger.Error(
				"Failed to get user tokens by filters and modifiers",
				zap.Object("filters", filters),
				zap.Object("modifiers", modifiers),
				zap.Error(err),
			)
			return outChan, err
		}
		processedTokens := 0
		for processedTokens < len(userTokens) {
			for _, userToken := range userTokens {
				message.Token = userToken.Token
				message.Locale = userToken.Locale
				outChan <- message
			}
			processedTokens += len(userTokens)
		}
	}
	return outChan, nil
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
