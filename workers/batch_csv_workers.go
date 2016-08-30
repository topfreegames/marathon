package workers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/kafka/producer"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/templates"
	"git.topfreegames.com/topfreegames/marathon/util"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/minio/minio-go"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// BatchCsvWorker contains all batch csv worker configs and channels
type BatchCsvWorker struct {
	ID                           uuid.UUID
	Config                       *viper.Viper
	Logger                       zap.Logger
	Notifier                     *models.Notifier
	App                          *models.App
	Message                      *messages.InputMessage
	Modifiers                    [][]interface{}
	StartedAt                    int64
	Bucket                       string
	Key                          string
	Reader                       *csv.Reader
	KafkaClient                  *cluster.Client
	KafkaTopic                   string
	InitialKafkaOffset           int64
	CurrentKafkaOffset           int64
	Db                           *models.DB
	ConfigPath                   string
	RedisClient                  *util.RedisClient
	TotalTokens                  int64
	ProcessedTokens              int
	TotalProcessedTokens         int
	TotalPages                   int64
	ProcessedPages               int64
	BatchCsvWorkerStatusDoneChan chan struct{}
	BatchCsvWorkerDoneChan       chan struct{}
	CsvToParserChan              chan string
	ParserDoneChan               chan struct{}
	ParserToFetcherChan          chan *messages.InputMessage
	FetcherDoneChan              chan struct{}
	FetcherToBuilderChan         chan *messages.TemplatedMessage
	BuilderDoneChan              chan struct{}
	BuilderToProducerChan        chan *messages.KafkaMessage
	ProducerDoneChan             chan struct{}
}

func (worker *BatchCsvWorker) createChannels() {
	worker.BatchCsvWorkerStatusDoneChan = make(chan struct{}, worker.Config.GetInt("batchcsvworkerstatusdonechan"))
	worker.BatchCsvWorkerDoneChan = make(chan struct{}, worker.Config.GetInt("batchcsvworkerdonechansize"))
	worker.CsvToParserChan = make(chan string, worker.Config.GetInt("csvtoparserchansize"))
	worker.ParserDoneChan = make(chan struct{}, worker.Config.GetInt("parserdonechansize"))
	worker.ParserToFetcherChan = make(chan *messages.InputMessage, worker.Config.GetInt("parsertofetcherchansize"))
	worker.FetcherDoneChan = make(chan struct{}, worker.Config.GetInt("fetcherdonechansize"))
	worker.FetcherToBuilderChan = make(chan *messages.TemplatedMessage, worker.Config.GetInt("fetchertobuilderchansize"))
	worker.BuilderDoneChan = make(chan struct{}, worker.Config.GetInt("builderdonechansize"))
	worker.BuilderToProducerChan = make(chan *messages.KafkaMessage, worker.Config.GetInt("buildertoproducerchansize"))
	worker.ProducerDoneChan = make(chan struct{}, worker.Config.GetInt("producerdonechansize"))
}

func (worker *BatchCsvWorker) connectDatabase() {
	host := worker.Config.GetString("postgres.host")
	user := worker.Config.GetString("postgres.user")
	dbName := worker.Config.GetString("postgres.dbname")
	password := worker.Config.GetString("postgres.password")
	port := worker.Config.GetInt("postgres.port")
	sslMode := worker.Config.GetString("postgres.sslMode")

	db, err := models.GetDB(worker.Logger, host, user, port, sslMode, dbName, password)

	if err != nil {
		worker.Logger.Panic(
			"Could not connect to postgres...",
			zap.String("host", host),
			zap.Int("port", port),
			zap.String("user", user),
			zap.String("dbName", dbName),
			zap.String("error", err.Error()),
		)
	}
	worker.Db = db
}

func (worker *BatchCsvWorker) connectRedis() {
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
	rl.Debug("Connecting to redis...")
	cli, err := util.GetRedisClient(redisHost, redisPort, redisPass, redisDB, redisMaxPoolSize, rl)
	if err != nil {
		rl.Panic(
			"Could not connect to redis...",
			zap.String("error", err.Error()),
		)
	}
	worker.RedisClient = cli
	rl.Info("Connected to redis successfully.")
}

// TODO: Set all default configs
func (worker *BatchCsvWorker) setConfigurationDefaults() {
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
	worker.Config.SetDefault("workers.modules.batchcsvworkerdonechan", 1)
	worker.Config.SetDefault("workers.modules.csvtoparserchan", 10000)
	worker.Config.SetDefault("workers.modules.parserdonechan", 1)
	worker.Config.SetDefault("workers.modules.parsertofetcherchan", 1000)
	worker.Config.SetDefault("workers.modules.fetcherdonechan", 1)
	worker.Config.SetDefault("workers.modules.fetchertobuilderchan", 1000)
	worker.Config.SetDefault("workers.modules.builderdonechan", 1)
	worker.Config.SetDefault("workers.modules.buildertoproducerchan", 1000)
	worker.Config.SetDefault("workers.modules.producerdonechan", 1)
}

func (worker *BatchCsvWorker) loadConfiguration() {
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

func (worker *BatchCsvWorker) getCsvFromS3() {
	s3AccessKeyID := worker.Config.GetString("s3.accessKey")
	s3SecretAccessKey := worker.Config.GetString("s3.secretAccessKey")
	ssl := true
	s3Client, err := minio.New("s3.amazonaws.com", s3AccessKeyID, s3SecretAccessKey, ssl)
	if err != nil {
		worker.Logger.Panic(
			"Could not authenticate with S3...",
			zap.String("error", err.Error()),
		)
	}

	csvFile, err := s3Client.GetObject(worker.Bucket, worker.Key)
	if err != nil {
		worker.Logger.Panic(
			"Could not download csv from S3...",
			zap.String("bucket", worker.Bucket),
			zap.String("key", worker.Key),
			zap.String("error", err.Error()),
		)
	}

	// TODO: get total tokerns
	worker.TotalTokens = 0
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}
	for {
		c, err := csvFile.Read(buf)
		worker.TotalTokens += int64(bytes.Count(buf[:c], lineSep))
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}

	// TODO: Bad idea, downloading file twice...
	csvFile, err = s3Client.GetObject(worker.Bucket, worker.Key)
	if err != nil {
		worker.Logger.Panic(
			"Could not download csv from S3...",
			zap.String("bucket", worker.Bucket),
			zap.String("key", worker.Key),
			zap.String("error", err.Error()),
		)
	}

	worker.Reader = csv.NewReader(csvFile)
}

func (worker *BatchCsvWorker) configureKafkaClient() {
	clusterConfig := cluster.NewConfig()
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Group.Return.Notifications = true
	clusterConfig.Version = sarama.V0_9_0_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	brokersString := worker.Config.GetString("workers.consumer.brokers")
	brokers := strings.Split(brokersString, ",")
	// consumerGroupTemlate := worker.Config.GetString("workers.consumer.consumergroupTemplate")
	topicTemplate := worker.Config.GetString("workers.consumer.topicTemplate")
	topic := fmt.Sprintf(topicTemplate, worker.App.Name, worker.Notifier.Service)
	// consumerGroup := fmt.Sprintf(consumerGroupTemlate, worker.App.Name, worker.Notifier.Service)
	worker.KafkaTopic = topic
	client, err := cluster.NewClient(brokers, clusterConfig)
	if err != nil {
		worker.Logger.Error(
			"Could not create kafka client",
			zap.String("error", err.Error()),
		)
	}
	worker.Logger.Debug(
		"Created kafka client",
		zap.String("client", fmt.Sprintf("%+v", client)),
	)

	currentOffset, err := client.GetOffset(topic, 0, sarama.OffsetNewest)
	if err != nil {
		worker.Logger.Error(
			"Could not get kafka offset",
			zap.String("error", err.Error()),
		)
	}
	worker.Logger.Debug(
		"Got kafka offset",
		zap.Int64("Offset", currentOffset),
	)

	worker.InitialKafkaOffset = currentOffset
	worker.KafkaClient = client
}

// Configure configures the worker
func (worker *BatchCsvWorker) Configure() {
	worker.ID = uuid.NewV4()
	worker.setConfigurationDefaults()
	worker.Logger = ConfigureLogger(Log{Level: "warn"}, worker.Config)
	worker.loadConfiguration()
	worker.Logger = ConfigureLogger(Log{}, worker.Config)
	worker.createChannels()
	worker.connectDatabase()
	worker.connectRedis()
	worker.getCsvFromS3()
	worker.configureKafkaClient()
}

// Start starts the workers according to the configuration and returns the workers object
func (worker *BatchCsvWorker) Start() {
	requireToken := false

	worker.StartedAt = time.Now().Unix()

	worker.Logger.Info("Starting worker pipeline...")

	go worker.updateStatus(worker.BatchCsvWorkerStatusDoneChan)

	// Run modules
	qtyParsers := worker.Config.GetInt("workers.modules.parsers")
	worker.Logger.Debug("Starting parser...", zap.Int("quantity", qtyParsers))
	for i := 0; i < qtyParsers; i++ {
		go templates.Parser(worker.Logger, requireToken, worker.CsvToParserChan, worker.ParserToFetcherChan, worker.ParserDoneChan)
	}
	worker.Logger.Debug("Started parser...", zap.Int("quantity", qtyParsers))

	qtyFetchers := worker.Config.GetInt("workers.modules.fetchers")
	worker.Logger.Debug("Starting fetcher...", zap.Int("quantity", qtyFetchers))
	for i := 0; i < qtyFetchers; i++ {
		go templates.Fetcher(worker.Logger, worker.ParserToFetcherChan, worker.FetcherToBuilderChan, worker.FetcherDoneChan, worker.Db)
	}
	worker.Logger.Debug("Started fetcher...", zap.Int("quantity", qtyFetchers))

	qtyBuilders := worker.Config.GetInt("workers.modules.builders")
	worker.Logger.Debug("Starting builder...", zap.Int("quantity", qtyBuilders))
	for i := 0; i < qtyBuilders; i++ {
		go templates.Builder(worker.Logger, worker.Config, worker.FetcherToBuilderChan, worker.BuilderToProducerChan, worker.BuilderDoneChan)
	}
	worker.Logger.Debug("Started builder...", zap.Int("quantity", qtyBuilders))

	qtyProducers := worker.Config.GetInt("workers.modules.producers")
	worker.Logger.Debug("Starting producer...", zap.Int("quantity", qtyProducers))
	for i := 0; i < qtyProducers; i++ {
		go producer.Producer(worker.Logger, worker.Config, worker.BuilderToProducerChan, worker.ProducerDoneChan)
	}
	worker.Logger.Debug("Started producer...", zap.Int("quantity", qtyProducers))

	worker.Logger.Debug("Starting csvReader...")
	go worker.csvReader(worker.Message, worker.Modifiers, worker.CsvToParserChan)
	worker.Logger.Debug("Started csvReader...")

	worker.Logger.Info("Started worker pipeline...")
}

// Close stops the modules of the instance
func (worker BatchCsvWorker) Close() {
	worker.Logger.Info("Stopping workers")
	// close(worker.CsvToKafkaChan)
	// close(worker.DoneChan)
	worker.Logger.Info("Stopped workers")
}

// GetBatchCsvWorker returns a new worker
func GetBatchCsvWorker(worker *BatchCsvWorker) (*BatchCsvWorker, error) {
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

// GetWorkerStatus returns a map[string]interface{} with the current worker status
func (worker BatchCsvWorker) GetWorkerStatus() map[string]interface{} {
	return map[string]interface{}{
		"key":                  worker.Key,
		"bucket":               worker.Bucket,
		"notificationID":       worker.ID,
		"startedAt":            worker.StartedAt,
		"totalTokens":          worker.TotalTokens,
		"totalProcessedTokens": worker.TotalProcessedTokens,
		"totalPages":           worker.TotalPages,
		"processedPages":       worker.ProcessedPages,
		"message":              worker.Message,
	}
}

// GetKafkaStatus returns a map[string]interface{} with the current kafka status
func (worker BatchCsvWorker) GetKafkaStatus() map[string]interface{} {
	currentOffset, err := worker.KafkaClient.GetOffset(worker.KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		worker.Logger.Error(
			"Could not get kafka offset",
			zap.String("error", err.Error()),
		)
	}
	worker.Logger.Debug(
		"Got kafka offset",
		zap.Int64("Offset", currentOffset),
	)

	worker.CurrentKafkaOffset = currentOffset
	return map[string]interface{}{
		"kafkaTopic":         worker.KafkaTopic,
		"initialKafkaOffset": worker.InitialKafkaOffset,
		"currentKafkaOffset": currentOffset,
	}
}

// SetStatus sets the current notification status in redis
func (worker *BatchCsvWorker) SetStatus() {
	cli := worker.RedisClient.Client
	redisKey := strings.Join([]string{worker.Notifier.ID.String(), worker.ID.String()}, "|")

	worker.Logger.Info("Set in redis", zap.String("key", redisKey))

	workerStatus := worker.GetWorkerStatus()
	byteWorkerStatus, err := json.Marshal(workerStatus)
	if err != nil {
		worker.Logger.Panic("Could not parse worker status", zap.Error(err))
	}
	workerStrStatus := string(byteWorkerStatus)

	kafkaStatus := worker.GetKafkaStatus()
	byteKafkaStatus, err := json.Marshal(kafkaStatus)
	if err != nil {
		worker.Logger.Panic("Could not parse kafka status", zap.Error(err))
	}
	kafkaStrStatus := string(byteKafkaStatus)

	status := map[string]interface{}{
		"workerStatus": workerStrStatus,
		"kafkaStatus":  kafkaStrStatus,
	}
	byteStatus, err := json.Marshal(status)
	if err != nil {
		worker.Logger.Panic("Could not parse status", zap.Error(err))
	}
	strStatus := string(byteStatus)

	// FIXME: What's the best TTL to set? 30 * time.Day ?
	if err = cli.Set(redisKey, strStatus, 0).Err(); err != nil {
		worker.Logger.Panic("Failed to set notification key in redis", zap.Error(err))
	}
}

func (worker *BatchCsvWorker) updateStatus(doneChan <-chan struct{}) {
	worker.Logger.Info("Starting status updater")
	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		default:
			worker.Logger.Debug("Update worker status")
			worker.SetStatus()
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// csvReader reads from csv in batches and sends the built messages to kafka
func (worker *BatchCsvWorker) csvReader(message *messages.InputMessage,
	modifiers [][]interface{}, outChan chan<- string) (chan<- string, error) {
	l := worker.Logger.With(
		zap.Object("message", message),
		zap.Object("modifiers", modifiers),
	)

	limit := -1
	for _, modifier := range modifiers {
		if modifier[0] == "LIMIT" {
			limit = modifier[1].(int)
		}
	}
	if limit <= 0 {
		worker.Logger.Fatal("Limit should be greater than 0", zap.Int("limit", limit))
	}
	l = l.With(zap.Int("limit", limit))

	worker.ProcessedTokens = 0

	l = l.With(zap.Int64("worker.TotalTokens", worker.TotalTokens))

	eof := false
	for eof == false {
		l.Debug("csvRead - Read page")

		userIDs := make([]string, 0, limit)
		for i := 0; i < limit; i++ {
			row, err := worker.Reader.Read()
			if err == io.EOF {
				eof = true
				break
			}
			if err != nil {
				l.Error(
					"Error reading from csv",
					zap.Error(err),
				)
				return outChan, err
			}
			// TODO: If more than one element per line this needs to be changed
			userIDs = append(userIDs, row[0])
		}

		userTokens, err := models.GetUserTokenBatchByUserID(
			worker.Db, message.App, message.Service, userIDs,
		)
		if err != nil {
			l.Error(
				"Failed to get user tokens by modifiers",
				zap.Error(err),
			)
			return outChan, err
		}

		l.Debug("csvRead", zap.Object("len(userTokens)", len(userTokens)))

		for i := 0; i < len(userTokens); {
			for _, userToken := range userTokens {
				message.Token = userToken.Token
				message.Locale = userToken.Locale

				strMsg, err := json.Marshal(message)
				if err != nil {
					l.Error("Failed to marshal msg", zap.Error(err))
					return outChan, err
				}
				l.Debug("csvRead - Send message to channel", zap.String("strMsg", string(strMsg)))
				outChan <- string(strMsg)
				i++
				worker.ProcessedTokens++
			}
		}
	}
	return outChan, nil
}
