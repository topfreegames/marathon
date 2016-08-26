package producer_test

import (
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	"git.topfreegames.com/topfreegames/marathon/kafka/producer"
	"git.topfreegames.com/topfreegames/marathon/messages"

	mt "git.topfreegames.com/topfreegames/marathon/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var _ = Describe("Producer", func() {
	var (
		l zap.Logger
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
	})
	It("Should send messages received in the inChan to kafka", func() {
		app := "producerApp1"
		service := "gcm"
		message := "message%d"

		var config = viper.New()
		config.SetConfigFile("./../../config/test.yaml")
		config.SetEnvPrefix("marathon")
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()
		err := config.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())

		config.Set("workers.consumer.topicTemplate", "%s-%s")
		topicTemplate := config.GetString("workers.consumer.topicTemplate")
		topic := fmt.Sprintf(topicTemplate, app, service)

		inChan := make(chan *messages.KafkaMessage, 10)
		doneChanProd := make(chan struct{}, 1)
		defer close(doneChanProd)

		go producer.Producer(l, config, inChan, doneChanProd)
		message1 := fmt.Sprintf(message, 1)
		message2 := fmt.Sprintf(message, 2)
		msg1 := &messages.KafkaMessage{Message: message1, Topic: topic}
		msg2 := &messages.KafkaMessage{Message: message2, Topic: topic}
		inChan <- msg1
		inChan <- msg2

		outChan := make(chan string, 10)
		doneChanCons := make(chan struct{}, 1)
		defer close(doneChanCons)
		go consumer.Consumer(l, config, app, service, outChan, doneChanCons)

		consumedMessage1 := <-outChan
		consumedMessage2 := <-outChan
		Expect(consumedMessage1).To(Equal(message1))
		Expect(consumedMessage2).To(Equal(message2))
	})

	It("Should not create a producer if no broker found", func() {
		brokers := []string{"localhost:3456"}

		var producerConfig = viper.New()
		producerConfig.Set("workers.producer.brokers", brokers)

		inChan := make(chan *messages.KafkaMessage, 10)
		doneChan := make(chan struct{}, 1)
		defer close(doneChan)

		// Producer returns here and don't get blocked
		producer.Producer(l, producerConfig, inChan, doneChan)
	})
})
