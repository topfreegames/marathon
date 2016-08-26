package consumer_test

import (
	"fmt"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/kafka/consumer"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

func produceMessage(brokers []string, topic string, message string, partition int32, offset int64) {
	// Produce message to the test
	producerConfig := sarama.NewConfig()
	producerConfig.Version = sarama.V0_9_0_0
	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	Expect(err).NotTo(HaveOccurred())
	defer producer.Close()

	saramaMessage := sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	part, off, err := producer.SendMessage(&saramaMessage)
	Expect(err).NotTo(HaveOccurred())
	Expect(part).To(Equal(partition))
	Expect(off).To(Equal(offset))
}

var _ = Describe("Consumer", func() {
	var (
		l      zap.Logger
		config *viper.Viper
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
		config = viper.New()
		config.SetConfigFile("./../../config/test.yaml")
		config.SetEnvPrefix("marathon")
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()
		err := config.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())
	})
	Describe("Consume", func() {
		It("Should not consume an empty message", func() {
			message := ""
			kafkaMsg := sarama.ConsumerMessage{Value: []byte(message)}
			_, err := consumer.Consume(l, &kafkaMsg)
			Expect(err).To(HaveOccurred())
		})
		It("Should consume one message correctly and retrieve it", func() {
			message := "message"
			kafkaMsg := sarama.ConsumerMessage{Value: []byte(message)}
			msg, err := consumer.Consume(l, &kafkaMsg)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(Equal(message))
		})
	})

	Describe("", func() {
		It("Should consume one message correctly and retrieve it", func() {
			app := "consumerApp1"
			service := "gcm"

			// config.Set("workers.consumer.topicTemplate", "%s-%s")
			brokers := config.GetStringSlice("workers.consumer.brokers")
			topicTemplate := config.GetString("workers.consumer.topicTemplate")
			topic := fmt.Sprintf(topicTemplate, app, service)

			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)
			go consumer.Consumer(l, config, app, service, outChan, doneChan)

			// Produce Messages
			message := "message"
			produceMessage(brokers, topic, message, int32(0), int64(0))

			consumedMessage := <-outChan
			Expect(consumedMessage).To(Equal(message))
		})

		It("Should consume two messages correctly and retrieve them", func() {
			app := "consumerApp2"
			service := "gcm"
			message := "message%d"

			// config.Set("workers.consumer.topicTemplate", "%s-%s")
			brokers := config.GetStringSlice("workers.consumer.brokers")
			topicTemplate := config.GetString("workers.consumer.topicTemplate")
			topic := fmt.Sprintf(topicTemplate, app, service)

			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)
			go consumer.Consumer(l, config, app, service, outChan, doneChan)

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
			app := "consumerApp3"
			service := "gcm"
			message := "message"

			// config.Set("workers.consumer.topicTemplate", "%s-%s")
			brokers := config.GetStringSlice("workers.consumer.brokers")
			topicTemplate := config.GetString("workers.consumer.topicTemplate")
			topic := fmt.Sprintf(topicTemplate, app, service)

			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)
			go consumer.Consumer(l, config, app, service, outChan, doneChan)

			produceMessage(brokers, topic, "", int32(0), int64(0))
			produceMessage(brokers, topic, "message", int32(0), int64(1))

			// The channel receives only non-empty messages
			consumedMessage := <-outChan
			Expect(consumedMessage).To(Equal(message))
		})

		It("Should not start a consumer if no broker found", func() {
			app := "consumerApp4"
			service := "gcm"
			brokers := []string{"localhost:0666"}
			config.Set("workers.consumer.brokers", brokers)
			// config.Set("workers.consumer.topicTemplate", "%s-%s")

			outChan := make(chan string, 10)
			doneChan := make(chan struct{}, 1)
			defer close(doneChan)

			// Consumer returns here and don't get blocked
			consumer.Consumer(l, config, app, service, outChan, doneChan)
		})
	})
})
