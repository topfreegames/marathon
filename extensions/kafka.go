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
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

// KafkaClient is the struct that connects to Kafka
type KafkaClient struct {
	ZookeeperClient       *ZookeeperClient
	Config                *viper.Viper
	ConfigPath            string
	Conn                  *zk.Conn
	Logger                zap.Logger
	KafkaBootstrapBrokers string
	Producer              *kafka.Producer
}

// NewKafkaClient creates a new client
func NewKafkaClient(zookeeperClient *ZookeeperClient, config *viper.Viper, logger zap.Logger) (*KafkaClient, error) {
	l := logger.With(
		zap.String("source", "KafkaExtension"),
	)
	client := &KafkaClient{
		ZookeeperClient: zookeeperClient,
		Config:          config,
		Logger:          l,
	}

	err := client.loadKafkaBrokers()
	if err != nil {
		return nil, err
	}

	err = client.connectToKafka()
	if err != nil {
		return nil, err
	}

	return client, nil
}

//LoadKafkaBrokers from Zookeeper
func (c *KafkaClient) loadKafkaBrokers() error {
	if c.ZookeeperClient == nil || !c.ZookeeperClient.IsConnected() {
		return fmt.Errorf("Failed to start kafka extension due to zookeeper not being connected.")
	}

	brokers, err := c.ZookeeperClient.GetKafkaBrokers()
	if err != nil {
		return err
	}
	c.KafkaBootstrapBrokers = strings.Join(brokers, ",")
	return nil
}

//ConnectToKafka connects with the Kafka from the broker
func (c *KafkaClient) connectToKafka() error {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": c.KafkaBootstrapBrokers,
	}
	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return err
	}

	c.Producer = p
	return nil
}

//Close the connections to kafka and zookeeper
func (c *KafkaClient) Close() error {
	c.Producer.Close()

	err := c.ZookeeperClient.Close()
	if err != nil {
		return err
	}
	return nil
}

//SendAPNSPush notification to Kafka
func (c *KafkaClient) SendAPNSPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64) error {
	msg := messages.NewAPNSMessage(
		deviceToken,
		pushExpiry,
		payload,
		messageMetadata,
		pushMetadata,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}
	c.sendPush(messages.NewKafkaMessage(topic, message))
	return nil
}

//SendGCMPush notification to Kafka
func (c *KafkaClient) SendGCMPush(topic, deviceToken string, payload, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, pushExpiry int64) error {
	msg := messages.NewGCMMessage(
		deviceToken,
		payload,
		messageMetadata,
		pushMetadata,
		pushExpiry,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return err
	}
	c.sendPush(messages.NewKafkaMessage(topic, message))
	return nil
}

//SendPush notification to Kafka
func (c *KafkaClient) sendPush(msg *messages.KafkaMessage) {
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
