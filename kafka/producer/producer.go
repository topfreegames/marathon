package producer

import (
	"os"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
	"github.com/uber-go/zap"
)

func getLogLevel() zap.Level {
	var level = zap.WarnLevel
	var environment = os.Getenv("ENV")
	if environment == "test" {
		level = zap.FatalLevel
	}
	return level
}

// Logger is the producer logger
var Logger = zap.NewJSON(getLogLevel())

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(kafkaConfig *Config, inChan <-chan *messages.KafkaMessage) {

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
