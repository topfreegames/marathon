package producer

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/extensions"
	"git.topfreegames.com/topfreegames/marathon/log"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(l zap.Logger, config *viper.Viper, inChan <-chan *messages.KafkaMessage, doneChan <-chan struct{}) {
	log.I(l, "Starting producer")
	saramaConfig := sarama.NewConfig()
	//TODO: replace
	zkClient := extensions.GetZkClient(config.ConfigFileUsed())

	brokers, err := zkClient.GetKafkaBrokers()

	if err != nil {
		panic(err)
	}
	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	zookeeperHosts := config.GetStringSlice("workers.zookeeperHosts")
	c, _, err := zk.Connect(zookeeperHosts, time.Second*10)
	if err != nil {
		log.P(l, "error connecting to zookeeper", func(cm log.CM) {
			cm.Write(
				zap.Object("zookeeperHosts", zookeeperHosts),
			)
		})
	}
	children, _, err := c.Children("/brokers/ids")
	if err != nil {
		panic(err)
	}
	log.D(l, "got brokers from zookeeper", func(cm log.CM) {
		cm.Write(zap.Object("brokers", children))
	})
	//
	if err != nil {
		log.E(l, "Failed to start kafka producer", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}
	defer producer.Close()

	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		case msg := <-inChan:
			saramaMessage := &sarama.ProducerMessage{
				Topic: msg.Topic,
				Value: sarama.StringEncoder(msg.Message),
			}

			partition, offset, err := producer.SendMessage(saramaMessage)
			if err != nil {
				log.E(l, "Error sending message", func(cm log.CM) {
					cm.Write(zap.Object("KafkaMessage", saramaMessage), zap.Error(err))
				})
			} else {
				log.I(l, "Sent message", func(cm log.CM) {
					cm.Write(
						zap.Object("KafkaMessage", saramaMessage),
						zap.String("topic", msg.Topic),
						zap.Int("partition", int(partition)),
						zap.Int64("offset", offset),
					)
				})
			}
		}
	}
}
