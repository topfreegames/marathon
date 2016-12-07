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

package extensions_test

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

func getNextMessageFrom(kafkaBrokers []string, topic string, partition int32, offset int64) (*sarama.ConsumerMessage, error) {
	consumer, err := sarama.NewConsumer(kafkaBrokers, nil)
	Expect(err).NotTo(HaveOccurred())
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	Expect(err).NotTo(HaveOccurred())
	defer partitionConsumer.Close()

	msg := <-partitionConsumer.Messages()
	return msg, nil
}

var _ = Describe("Kafka Extension", func() {
	var logger zap.Logger
	var config *viper.Viper

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		config = viper.New()
	})

	Describe("Creating new client", func() {
		It("should return connected client", func() {
			client := getConnectedZookeeper(logger)
			kafka, err := extensions.NewKafkaClient(client, config, logger)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			Expect(kafka.KafkaBrokers).To(HaveLen(1))
			Expect(kafka.Producer).NotTo(BeNil())
		})
	})

	Describe("Send GCM Message", func() {
		It("should send GCM message", func() {
			client := getConnectedZookeeper(logger)
			kafka, err := extensions.NewKafkaClient(client, client.Config, logger)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			payload := map[string]interface{}{"x": 1}
			meta := map[string]interface{}{"a": 1}
			expiry := time.Now().Unix()
			partition, offset, err := kafka.SendGCMPush("consumergcm", "device-token", payload, meta, expiry)
			Expect(err).NotTo(HaveOccurred())

			msg, err := getNextMessageFrom(kafka.KafkaBrokers, "consumergcm", partition, offset)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).NotTo(BeNil())

			var gcmMessage messages.GCMMessage
			err = json.Unmarshal(msg.Value, &gcmMessage)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcmMessage.To).To(Equal("device-token"))
			Expect(gcmMessage.PushExpiry).To(BeEquivalentTo(expiry))
			Expect(gcmMessage.Data["x"]).To(BeEquivalentTo(1))
			Expect(gcmMessage.Data["m"].(map[string]interface{})["a"]).To(BeEquivalentTo(1))
		})
	})

	Describe("Send APNS Message", func() {
		It("should send APNS message", func() {
			client := getConnectedZookeeper(logger)
			kafka, err := extensions.NewKafkaClient(client, client.Config, logger)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			payload := map[string]interface{}{"x": 1}
			meta := map[string]interface{}{"a": 1}
			expiry := time.Now().Unix()
			partition, offset, err := kafka.SendAPNSPush("consumerapns", "device-token", payload, meta, expiry)
			Expect(err).NotTo(HaveOccurred())

			msg, err := getNextMessageFrom(kafka.KafkaBrokers, "consumerapns", partition, offset)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).NotTo(BeNil())

			var apnsMessage messages.APNSMessage
			err = json.Unmarshal(msg.Value, &apnsMessage)
			Expect(err).NotTo(HaveOccurred())
			Expect(apnsMessage.DeviceToken).To(Equal("device-token"))
			Expect(apnsMessage.PushExpiry).To(BeEquivalentTo(expiry))
			Expect(apnsMessage.Payload.Aps["x"]).To(BeEquivalentTo(1))
			Expect(apnsMessage.Payload.M["a"]).To(BeEquivalentTo(1))
		})
	})
})
