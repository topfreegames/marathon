/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package extensions

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	raven "github.com/getsentry/raven-go"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

// KafkaProducer is the struct that connects to Kafka
type KafkaProducer struct {
	Config           *viper.Viper
	ConfigPath       string
	Logger           zap.Logger
	BootstrapBrokers string
	BatchSize        int
	LingerMS         int
	Producer         *kafka.Producer
}

// NewKafkaProducer creates a new kafka producer
func NewKafkaProducer(config *viper.Viper, logger zap.Logger) (*KafkaProducer, error) {
	l := logger.With(
		zap.String("source", "KafkaExtension"),
	)
	client := &KafkaProducer{
		Config: config,
		Logger: l,
	}

	client.loadConfigurationDefaults()
	client.configure()

	err := client.connectToKafka()
	if err != nil {
		return nil, err
	}

	go client.listenForKafkaResponses()
	l.Info("configured kafka producer")
	return client, nil
}

func (c *KafkaProducer) loadConfigurationDefaults() {
	c.Config.SetDefault("kafka.bootstrapServers", "localhost:9940")
	c.Config.SetDefault("kafka.batch.size", 100)
	c.Config.SetDefault("kafka.linger.ms", 10)
}

func (c *KafkaProducer) configure() {
	c.BootstrapBrokers = c.Config.GetString("kafka.bootstrapServers")
	c.BatchSize = c.Config.GetInt("kafka.batch.size")
	c.LingerMS = c.Config.GetInt("kafka.linger.ms")
}

//ConnectToKafka connects with the Kafka from the broker
func (c *KafkaProducer) connectToKafka() error {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":          c.BootstrapBrokers,
		"queue.buffering.max.kbytes": c.BatchSize,
		"queue.buffering.max.ms":     c.LingerMS,
	}
	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return err
	}

	c.Producer = p
	return nil
}

func (c KafkaProducer) listenForKafkaResponses() {
	l := c.Logger.With(
		zap.String("source", "KafkaExtension"),
		zap.String("method", "listenForKafkaResponses"),
	)
	for e := range c.Producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				raven.CaptureError(m.TopicPartition.Error, map[string]string{
					"extension": "kafka-producer",
				})
				log.E(l, "Kafka producer error", func(cm log.CM) {
					cm.Write(zap.Error(m.TopicPartition.Error))
				})
			}
		}
	}
}

//Close the connections to kafka
func (c *KafkaProducer) Close() {
	c.Producer.Close()
}

//SendAPNSPush notification to Kafka
func (c *KafkaProducer) SendAPNSPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64, templateName string) error {
	msg := messages.NewAPNSMessage(
		deviceToken,
		pushExpiry,
		payload,
		messageMetadata,
		pushMetadata,
		templateName,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}
	c.sendPush(messages.NewKafkaMessage(topic, message))
	return nil
}

//SendGCMPush notification to Kafka
func (c *KafkaProducer) SendGCMPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64, templateName string) error {
	msg := messages.NewGCMMessage(
		deviceToken,
		payload,
		messageMetadata,
		pushMetadata,
		pushExpiry,
		templateName,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}
	c.sendPush(messages.NewKafkaMessage(topic, message))
	return nil
}

//SendPush notification to Kafka
func (c *KafkaProducer) sendPush(msg *messages.KafkaMessage) {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(msg.Message),
	}

	c.Producer.ProduceChannel() <- message
	log.D(c.Logger, "Sent message", func(cm log.CM) {
		cm.Write(
			zap.Object("KafkaMessage", message),
			zap.String("topic", msg.Topic),
		)
	})
}
