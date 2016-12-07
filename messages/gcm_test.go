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
			msg := messages.NewGCMMessage("to", data, 357)
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.Data).To(BeEquivalentTo(data))
			Expect(msg.PushExpiry).To(BeEquivalentTo(357))
		})

		It("should return message if data is nil", func() {
			msg := messages.NewGCMMessage("to", nil, 357)
			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("to"))
			Expect(msg.Data).To(BeEquivalentTo(map[string]interface{}{}))
			Expect(msg.PushExpiry).To(BeEquivalentTo(357))
		})
	})
})
