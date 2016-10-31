package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

//ZkKafkaInfo a struct that will be used to unmashal the info from zk, e.g.	"{\"jmx_port\":-1,\"timestamp\":\"1477940170730\",\"endpoints\":[\"PLAINTEXT://10.0.22.198:9940\"],\"host\":\"10.0.22.198\",\"version\":2,\"port\":9940}"
type ZkKafkaInfo struct {
	JmxPort   int      `json:"jmx_port"`
	Timestamp int64    `json:"timestamp"`
	Endpoints []string `json:"endpoints"`
	Host      string   `json:"host"`
	Version   int      `json:"version"`
	Port      int      `json:"port"`
}

// ZkClient is the struct that will keep a zk Conn
type ZkClient struct {
	config     *viper.Viper
	configPath string
	Conn       *zk.Conn
	Logger     zap.Logger
}

var once sync.Once
var zkClient *ZkClient
var config *viper.Viper

// GetZkClient gets a kzClient configure
func GetZkClient(configPath string) *ZkClient {
	once.Do(func() {
		if configPath == "" {
			panic(errors.New("No configPath passed to configure ZkClient"))
		}
		zkClient = &ZkClient{
			configPath: configPath,
		}
		zkClient.configureLogger()
		zkClient.loadConfiguration(configPath)
		zkClient.loadDefaults()
		zkClient.configureConn()
	})
	return zkClient
}

func (z *ZkClient) loadDefaults() {
	z.config.SetDefault("workers.zookeeperHosts", []string{"localhost:2181"})
}

func (z *ZkClient) configureLogger() {
	z.Logger = zap.New(
		zap.NewJSONEncoder(),
		zap.DebugLevel,
		zap.AddCaller(),
	)
}

func (z *ZkClient) loadConfiguration(configPath string) {
	z.config = viper.New()
	z.config.SetConfigFile(z.configPath)
	z.config.SetConfigType("yaml")
	z.config.SetEnvPrefix("marathon")
	z.config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	z.config.AutomaticEnv()
	err := z.config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func (z *ZkClient) configureConn() {
	zookeeperHosts := z.config.GetStringSlice("workers.zookeeperHosts")
	z.Logger.Debug("Connecting to zookeeper", zap.Object("zookeeperHosts", zookeeperHosts))
	zk, _, err := zk.Connect(zookeeperHosts, time.Second*10)
	if err != nil {
		panic(err)
	}
	z.Logger.Debug("Connected to zookeeper", zap.Object("zookeeperHosts", zookeeperHosts))
	z.Conn = zk
}

//GetKafkaBrokers gets a slice with the hostname of the kafka brokers
func (z *ZkClient) GetKafkaBrokers() ([]string, error) {
	if zkClient == nil {
		return nil, errors.New("Zookeeper extension is not initialized yet, call GetZkClient first!")
	}

	ids, _, err := z.Conn.Children("/brokers/ids")

	if err != nil {
		z.Logger.Error("Error getting kafka brokers", zap.Error(err))
		return nil, err
	}

	z.Logger.Debug("Got broker ids from zookeeper, pg", zap.Object("brokerIds", ids))

	var brokers []string

	for _, id := range ids {
		z.Logger.Debug("Grabbing broker info", zap.Object("brokerId", id))
		info, _, err := z.Conn.Get(fmt.Sprintf("%s%s", "/brokers/ids/", id))
		if err != nil {
			return nil, err
		}

		var kafkaInfo ZkKafkaInfo
		json.Unmarshal(info, &kafkaInfo)
		z.Logger.Debug("Got Broker info",
			zap.Object("kafkaInfo", kafkaInfo),
		)
		brokers = append(brokers, fmt.Sprintf("%s:%d", kafkaInfo.Host, kafkaInfo.Port))
	}
	return brokers, nil
}
