package kafka

import (
	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
	"github.com/uber-go/zap"
)

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(kafkaConfig *ProducerConfig, inChan <-chan *messages.KafkaMessage) {

	saramaConfig := sarama.NewConfig()
	producer, err := sarama.NewSyncProducer(kafkaConfig.Brokers, saramaConfig)
	if err != nil {
		Logger.Error(
			"Failed to start kafka produces",
			zap.Error(err),
		)
		return
	}
	defer producer.Close()

	for msg := range inChan {
		saramaMessage := &sarama.ProducerMessage{
			Topic: msg.Topic,
			Value: sarama.StringEncoder(msg.Message),
		}

		_, _, err = producer.SendMessage(saramaMessage)
		if err != nil {
			Logger.Error(
				"Error sending message",
				zap.Error(err),
			)
		} else {
			Logger.Info(
				"Sent message",
				zap.String("topic", msg.Topic),
			)
		}
	}
}
