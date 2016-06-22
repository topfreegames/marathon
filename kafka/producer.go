package kafka

import (
	"github.com/golang/glog"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/Shopify/sarama"
)

// Producer continuosly reads from inChan and sends the received messages to kafka
func Producer(kafkaConfig *ProducerConfig, inChan <-chan *messages.KafkaMessage) {
	saramaConfig := sarama.NewConfig()
	producer, err := sarama.NewSyncProducer(kafkaConfig.Brokers, saramaConfig)
	if err != nil {
		glog.Errorf("Failed to start kafka produces: %v", err)
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
			glog.Errorf("Error sending message %+v", saramaMessage)
		} else {
			glog.Infof("Sent message to topic: %s", msg.Topic)
		}
	}
}
