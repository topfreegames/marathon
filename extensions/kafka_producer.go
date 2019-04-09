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
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/Shopify/sarama"
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
	Producer         sarama.AsyncProducer
	Statsd           *statsd.Client
	MaxMessageBytes  int
	Retries          int
}

// NewKafkaProducer creates a new kafka producer
func NewKafkaProducer(config *viper.Viper, logger zap.Logger, statsd *statsd.Client) (*KafkaProducer, error) {
	l := logger.With(
		zap.String("source", "KafkaExtension"),
	)
	client := &KafkaProducer{
		Config: config,
		Logger: l,
		Statsd: statsd,
	}

	client.loadConfigurationDefaults()
	client.configure()

	client.connectToKafka()
	l.Info("configured kafka producer")
	return client, nil
}

func (c *KafkaProducer) loadConfigurationDefaults() {
	c.Config.SetDefault("kafka.bootstrapServers", "localhost:9940")
	c.Config.SetDefault("kafka.batch.size", 10)
	c.Config.SetDefault("kafka.linger.ms", 10)
	c.Config.SetDefault("kafka.maxMessageBytes", 1000000)
	c.Config.SetDefault("kafka.retries", 10)
}

func (c *KafkaProducer) configure() {
	c.BootstrapBrokers = c.Config.GetString("kafka.bootstrapServers")
	c.BatchSize = c.Config.GetInt("kafka.batch.size")
	c.LingerMS = c.Config.GetInt("kafka.linger.ms")
	c.MaxMessageBytes = c.Config.GetInt("kafka.maxMessageBytes")
	c.Retries = c.Config.GetInt("kafka.retries")
}

//ConnectToKafka connects with the Kafka from the broker
func (c *KafkaProducer) connectToKafka() error {
	config := sarama.NewConfig()
	config.Producer.Flush.Messages = c.BatchSize
	config.Producer.Flush.MaxMessages = c.BatchSize
	config.Producer.Flush.Frequency = time.Duration(c.LingerMS) * time.Millisecond

	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = c.Retries
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = c.MaxMessageBytes

	producer, err := sarama.NewAsyncProducer([]string{c.BootstrapBrokers}, config)
	if err != nil {
		return err
	}
	c.Producer = producer

	go func() {
		for range producer.Successes() {
			c.Statsd.Incr("send_message_return", []string{"error:false"}, 1)
		}
	}()

	go func() {
		for range producer.Errors() {
			c.Statsd.Incr("send_message_return", []string{"error:true"}, 1)
		}
	}()

	return nil
}

//Close the connections to kafka
func (c *KafkaProducer) Close() {
	c.Producer.AsyncClose()
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

	if val, ok := pushMetadata["dryRun"]; ok {
		if dryRun, _ := val.(bool); dryRun {
			msg.DeviceToken = GenerateFakeID(64)
		}
	}

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

	if val, ok := pushMetadata["dryRun"]; ok {
		if dryRun, _ := val.(bool); dryRun {
			msg.To = GenerateFakeID(152)
			msg.DryRun = true
		}
	}

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}
	c.sendPush(messages.NewKafkaMessage(topic, message))
	return nil
}

//SendPush notification to Kafka
func (c *KafkaProducer) sendPush(msg *messages.KafkaMessage) {
	message := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.StringEncoder(msg.Message),
	}
	c.Producer.Input() <- message
	log.D(c.Logger, "Sent message", func(cm log.CM) {
		cm.Write(
			zap.Object("KafkaMessage", message),
			zap.String("topic", msg.Topic),
		)
	})
}
