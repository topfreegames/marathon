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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/topfreegames/marathon/extensions"
	"github.com/uber-go/zap"
)

func getConnectedClient() *extensions.ZookeeperClient {
	client, err := extensions.NewZookeeperClient("../config/test.yaml", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(client).NotTo(BeNil())

	time.Sleep(10 * time.Millisecond)
	Expect(client.Conn).NotTo(BeNil())
	Expect(client.Conn.State()).To(Equal(zk.StateConnected))

	return client
}

var _ = Describe("Zookeeper Extension", func() {
	var logger zap.Logger
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
	})

	Describe("Creating new client", func() {
		It("should return connected client", func() {
			client, err := extensions.NewZookeeperClient("../config/test.yaml", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

			time.Sleep(10 * time.Millisecond)
			Expect(client.Conn).NotTo(BeNil())
			Expect(client.Conn.State()).To(Equal(zk.StateConnected))
		})
	})

	Describe("Getting Kafka Brokers", func() {
		It("should return kafka brokers", func() {
			client := getConnectedClient()

			brokerInfo, err := client.GetKafkaBrokers()
			Expect(err).NotTo(HaveOccurred())
			Expect(brokerInfo).To(HaveLen(1))
			Expect(brokerInfo[0]).To(ContainSubstring(":9940"))
		})
	})
})
