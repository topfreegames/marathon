package kafka

import (
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	. "github.com/franela/goblin"
)

func produceMessage(t *testing.T, brokers []string, topic string, message string,
	partition int32, offset int64) {
	g := Goblin(t)
	// Produce message to the test
	producerConfig := sarama.NewConfig()
	producerConfig.Version = sarama.V0_9_0_0
	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	g.Assert(err == nil).IsTrue()
	saramaMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	part, off, err := producer.SendMessage(saramaMessage)
	g.Assert(err).Equal(nil)
	g.Assert(part).Equal(partition)
	g.Assert(off).Equal(offset)
}

func TestConsumer(t *testing.T) {
	g := Goblin(t)

	g.Describe("Consumer.Consume", func() {
		g.It("Should consume message correctly and retrieve it", func() {
			kafkaMsg := &sarama.ConsumerMessage{Value: []byte("test")}
			_, err := Consume(kafkaMsg)
			g.Assert(err).Equal(nil)
		})

		g.It("Should not consume an empty message", func() {
			kafkaMsg := &sarama.ConsumerMessage{Value: []byte("")}
			msg, err := Consume(kafkaMsg)
			g.Assert(msg).Equal("")
			g.Assert(err == nil).IsFalse()
		})

		g.It("Should consume one message correctly and retrieve it", func() {
			topic := "test-consumer-1"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test"

			config := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)

			go Consumer(config, outChan, done)
		})
	})

	g.Describe("Consumer", func() {
		g.It("Should consume one message correctly and retrieve it", func() {
			topic := "test-consumer-1"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test"

			config := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go Consumer(config, outChan, done)

			// Produce Messages
			message := "message"
			produceMessage(t, brokers, topic, message, int32(0), int64(0))

			consumedMessage := <-outChan
			g.Assert(consumedMessage).Equal(message)
		})

		g.It("Should consume two messages correctly and retrieve them", func() {
			topic := "test-consumer-2"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test"
			message := "message%d"

			config := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go Consumer(config, outChan, done)

			// Produce Messages
			message1 := fmt.Sprintf(message, 1)
			message2 := fmt.Sprintf(message, 1)
			produceMessage(t, brokers, topic, message1, int32(0), int64(0))
			produceMessage(t, brokers, topic, message2, int32(0), int64(1))

			consumedMessage1 := <-outChan
			consumedMessage2 := <-outChan
			g.Assert(consumedMessage1).Equal(message1)
			g.Assert(consumedMessage2).Equal(message1)
		})

		g.It("Should not output an empty message", func() {
			topic := "test-consumer-3"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test"

			config := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go Consumer(config, outChan, done)

			produceMessage(t, brokers, topic, "", int32(0), int64(0))
			produceMessage(t, brokers, topic, "message", int32(0), int64(1))

			// The channel receives only non-empty messages
			consumedMessage := <-outChan
			g.Assert(consumedMessage).Equal("message")
		})

		g.It("Should not start a consumer if no broker found", func() {
			topic := "test-consumer-3"
			topics := []string{topic}
			brokers := []string{"localhost:0666"}
			consumerGroup := "consumer-group-test"

			config := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}

			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)

			// Consumer returns here and don't get blocked
			Consumer(config, outChan, done)
		})
	})
}
