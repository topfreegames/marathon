package consumer_test

import (
	"fmt"
	"testing"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConsumer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consumer")
}

func produceMessage(brokers []string, topic string, message string, partition int32, offset int64) {
	// Produce message to the test
	producerConfig := sarama.NewConfig()
	producerConfig.Version = sarama.V0_9_0_0
	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	Expect(err).To(BeNil())
	defer producer.Close()

	saramaMessage := sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	part, off, err := producer.SendMessage(&saramaMessage)
	Expect(err).To(BeNil())
	Expect(part).To(Equal(partition))
	Expect(off).To(Equal(offset))
}

var _ = Describe("Consumer", func() {
	Describe("Consume", func() {
		It("Should not consume an empty message", func() {
			message := ""
			kafkaMsg := sarama.ConsumerMessage{Value: []byte(message)}
			_, err := consumer.Consume(&kafkaMsg)
			Expect(err).NotTo(BeNil())
		})
		It("Should consume one message correctly and retrieve it", func() {
			message := "message"
			kafkaMsg := sarama.ConsumerMessage{Value: []byte(message)}
			msg, err := consumer.Consume(&kafkaMsg)
			Expect(err).To(BeNil())
			Expect(msg).To(Equal(message))
		})
	})

	Describe("", func() {
		It("Should consume one message correctly and retrieve it", func() {
			topic := "test-consumer-1"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test-consumer-1"

			consumerConfig := consumer.Config{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go consumer.Consumer(&consumerConfig, outChan, done)

			// Produce Messages
			message := "message"
			produceMessage(brokers, topic, message, int32(0), int64(0))

			consumedMessage := <-outChan
			Expect(consumedMessage).To(Equal(message))
		})

		It("Should consume two messages correctly and retrieve them", func() {
			topic := "test-consumer-2"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test-consumer-2"
			message := "message%d"

			consumerConfig := consumer.Config{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go consumer.Consumer(&consumerConfig, outChan, done)

			// Produce Messages
			message1 := fmt.Sprintf(message, 1)
			message2 := fmt.Sprintf(message, 1)
			produceMessage(brokers, topic, message1, int32(0), int64(0))
			produceMessage(brokers, topic, message2, int32(0), int64(1))

			consumedMessage1 := <-outChan
			consumedMessage2 := <-outChan
			Expect(consumedMessage1).To(Equal(message1))
			Expect(consumedMessage2).To(Equal(message2))
		})

		It("Should not output an empty message", func() {
			topic := "test-consumer-3"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test-consumer-3"
			message := "message"

			consumerConfig := consumer.Config{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go consumer.Consumer(&consumerConfig, outChan, done)

			produceMessage(brokers, topic, "", int32(0), int64(0))
			produceMessage(brokers, topic, "message", int32(0), int64(1))

			// The channel receives only non-empty messages
			consumedMessage := <-outChan
			Expect(consumedMessage).To(Equal(message))
		})

		It("Should not start a consumer if no broker found", func() {
			topic := "test-consumer-4"
			topics := []string{topic}
			brokers := []string{"localhost:0666"}
			consumerGroup := "consumer-group-test-consumer-4"

			consumerConfig := consumer.Config{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)

			// Consumer returns here and don't get blocked
			consumer.Consumer(&consumerConfig, outChan, done)
		})
	})
})
