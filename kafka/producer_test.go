package kafka

import (
	"fmt"
	"testing"

	"git.topfreegames.com/topfreegames/marathon/messages"

	. "github.com/franela/goblin"
)

func TestProducer(t *testing.T) {
	g := Goblin(t)

	g.Describe("Producer", func() {
		g.It("Should send messages received in the inChan to kafka", func() {
			topic := "test-producer-1"
			topics := []string{topic}
			brokers := []string{"localhost:3536"}
			consumerGroup := "consumer-group-test"
			message := "message%d"

			producerConfig := &ProducerConfig{Brokers: brokers}
			inChan := make(chan *messages.KafkaMessage)
			defer close(inChan)
			go Producer(producerConfig, inChan)
			message1 := fmt.Sprintf(message, 1)
			message2 := fmt.Sprintf(message, 1)
			msg1 := &messages.KafkaMessage{Message: message1, Topic: topic}
			msg2 := &messages.KafkaMessage{Message: message2, Topic: topic}
			inChan <- msg1
			inChan <- msg2

			consumerConfig := &ConsumerConfig{
				ConsumerGroup: consumerGroup,
				Topics:        topics,
				Brokers:       brokers,
			}
			outChan := make(chan string, 10)
			defer close(outChan)
			done := make(chan struct{}, 1)
			defer close(done)
			go Consumer(consumerConfig, outChan, done)

			consumedMessage1 := <-outChan
			consumedMessage2 := <-outChan
			g.Assert(consumedMessage1 == message1).IsTrue()
			g.Assert(consumedMessage2 == message2).IsTrue()
		})

		g.It("Should not create a producer if no broker found", func() {
			brokers := []string{"localhost:0666"}

			producerConfig := &ProducerConfig{Brokers: brokers}
			inChan := make(chan *messages.KafkaMessage)
			defer close(inChan)
			Producer(producerConfig, inChan)
		})
	})
}
