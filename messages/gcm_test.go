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

package messages_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/messages"
)

var _ = Describe("GCM Message", func() {
	Describe("Creating new message", func() {
		It("should return message", func() {
			data := map[string]interface{}{"x": 1}
			msg := messages.NewGCMMessage("to", data, nil, nil, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			fullData := map[string]interface{}{"x": 1, "templateName": "my-template"}
			Expect(msg.Data).To(BeEquivalentTo(fullData))
			Expect(msg.TimeToLive).To(BeEquivalentTo(357))
			Expect(msg.DryRun).To(Equal(false))
			Expect(msg.DeliveryReceiptRequest).To(Equal(false))
			Expect(msg.DelayWhileIdle).To(Equal(false))
			Expect(msg.MessageID).To(Equal(""))
		})

		It("should return message if data is nil", func() {
			msg := messages.NewGCMMessage("to", nil, nil, nil, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.Data).To(BeEquivalentTo(map[string]interface{}{"templateName": "my-template"}))
			Expect(msg.TimeToLive).To(BeEquivalentTo(357))
			Expect(msg.DryRun).To(Equal(false))
			Expect(msg.DeliveryReceiptRequest).To(Equal(false))
			Expect(msg.DelayWhileIdle).To(Equal(false))
			Expect(msg.MessageID).To(Equal(""))
		})

		It("should return message if metadata is not nil", func() {
			mtd := map[string]interface{}{"a": 1, "b": 2}
			msg := messages.NewGCMMessage("to", nil, mtd, nil, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.Data["m"].(map[string]interface{})).To(BeEquivalentTo(mtd))
			Expect(msg.TimeToLive).To(BeEquivalentTo(357))
			Expect(msg.DryRun).To(Equal(false))
			Expect(msg.DeliveryReceiptRequest).To(Equal(false))
			Expect(msg.DelayWhileIdle).To(Equal(false))
			Expect(msg.MessageID).To(Equal(""))
		})

		It("should return message if pushMetadata is not nil", func() {
			mtd := map[string]interface{}{"a": 1, "b": 2}
			msg := messages.NewGCMMessage("to", nil, nil, mtd, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.Metadata).To(BeEquivalentTo(mtd))
			Expect(msg.TimeToLive).To(BeEquivalentTo(357))
			Expect(msg.DryRun).To(Equal(false))
			Expect(msg.DeliveryReceiptRequest).To(Equal(false))
			Expect(msg.DelayWhileIdle).To(Equal(false))
			Expect(msg.MessageID).To(Equal(""))
		})

		It("should return message if pushMetadata is nil", func() {
			msg := messages.NewGCMMessage("to", nil, nil, nil, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.TimeToLive).To(BeEquivalentTo(357))
			Expect(msg.DryRun).To(Equal(false))
			Expect(msg.DeliveryReceiptRequest).To(Equal(false))
			Expect(msg.DelayWhileIdle).To(Equal(false))
			Expect(msg.MessageID).To(Equal(""))
		})

		It("should contain ttl in json message if ttl is greater than 0", func() {
			msg := messages.NewGCMMessage("to", nil, nil, nil, 357, "my-template")
			Expect(msg).NotTo(BeNil())
			msgStr, err := msg.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(msgStr).To(ContainSubstring("time_to_live"))
		})

		It("should not contain ttl in json message if ttl is equal to 0", func() {
			msg := messages.NewGCMMessage("to", nil, nil, nil, 0, "my-template")
			Expect(msg).NotTo(BeNil())
			msgStr, err := msg.ToJSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(msgStr).NotTo(ContainSubstring("time_to_live"))
		})
	})
})
