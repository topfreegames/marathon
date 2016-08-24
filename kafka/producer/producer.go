package producer

import (
	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(l zap.Logger, config *viper.Viper, inChan <-chan *messages.KafkaMessage, doneChan <-chan struct{}) {
	l.Info("Starting producer")
	saramaConfig := sarama.NewConfig()
	l = l.With(zap.Object("saramaConfig", saramaConfig))
	producer, err := sarama.NewSyncProducer(
		config.GetStringSlice("workers.producer.brokers"),
		saramaConfig)
	if err != nil {
		l.Error("Failed to start kafka producer", zap.Error(err))
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
			l = l.With(zap.Object("KafkaMessage", saramaMessage))

			_, _, err = producer.SendMessage(saramaMessage)
			if err != nil {
				l.Error("Error sending message", zap.Error(err))
			} else {
				l.Info("Sent message", zap.String("topic", msg.Topic))
			}
		}
	}
}
