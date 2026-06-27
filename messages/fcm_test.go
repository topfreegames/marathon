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
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/messages"
)

var _ = Describe("FCM Message", func() {
	Describe("Creating new message", func() {
		It("maps a string aps alert to notification.body (the visible banner)", func() {
			payload := map[string]interface{}{"alert": "Everyone just liked your village!"}
			msg := messages.NewFCMMessage("fcm-token", payload, nil, nil, 0, "my-template")

			Expect(msg).NotTo(BeNil())
			Expect(msg.To).To(Equal("fcm-token"))
			Expect(msg.Notification).NotTo(BeNil())
			Expect(msg.Notification.Body).To(Equal("Everyone just liked your village!"))
			Expect(msg.Notification.Title).To(Equal(""))
			Expect(msg.DryRun).To(Equal(false))
		})

		It("maps a dictionary aps alert to title/body and loc keys", func() {
			payload := map[string]interface{}{
				"alert": map[string]interface{}{
					"title":         "Village",
					"body":          "Someone liked it",
					"title-loc-key": "TITLE_KEY",
					"loc-key":       "BODY_KEY",
					"loc-args":      []interface{}{"Everyone", "village"},
				},
			}
			msg := messages.NewFCMMessage("fcm-token", payload, nil, nil, 0, "tpl")

			Expect(msg.Notification).NotTo(BeNil())
			Expect(msg.Notification.Title).To(Equal("Village"))
			Expect(msg.Notification.Body).To(Equal("Someone liked it"))
			Expect(msg.Notification.TitleLocKey).To(Equal("TITLE_KEY"))
			Expect(msg.Notification.BodyLocKey).To(Equal("BODY_KEY"))
			Expect(msg.Notification.BodyLocArgs).To(Equal("Everyone,village"))
		})

		It("maps sound, badge and content-available", func() {
			payload := map[string]interface{}{
				"alert":             "hi",
				"sound":             "chime.caf",
				"badge":             float64(7),
				"content-available": float64(1),
			}
			msg := messages.NewFCMMessage("fcm-token", payload, nil, nil, 0, "tpl")

			Expect(msg.Notification.Sound).To(Equal("chime.caf"))
			Expect(msg.Notification.Badge).To(Equal("7"))
			Expect(msg.ContentAvailable).To(BeTrue())
		})

		It("reads the name from a critical-alert sound dictionary", func() {
			payload := map[string]interface{}{
				"alert": "hi",
				"sound": map[string]interface{}{"critical": float64(1), "name": "siren.caf", "volume": 1.0},
			}
			msg := messages.NewFCMMessage("fcm-token", payload, nil, nil, 0, "tpl")
			Expect(msg.Notification.Sound).To(Equal("siren.caf"))
		})

		It("carries templateName and message metadata in data (parity with gcm)", func() {
			mtd := map[string]interface{}{"deeplink": "village/123"}
			msg := messages.NewFCMMessage("fcm-token", map[string]interface{}{"alert": "hi"}, mtd, nil, 0, "my-template")

			Expect(msg.Data["templateName"]).To(Equal("my-template"))
			Expect(msg.Data["m"]).To(BeEquivalentTo(mtd))
		})

		It("passes pushMetadata through to the top-level metadata field", func() {
			pmtd := map[string]interface{}{"userId": "u1", "pushType": "massive"}
			msg := messages.NewFCMMessage("fcm-token", map[string]interface{}{"alert": "hi"}, nil, pmtd, 0, "tpl")
			Expect(msg.Metadata).To(BeEquivalentTo(pmtd))
		})

		It("returns a nil notification when the payload has no alert or sound/badge", func() {
			msg := messages.NewFCMMessage("fcm-token", map[string]interface{}{"content-available": float64(1)}, nil, nil, 0, "tpl")
			Expect(msg.Notification).To(BeNil())
			Expect(msg.ContentAvailable).To(BeTrue())
		})

		It("tolerates a nil payload", func() {
			msg := messages.NewFCMMessage("fcm-token", nil, nil, nil, 0, "tpl")
			Expect(msg).NotTo(BeNil())
			Expect(msg.Notification).To(BeNil())
		})

		Describe("push_expiry (the seconds->milliseconds seam)", func() {
			It("converts seconds to the milliseconds pusher's MakeTimestamp compares against", func() {
				// 1_700_000_000s -> 1_700_000_000_000ms. If left in seconds, pusher
				// would treat every iOS message as already expired and silently drop it.
				msg := messages.NewFCMMessage("fcm-token", map[string]interface{}{"alert": "hi"}, nil, nil, 1700000000, "tpl")
				Expect(msg.PushExpiry).To(Equal(int64(1700000000000)))
			})

			It("omits push_expiry when there is no expiry (zero)", func() {
				msg := messages.NewFCMMessage("fcm-token", map[string]interface{}{"alert": "hi"}, nil, nil, 0, "tpl")
				Expect(msg.PushExpiry).To(Equal(int64(0)))

				msgStr, err := msg.ToJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(msgStr).NotTo(ContainSubstring("push_expiry"))
			})
		})

		Describe("wire format", func() {
			It("serializes the alert as a TOP-LEVEL notification, not buried in data", func() {
				payload := map[string]interface{}{"alert": "Everyone just liked your village!"}
				msg := messages.NewFCMMessage("fcm-token", payload, nil, nil, 0, "tpl")
				msgStr, err := msg.ToJSON()
				Expect(err).NotTo(HaveOccurred())

				// Decode generically and assert the shape pusher's buildIOSMessage reads.
				var decoded map[string]interface{}
				Expect(json.Unmarshal([]byte(msgStr), &decoded)).To(Succeed())

				notification, ok := decoded["notification"].(map[string]interface{})
				Expect(ok).To(BeTrue(), "notification must be a top-level object")
				Expect(notification["body"]).To(Equal("Everyone just liked your village!"))

				// The rendered alert must NOT live under data (that is the gcm shape
				// that yields a silent push on iOS).
				if data, ok := decoded["data"].(map[string]interface{}); ok {
					Expect(data).NotTo(HaveKey("alert"))
				}
			})
		})
	})
})
