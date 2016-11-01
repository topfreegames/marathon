package workers

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/extensions"
	"git.topfreegames.com/topfreegames/marathon/kafka/producer"
	"git.topfreegames.com/topfreegames/marathon/log"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/templates"
	"git.topfreegames.com/topfreegames/marathon/util"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// BatchPGWorker contains all batch pg worker configs and channels
type BatchPGWorker struct {
	ID                          uuid.UUID
	Config                      *viper.Viper
	zkClient                    *extensions.ZkClient
	Logger                      zap.Logger
	Notifier                    *models.Notifier
	App                         *models.App
	Message                     *messages.InputMessage
	Filters                     [][]interface{}
	Modifiers                   [][]interface{}
	StartedAt                   int64
	FinishedAt                  int64
	KafkaClient                 *cluster.Client
	KafkaTopic                  string
	InitialKafkaOffset          int64
	CurrentKafkaOffset          int64
	Db                          *models.DB
	ConfigPath                  string
	RedisClient                 *util.RedisClient
	TotalTokens                 int64
	TotalProcessedTokens        int
	TotalPages                  int64
	ProcessedPages              int64
	BatchPGWorkerStatusDoneChan chan struct{}
	BatchPGWorkerDoneChan       chan struct{}
	PgToParserChan              chan string
	ParserDoneChan              chan struct{}
	ParserToFetcherChan         chan *messages.InputMessage
	FetcherDoneChan             chan struct{}
	FetcherToBuilderChan        chan *messages.TemplatedMessage
	BuilderDoneChan             chan struct{}
	BuilderToProducerChan       chan *messages.KafkaMessage
	ProducerDoneChan            chan struct{}
}

type parseError struct {
	Message string
}

func (e parseError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

func (worker *BatchPGWorker) createChannels() {
	worker.BatchPGWorkerStatusDoneChan = make(chan struct{}, worker.Config.GetInt("batchpgworkerstatusdonechan"))
	worker.BatchPGWorkerDoneChan = make(chan struct{}, worker.Config.GetInt("batchpgworkerdonechansize"))
	worker.PgToParserChan = make(chan string, worker.Config.GetInt("pgtoparserchansize"))
	worker.ParserDoneChan = make(chan struct{}, worker.Config.GetInt("parserdonechansize"))
	worker.ParserToFetcherChan = make(chan *messages.InputMessage, worker.Config.GetInt("parsertofetcherchansize"))
	worker.FetcherDoneChan = make(chan struct{}, worker.Config.GetInt("fetcherdonechansize"))
	worker.FetcherToBuilderChan = make(chan *messages.TemplatedMessage, worker.Config.GetInt("fetchertobuilderchansize"))
	worker.BuilderDoneChan = make(chan struct{}, worker.Config.GetInt("builderdonechansize"))
	worker.BuilderToProducerChan = make(chan *messages.KafkaMessage, worker.Config.GetInt("buildertoproducerchansize"))
	worker.ProducerDoneChan = make(chan struct{}, worker.Config.GetInt("producerdonechansize"))
}

func (worker *BatchPGWorker) connectDatabase() error {
	host := worker.Config.GetString("postgres.host")
	user := worker.Config.GetString("postgres.user")
	dbName := worker.Config.GetString("postgres.dbname")
	password := worker.Config.GetString("postgres.password")
	port := worker.Config.GetInt("postgres.port")
	sslMode := worker.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(worker.Logger, host, user, port, sslMode, dbName, password)

	if err != nil {
		log.E(worker.Logger, "Could not connect to postgres...", func(cm log.CM) {
			cm.Write(
				zap.String("host", host),
				zap.Int("port", port),
				zap.String("user", user),
				zap.String("dbName", dbName),
				zap.String("error", err.Error()),
			)
		})
		return err
	}
	worker.Db = db
	return nil
}

func (worker *BatchPGWorker) configureZookeeperClient() {
	worker.zkClient = extensions.GetZkClient(worker.ConfigPath)
}

func (worker *BatchPGWorker) connectRedis() error {
	redisHost := worker.Config.GetString("redis.host")
	redisPort := worker.Config.GetInt("redis.port")
	redisPass := worker.Config.GetString("redis.password")
	redisDB := worker.Config.GetInt("redis.db")
	redisMaxPoolSize := worker.Config.GetInt("redis.maxPoolSize")

	rl := worker.Logger.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
		zap.Int("db", redisDB),
		zap.Int("maxPoolSize", redisMaxPoolSize),
	)
	log.D(rl, "Connecting to redis...")
	cli, err := util.GetRedisClient(redisHost, redisPort, redisPass, redisDB, redisMaxPoolSize, rl)
	if err != nil {
		log.E(rl, "Could not connect to redis...", func(cm log.CM) {
			cm.Write(zap.String("error", err.Error()))
		})
		return err
	}
	worker.RedisClient = cli
	log.I(rl, "Connected to redis successfully.")
	return nil
}

// TODO: Chec if all default configs are set
func (worker *BatchPGWorker) setConfigurationDefaults() {
	worker.Config.SetDefault("healthcheck.workingtext", "working")
	worker.Config.SetDefault("postgres.host", "localhost")
	worker.Config.SetDefault("postgres.user", "marathon")
	worker.Config.SetDefault("postgres.dbname", "marathon")
	worker.Config.SetDefault("postgres.port", 5432)
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("postgres.sslmode", "disable")
	worker.Config.SetDefault("workers.modules.logger.level", "error")
	worker.Config.SetDefault("workers.modules.producers", 1)
	worker.Config.SetDefault("workers.postgres.defaults.pushexpiry", 0)
	worker.Config.SetDefault("workers.postgres.defaults.modifiers.limit", 1000)
	worker.Config.SetDefault("workers.postgres.defaults.modifiers.order", "updated_at ASC")
	worker.Config.SetDefault("workers.modules.batchpgworkerdonechan", 1)
	worker.Config.SetDefault("workers.modules.pgtoparserchan", 10000)
	worker.Config.SetDefault("workers.modules.parserdonechan", 1)
	worker.Config.SetDefault("workers.modules.parsertofetcherchan", 1000)
	worker.Config.SetDefault("workers.modules.fetcherdonechan", 1)
	worker.Config.SetDefault("workers.modules.fetchertobuilderchan", 1000)
	worker.Config.SetDefault("workers.modules.builderdonechan", 1)
	worker.Config.SetDefault("workers.modules.buildertoproducerchan", 1000)
	worker.Config.SetDefault("workers.modules.producerdonechan", 1)
}

func (worker *BatchPGWorker) loadConfiguration() {
	worker.Config.SetConfigFile(worker.ConfigPath)
	worker.Config.SetConfigType("yaml")
	worker.Config.SetEnvPrefix("marathon")
	worker.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	worker.Config.AutomaticEnv()

	if err := worker.Config.ReadInConfig(); err == nil {
		log.I(worker.Logger, "Loaded config file.", func(cm log.CM) {
			cm.Write(zap.String("configFile", worker.Config.ConfigFileUsed()))
		})
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", worker.ConfigPath))
	}
}

func (worker *BatchPGWorker) configureKafkaClient() error {
	clusterConfig := cluster.NewConfig()
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Group.Return.Notifications = true
	clusterConfig.Version = sarama.V0_9_0_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	topicTemplate := worker.Config.GetString("workers.producer.topicTemplate")
	topic := fmt.Sprintf(topicTemplate, worker.App.Name, worker.Notifier.Service)
	worker.KafkaTopic = topic

	brokers, err := worker.zkClient.GetKafkaBrokers()

	l := worker.Logger.With(
		zap.String("clusterConfig", fmt.Sprintf("%+v", clusterConfig)),
		zap.String("topicTemplate", topicTemplate),
		zap.String("topic", topic),
		zap.Object("brokers", brokers),
	)

	if err != nil {
		log.E(l, "Could not get kafka brokers", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}

	client, err := cluster.NewClient(brokers, clusterConfig)
	if err != nil {
		log.E(l, "Could not create kafka client", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}
	log.D(l, "Created kafka client", func(cm log.CM) {
		cm.Write(zap.String("client", fmt.Sprintf("%+v", client)))
	})

	partitionID := int32(0)
	currentOffset := int64(-1)
	currentOffset, err = client.GetOffset(topic, partitionID, sarama.OffsetNewest)
	if err != nil {
		log.E(l, "Could not get kafka offset", func(cm log.CM) {
			cm.Write(
				zap.Int("partitionID", int(partitionID)),
				zap.String("error", err.Error()),
			)
		})
		return err
	}
	if currentOffset == int64(-1) {
		errStr := "Could not get offset from kafka"
		log.E(l, errStr, func(cm log.CM) {
			cm.Write(zap.Object("currentOffset", currentOffset))
		})
		err = parseError{errStr}
		return err
	}
	log.D(l, "Got kafka offset", func(cm log.CM) {
		cm.Write(zap.Int64("Offset", currentOffset))
	})

	worker.InitialKafkaOffset = currentOffset
	worker.KafkaClient = client
	return nil
}

// Configure configures the worker
func (worker *BatchPGWorker) Configure() error {
	worker.ID = uuid.NewV4()
	if worker.Config == nil {
		worker.Config = viper.New()
		worker.setConfigurationDefaults()
		worker.Logger = ConfigureLogger(Log{Level: "warn"}, worker.Config)
		worker.loadConfiguration()
	}
	worker.Logger = ConfigureLogger(Log{}, worker.Config)
	worker.createChannels()
	err := worker.connectDatabase()
	if err != nil {
		return err
	}

	err = worker.connectRedis()
	if err != nil {
		return err
	}

	worker.configureZookeeperClient()

	err = worker.configureKafkaClient()
	if err != nil {
		return err
	}

	return nil
}

// Start starts the workers according to the configuration and returns the workers object
func (worker *BatchPGWorker) Start() {
	requireToken := false

	worker.StartedAt = time.Now().Unix()
	worker.FinishedAt = 0

	log.I(worker.Logger, "Starting worker pipeline...")

	go worker.updateStatus(worker.BatchPGWorkerStatusDoneChan)

	// Run modules
	qtyParsers := worker.Config.GetInt("workers.modules.parsers")
	log.D(worker.Logger, "Starting parser...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyParsers))
	})
	for i := 0; i < qtyParsers; i++ {
		go templates.Parser(worker.Logger, requireToken, worker.PgToParserChan, worker.ParserToFetcherChan, worker.ParserDoneChan)
	}
	log.D(worker.Logger, "Started parser...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyParsers))
	})

	qtyFetchers := worker.Config.GetInt("workers.modules.fetchers")
	log.D(worker.Logger, "Starting fetcher...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyFetchers))
	})
	for i := 0; i < qtyFetchers; i++ {
		go templates.Fetcher(worker.Logger, worker.ParserToFetcherChan, worker.FetcherToBuilderChan, worker.FetcherDoneChan, worker.Db)
	}
	log.D(worker.Logger, "Started fetcher...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyFetchers))
	})

	qtyBuilders := worker.Config.GetInt("workers.modules.builders")
	log.D(worker.Logger, "Starting builder...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyBuilders))
	})
	for i := 0; i < qtyBuilders; i++ {
		go templates.Builder(worker.Logger, worker.Config, worker.FetcherToBuilderChan, worker.BuilderToProducerChan, worker.BuilderDoneChan)
	}
	log.D(worker.Logger, "Started builder...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyBuilders))
	})

	qtyProducers := worker.Config.GetInt("workers.modules.producers")
	log.D(worker.Logger, "Starting producer...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyProducers))
	})
	for i := 0; i < qtyProducers; i++ {
		go producer.Producer(worker.Logger, worker.Config, worker.BuilderToProducerChan, worker.ProducerDoneChan)
	}
	log.D(worker.Logger, "Started producer...", func(cm log.CM) {
		cm.Write(zap.Int("quantity", qtyProducers))
	})

	log.D(worker.Logger, "Starting pgReader...")
	go worker.pgReader(worker.Message, worker.Filters, worker.Modifiers, worker.PgToParserChan)
	log.D(worker.Logger, "Started pgReader...")

	log.I(worker.Logger, "Started worker pipeline...")
}

// Close stops the modules of the instance
func (worker BatchPGWorker) Close() {
	log.I(worker.Logger, "Stopping workers")
	// close(worker.PgToKafkaChan)
	// close(worker.DoneChan)
	log.I(worker.Logger, "Stopped workers")
}

// GetBatchPGWorker returns a new worker
func GetBatchPGWorker(worker *BatchPGWorker) (*BatchPGWorker, error) {
	if worker.ConfigPath == "" && worker.Config == nil {
		errStr := "Invalid worker config. Either Config or ConfigPath should be set"
		log.E(worker.Logger, errStr, func(cm log.CM) {
			cm.Write(zap.Object("worker", worker))
		})
		err := parseError{errStr}
		return nil, err
	}
	err := worker.Configure()
	if err != nil {
		return nil, err
	}
	return worker, nil
}

// GetWorkerStatus returns a map[string]interface{} with the current worker status
func (worker BatchPGWorker) GetWorkerStatus() map[string]interface{} {
	return map[string]interface{}{
		"notificationID":       worker.ID,
		"startedAt":            worker.StartedAt,
		"totalTokens":          worker.TotalTokens,
		"totalProcessedTokens": worker.TotalProcessedTokens,
		"totalPages":           worker.TotalPages,
		"processedPages":       worker.ProcessedPages,
		"message":              worker.Message,
		"filters":              worker.Filters,
	}
}

// GetKafkaStatus returns a map[string]interface{} with the current kafka status
func (worker BatchPGWorker) GetKafkaStatus() map[string]interface{} {
	currentOffset, err := worker.KafkaClient.GetOffset(worker.KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		log.W(worker.Logger, "Could not get kafka offset", func(cm log.CM) {
			cm.Write(
				zap.String("topic", worker.KafkaTopic),
				zap.String("error", err.Error()),
			)
		})
	}
	worker.CurrentKafkaOffset = currentOffset
	log.D(worker.Logger, "Got kafka offset", func(cm log.CM) {
		cm.Write(
			zap.Int64("Offset", worker.CurrentKafkaOffset),
			zap.Int64("InitialOffset", worker.InitialKafkaOffset),
			zap.String("topic", worker.KafkaTopic),
		)
	})
	return map[string]interface{}{
		"kafkaTopic":         worker.KafkaTopic,
		"initialKafkaOffset": worker.InitialKafkaOffset,
		"currentKafkaOffset": worker.CurrentKafkaOffset,
	}
}

// SetStatus sets the current notification status in redis
func (worker *BatchPGWorker) SetStatus() {
	cli := worker.RedisClient.Client
	redisKey := strings.Join([]string{worker.Notifier.ID.String(), worker.ID.String()}, "|")

	log.I(worker.Logger, "Set in redis", func(cm log.CM) {
		cm.Write(zap.String("key", redisKey))
	})

	workerStatus := worker.GetWorkerStatus()
	byteWorkerStatus, err := json.Marshal(workerStatus)
	if err != nil {
		log.P(worker.Logger, "Could not parse worker status", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
	workerStrStatus := string(byteWorkerStatus)

	kafkaStatus := worker.GetKafkaStatus()
	byteKafkaStatus, err := json.Marshal(kafkaStatus)
	if err != nil {
		log.P(worker.Logger, "Could not parse kafka status", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
	kafkaStrStatus := string(byteKafkaStatus)

	status := map[string]interface{}{
		"workerStatus": workerStrStatus,
		"kafkaStatus":  kafkaStrStatus,
	}
	byteStatus, err := json.Marshal(status)
	if err != nil {
		log.P(worker.Logger, "Could not parse status", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
	strStatus := string(byteStatus)

	// FIXME: What's the best TTL to set? 30 * time.Day ?
	if err = cli.Set(redisKey, strStatus, 0).Err(); err != nil {
		log.P(worker.Logger, "Failed to set notification key in redis", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
}

func (worker *BatchPGWorker) updateStatus(doneChan <-chan struct{}) {
	log.I(worker.Logger, "Starting status updater")
	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		default:
			log.D(worker.Logger, "Update worker status")
			worker.SetStatus()
			time.Sleep(250 * time.Millisecond)
			if worker.CurrentKafkaOffset-worker.InitialKafkaOffset >= worker.TotalTokens {
				// TODO: We should stop the worker here
				log.I(worker.Logger, "Finished sending tokens. Stopping status updater")
				worker.FinishedAt = time.Now().Unix()
				return
			}
		}
	}
}

// pgReader reads from pg in batches and sends the built messages to kafka
func (worker *BatchPGWorker) pgReader(message *messages.InputMessage, filters [][]interface{},
	modifiers [][]interface{}, outChan chan<- string) (chan<- string, error) {
	l := worker.Logger.With(
		zap.Object("message", message),
		zap.Object("filters", filters),
		zap.Object("modifiers", modifiers),
	)

	log.D(l, "Get limit")
	limit := -1
	for _, modifier := range modifiers {
		if modifier[0] == "LIMIT" {
			limit = modifier[1].(int)
		}
	}
	if limit <= 0 {
		log.F(worker.Logger, "Limit should be greater than 0", func(cm log.CM) {
			cm.Write(zap.Int("limit", limit))
		})
	}
	l = l.With(zap.Int("limit", limit))
	log.D(l, "Got limit")

	log.D(l, "Count userTokens")
	userTokensCount, err := models.CountUserTokensByFilters(
		worker.Db, message.App, message.Service, filters, modifiers,
	)
	if err != nil {
		log.F(l, "Error while counting tokens", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}

	// TODO: why can't we assign this directly from `models.CountUserTokensByFilters` ?
	worker.TotalTokens = userTokensCount

	if worker.TotalTokens < int64(0) {
		log.F(l, "worker.TotalTokens lower than 0", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	}
	l = l.With(zap.Int64("worker.TotalTokens", worker.TotalTokens))

	log.D(l, "Counted userTokens")

	log.D(l, "Get pages")
	pages := int64(math.Ceil(float64(userTokensCount) / float64(limit)))

	worker.TotalPages = pages
	l = l.With(zap.Int64("worker.TotalPages", worker.TotalPages))
	log.D(l, "Got pages")

	log.D(l, "Build workerModifiers")
	// TODO: Should we remove "ORDER BY" ?
	// TODO: We're rebuilding the structure
	workerModifiers := [][]interface{}{
		{"ORDER BY", "updated_at ASC"},
		{"LIMIT", limit},
		{"OFFSET", int64(0)},
	}
	l = l.With(zap.Object("workerModifiers", workerModifiers))
	log.D(l, "Built workerModifiers")

	worker.TotalProcessedTokens = 0

	for worker.ProcessedPages = int64(0); worker.ProcessedPages < worker.TotalPages; worker.ProcessedPages++ {
		log.D(l, "pgRead - Read page", func(cm log.CM) {
			cm.Write(zap.Int64("worker.ProcessedPages", worker.ProcessedPages))
		})

		workerModifiers[2] = []interface{}{"OFFSET", worker.ProcessedPages}
		userTokens, err := models.GetUserTokensBatchByFilters(
			worker.Db, message.App, message.Service, filters, workerModifiers,
		)
		if err != nil {
			log.E(l, "Failed to get user tokens by filters and modifiers", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return outChan, err
		}

		log.D(l, "pgRead", func(cm log.CM) {
			cm.Write(zap.Object("len(userTokens)", len(userTokens)))
		})

		for _, userToken := range userTokens {
			message.Token = userToken.Token
			message.Locale = userToken.Locale

			strMsg, err := json.Marshal(message)
			if err != nil {
				log.E(l, "Failed to marshal msg", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return outChan, err
			}
			log.D(l, "pgRead - Send message to channel", func(cm log.CM) {
				cm.Write(zap.String("strMsg", string(strMsg)))
			})
			outChan <- string(strMsg)
			worker.TotalProcessedTokens++
		}
	}
	return outChan, nil
}
