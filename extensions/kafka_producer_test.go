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

	"github.com/DataDog/datadog-go/statsd"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
)

func getNextMessageFrom(consumer *kafka.Consumer) (*kafka.Message, error) {
	for ev := range consumer.Events() {
		if ev == nil {
			continue
		}
		switch ev.(type) {
		case *kafka.Message:
			return ev.(*kafka.Message), nil
		case kafka.PartitionEOF:
			continue
		case kafka.Error:
			return nil, ev.(error)
		default:
			continue
		}
	}
	return nil, nil
}

func waitForConsumer(consumer *kafka.Consumer) error {
	for ev := range consumer.Events() {
		if ev == nil {
			continue
		}
		switch ev.(type) {
		case kafka.PartitionEOF:
			return nil
		case kafka.Error:
			return ev.(error)
		default:
			continue
		}
	}
	return nil
}

var _ = Describe("Kafka Extension", func() {
	var logger zap.Logger
	var config *viper.Viper
	var testConsumer *kafka.Consumer
	var statsdClient *statsd.Client

	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
		config = viper.New()
		var err error
		testConsumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":        "localhost:9940",
			"group.id":                 uuid.NewV4().String(),
			"go.events.channel.enable": true,
		})
		Expect(err).NotTo(HaveOccurred())
		err = testConsumer.Subscribe("consumer", nil)
		Expect(err).NotTo(HaveOccurred())
		err = waitForConsumer(testConsumer)
		Expect(err).NotTo(HaveOccurred())

		statsdClient, err = statsd.New("localhost:1234")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := testConsumer.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Creating new client", func() {
		It("should return connected client", func() {
			kafka, err := extensions.NewKafkaProducer(config, logger, statsdClient)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			Expect(kafka.BootstrapBrokers).NotTo(BeNil())
			Expect(kafka.Producer).NotTo(BeNil())
		})
	})

	Describe("Send GCM Message", func() {
		It("should send GCM message", func() {
			kafka, err := extensions.NewKafkaProducer(config, logger, statsdClient)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			payload := map[string]interface{}{"x": 1}
			meta := map[string]interface{}{"a": 1}
			expiry := time.Now().Unix()
			kafka.SendGCMPush("consumer", "device-token", payload, meta, nil, expiry, "template")
			msg, err := getNextMessageFrom(testConsumer)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).NotTo(BeNil())

			var gcmMessage messages.GCMMessage
			err = json.Unmarshal(msg.Value, &gcmMessage)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcmMessage.To).To(Equal("device-token"))
			Expect(gcmMessage.TimeToLive).To(BeEquivalentTo(expiry))
			Expect(gcmMessage.Data["x"]).To(BeEquivalentTo(1))
			Expect(gcmMessage.Data["m"].(map[string]interface{})["a"]).To(BeEquivalentTo(1))
		})
	})

	Describe("Send APNS Message", func() {
		It("should send APNS message", func() {
			kafka, err := extensions.NewKafkaProducer(config, logger, statsdClient)
			Expect(err).NotTo(HaveOccurred())
			defer kafka.Close()

			payload := map[string]interface{}{"x": 1}
			meta := map[string]interface{}{"a": 1}
			expiry := time.Now().Unix()
			kafka.SendAPNSPush("consumer", "device-token", payload, meta, nil, expiry, "template")

			msg, err := getNextMessageFrom(testConsumer)
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
