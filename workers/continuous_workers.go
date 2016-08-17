package workers

import (
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/kafka/producer"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/templates"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// ContinuousWorker contains all continuous worker configs and channels
type ContinuousWorker struct {
	App                    string
	Service                string
	KafkaDoneChan          chan struct{}
	FetcherDoneChan        chan struct{}
	InputKafkaChan         chan string
	KafkaToParserChan      chan string
	ParserToFetcherChan    chan *messages.InputMessage
	FetcherToBuilderChan   chan *messages.TemplatedMessage
	BuilderToKafkaChan     chan *messages.KafkaMessage
	BuilderToKafkaDoneChan chan struct{}
	ParserDoneChan         chan struct{}
	BuilderDoneChan        chan struct{}
	Config                 *viper.Viper
	Logger                 zap.Logger
	Db                     *models.DB
	ConfigPath             string
}

func (worker *ContinuousWorker) createChannels() {
	worker.KafkaDoneChan = make(chan struct{},
		worker.Config.GetInt("workers.modules.kafkadonechansize"))
	worker.FetcherDoneChan = make(chan struct{},
		worker.Config.GetInt("workers.modules.fetcherdonechansize"))
	worker.InputKafkaChan = make(chan string,
		worker.Config.GetInt("workers.modules.inputkafkachansize"))
	worker.KafkaToParserChan = make(chan string,
		worker.Config.GetInt("workers.modules.kafkatoparserchansize"))
	worker.ParserToFetcherChan = make(chan *messages.InputMessage,
		worker.Config.GetInt("workers.modules.parsertofetcherchansize"))
	worker.FetcherToBuilderChan = make(chan *messages.TemplatedMessage,
		worker.Config.GetInt("workers.modules.fetchertobuilderchansize"))
	worker.BuilderToKafkaChan = make(chan *messages.KafkaMessage,
		worker.Config.GetInt("workers.modules.buildertokafkachansize"))
	worker.BuilderToKafkaDoneChan = make(chan struct{},
		worker.Config.GetInt("workers.modules.buildertokafkadonechansize"))
	worker.ParserDoneChan = make(chan struct{},
		worker.Config.GetInt("workers.modules.parserdonechansize"))
	worker.BuilderDoneChan = make(chan struct{},
		worker.Config.GetInt("workers.modules.builderdonechansize"))
}

func (worker *ContinuousWorker) connectDatabase() {
	host := worker.Config.GetString("postgres.host")
	user := worker.Config.GetString("postgres.user")
	dbName := worker.Config.GetString("postgres.dbname")
	password := worker.Config.GetString("postgres.password")
	port := worker.Config.GetInt("postgres.port")
	sslMode := worker.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(worker.Logger, host, user, port, sslMode, dbName, password)

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

func (worker *ContinuousWorker) setConfigurationDefaults() {
	worker.Config.SetDefault("healthcheck.workingtext", "working")
	worker.Config.SetDefault("postgres.host", "localhost")
	worker.Config.SetDefault("postgres.user", "marathon")
	worker.Config.SetDefault("postgres.dbname", "marathon")
	worker.Config.SetDefault("postgres.port", 5432)
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("workers.consumer.consumergroup", "test-consumer-group")
	worker.Config.SetDefault("workers.consumer.topics", []string{"input-topic"})
	worker.Config.SetDefault("workers.consumer.brokers", []string{"localhost:9092"})
	worker.Config.SetDefault("workers.modules.logger.level", "info")
	worker.Config.SetDefault("workers.modules.consumers", 1)
	worker.Config.SetDefault("workers.modules.parsers", 1)
	worker.Config.SetDefault("workers.modules.fetchers", 1)
	worker.Config.SetDefault("workers.modules.builders", 1)
	worker.Config.SetDefault("workers.modules.producers", 1)
	worker.Config.SetDefault("workers.modules.kafkadonechansize", 1)
	worker.Config.SetDefault("workers.modules.fecherdonechansize", 1)
	worker.Config.SetDefault("workers.modules.inputkafkachansize", 10)
	worker.Config.SetDefault("workers.modules.inputchansize", 10)
	worker.Config.SetDefault("workers.modules.kafkatoparserchansize", 10)
	worker.Config.SetDefault("workers.modules.parsertofetcherchansize", 10)
	worker.Config.SetDefault("workers.modules.fetchertobuilderchansize", 10)
	worker.Config.SetDefault("workers.modules.buildertokafkachansize", 10)
}

func (worker *ContinuousWorker) loadConfiguration() {
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
func (worker *ContinuousWorker) Configure() {
	worker.setConfigurationDefaults()
	worker.Logger = ConfigureLogger(Log{Level: "warn"}, worker.Config)
	worker.loadConfiguration()
	worker.Logger = ConfigureLogger(Log{}, worker.Config)
	worker.createChannels()
	worker.connectDatabase()
}

// StartWorker starts the workers according to the configuration and returns the workers object
func (worker *ContinuousWorker) StartWorker() {
	// Run modules
	for i := 0; i < worker.Config.GetInt("workers.modules.consumers"); i++ {
		go consumer.Consumer(worker.Logger, worker.Config, "workers", worker.App, worker.Service,
			worker.InputKafkaChan, worker.KafkaDoneChan)
	}
	for i := 0; i < worker.Config.GetInt("workers.modules.parsers"); i++ {
		go templates.Parser(worker.Logger, true, worker.KafkaToParserChan, worker.ParserToFetcherChan, worker.ParserDoneChan)
	}
	for i := 0; i < worker.Config.GetInt("workers.modules.fetchers"); i++ {
		go templates.Fetcher(worker.Logger, worker.ParserToFetcherChan, worker.FetcherToBuilderChan, worker.FetcherDoneChan, worker.Db)
	}
	for i := 0; i < worker.Config.GetInt("workers.modules.builders"); i++ {
		go templates.Builder(worker.Logger, worker.Config, "workers", worker.FetcherToBuilderChan, worker.BuilderToKafkaChan, worker.BuilderDoneChan)
	}
	for i := 0; i < worker.Config.GetInt("workers.modules.producers"); i++ {
		go producer.Producer(worker.Logger, worker.Config, "workers", worker.BuilderToKafkaChan, worker.BuilderToKafkaDoneChan)
	}
}

// Close stops the modules of the instance
func (worker ContinuousWorker) Close() {
	worker.Logger.Warn("Stopping workers")
	// close(worker.InputKafkaChan)
	// close(worker.KafkaDoneChan)
	// close(worker.KafkaToParserChan)
	// close(worker.ParserToFetcherChan)
	// close(worker.FetcherToBuilderChan)
	// close(worker.FetcherDoneChan)
	// close(worker.BuilderToKafkaChan)
	// close(worker.BuilderToKafkaDoneChan)
	worker.Logger.Warn("Stopped workers")
}

// GetContinuousWorker returns a new worker
func GetContinuousWorker(worker *ContinuousWorker) (*ContinuousWorker, error) {
	if worker.ConfigPath == "" {
		errStr := "Invalid worker config - ConfigPath is required"
		worker.Logger.Error(errStr, zap.Object("worker", worker))
		e := parseError{errStr}
		return nil, e
	}
	if worker.App == "" {
		errStr := "Invalid worker config - App is required"
		worker.Logger.Error(errStr, zap.Object("worker", worker))
		e := parseError{errStr}
		return nil, e
	}
	if worker.Service == "" {
		errStr := "Invalid worker config - Service is required"
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
