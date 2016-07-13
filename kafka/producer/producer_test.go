package producer_test

import (
	"fmt"
	"testing"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/kafka/producer"
	"git.topfreegames.com/topfreegames/marathon/messages"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestProducer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Producer")
}

var _ = Describe("Producer", func() {

	It("Should send messages received in the inChan to kafka", func() {
		topic := "test-producer-1"
		topics := []string{topic}
		brokers := []string{"localhost:3536"}
		consumerGroup := "consumer-group-test-producer-1"
		message := "message%d"

		var producerConfig = viper.New()
		producerConfig.SetDefault("workers.producer.brokers", brokers)

		inChan := make(chan *messages.KafkaMessage)
		defer close(inChan)

		go producer.Producer(producerConfig, inChan)
		message1 := fmt.Sprintf(message, 1)
		message2 := fmt.Sprintf(message, 1)
		msg1 := &messages.KafkaMessage{Message: message1, Topic: topic}
		msg2 := &messages.KafkaMessage{Message: message2, Topic: topic}
		inChan <- msg1
		inChan <- msg2

		// Consuming
		var consumerConfig = viper.New()
		consumerConfig.SetDefault("workers.consumer.brokers", brokers)
		consumerConfig.SetDefault("workers.consumer.consumergroup", consumerGroup)
		consumerConfig.SetDefault("workers.consumer.topics", topics)

		outChan := make(chan string, 10)
		defer close(outChan)
		done := make(chan struct{}, 1)
		defer close(done)
		go consumer.Consumer(consumerConfig, outChan, done)

		consumedMessage1 := <-outChan
		consumedMessage2 := <-outChan
		Expect(consumedMessage1).To(Equal(message1))
		Expect(consumedMessage2).To(Equal(message2))
	})

	It("Should not create a producer if no broker found", func() {
		brokers := []string{"localhost:3555"}

		var producerConfig = viper.New()
		producerConfig.SetDefault("workers.producer.brokers", brokers)

		inChan := make(chan *messages.KafkaMessage)
		defer close(inChan)

		// Producer returns here and don't get blocked
		producer.Producer(producerConfig, inChan)
	})
})
