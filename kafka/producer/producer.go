package producer

import (
	"strings"

	"git.topfreegames.com/topfreegames/marathon/log"
	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(l zap.Logger, config *viper.Viper, inChan <-chan *messages.KafkaMessage, doneChan <-chan struct{}) {
	log.I(l, "Starting producer")
	saramaConfig := sarama.NewConfig()
	brokersStr := config.GetString("workers.producer.brokers")
	brokers := strings.Split(brokersStr, ",")
	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
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
