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

	"github.com/Shopify/sarama"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

// KafkaClient is the struct that connects to Kafka
type KafkaClient struct {
	ZookeeperClient *ZookeeperClient
	Config          *viper.Viper
	ConfigPath      string
	Conn            *zk.Conn
	Logger          zap.Logger
	KafkaBrokers    []string
	Producer        sarama.SyncProducer // TODO: should we use asynProducer?
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

	err := client.LoadKafkaBrokers()
	if err != nil {
		return nil, err
	}

	err = client.ConnectToKafka()
	if err != nil {
		return nil, err
	}

	return client, nil
}

//LoadKafkaBrokers from Zookeeper
func (c *KafkaClient) LoadKafkaBrokers() error {
	if c.ZookeeperClient == nil || !c.ZookeeperClient.IsConnected() {
		return fmt.Errorf("Failed to start kafka extension due to zookeeper not being connected.")
	}

	brokers, err := c.ZookeeperClient.GetKafkaBrokers()
	if err != nil {
		return err
	}
	c.KafkaBrokers = brokers
	return nil
}

//ConnectToKafka connects with the Kafka from the broker
func (c *KafkaClient) ConnectToKafka() error {
	saramaConfig := sarama.NewConfig()

	producer, err := sarama.NewSyncProducer(c.KafkaBrokers, saramaConfig)
	if err != nil {
		return err
	}

	c.Producer = producer
	return nil
}

//Close the connections to kafka and zookeeper
func (c *KafkaClient) Close() error {
	err := c.Producer.Close()
	if err != nil {
		return err
	}

	err = c.ZookeeperClient.Close()
	if err != nil {
		return err
	}
	return nil
}

//SendAPNSPush notification to Kafka
func (c *KafkaClient) SendAPNSPush(topic, deviceToken string, payload, metadata map[string]interface{}, pushExpiry int64) (int32, int64, error) {
	msg := messages.NewAPNSMessage(
		deviceToken,
		pushExpiry,
		payload,
		metadata,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return 0, 0, err
	}

	return c.SendPush(messages.NewKafkaMessage(topic, message))
}

//SendGCMPush notification to Kafka
func (c *KafkaClient) SendGCMPush(topic, deviceToken string, payload, metadata map[string]interface{}, pushExpiry int64) (int32, int64, error) {
	msg := messages.NewGCMMessage(
		deviceToken,
		payload,
		metadata,
		pushExpiry,
	)

	message, err := msg.ToJSON()
	if err != nil {
		return 0, 0, err
	}

	return c.SendPush(messages.NewKafkaMessage(topic, message))
}

//SendPush notification to Kafka
func (c *KafkaClient) SendPush(msg *messages.KafkaMessage) (int32, int64, error) {
	saramaMessage := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.StringEncoder(msg.Message),
	}

	partition, offset, err := c.Producer.SendMessage(saramaMessage) // TODO: use SendMessages instead?
	if err != nil {
		log.E(c.Logger, "Error sending message", func(cm log.CM) {
			cm.Write(zap.Object("KafkaMessage", saramaMessage), zap.Error(err))
		})
		return 0, 0, err
	}
	log.I(c.Logger, "Sent message", func(cm log.CM) {
		cm.Write(
			zap.Object("KafkaMessage", saramaMessage),
			zap.String("topic", msg.Topic),
			zap.Int("partition", int(partition)),
			zap.Int64("offset", offset),
		)
	})
	return partition, offset, nil
}
