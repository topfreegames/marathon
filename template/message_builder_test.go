package template_test

import (
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Describe("MessageBuilder", func() {
		Describe("MessageBuilder.replaceTemplate", func() {
			It("Should replace variables", func() {
				params := map[string]string{
					"param1":  "hello",
					"someone": "banduk",
				}

				message := "{{param1}}, {{someone}} is awesome"
				resp, rplTplErr := template.ReplaceTemplate(message, &params)
				Expect(rplTplErr).To(BeNil())
				Expect(resp).To(Equal("hello, banduk is awesome"))
			})

			It("Should replace multiple variables", func() {
				params := map[string]string{
					"param1": "hello",
				}
				message := "{{param1}}, {{param1}}, {{param1}}"
				resp, rplTplErr := template.ReplaceTemplate(message, &params)
				Expect(rplTplErr).To(BeNil())
				Expect(resp).To(Equal("hello, hello, hello"))
			})

			It("Should not replace variables if template has syntax error", func() {
				params := map[string]string{
					"param1": "hello",
				}
				message := "{{param1}"
				resp, rplTplErr := template.ReplaceTemplate(message, &params)
				Expect(rplTplErr).NotTo(BeNil())
				Expect(resp).To(Equal(""))
			})
		})

		Describe("MessageBuilder.buildApnsMsg", func() {
			It("Should build message correctly with metadata", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
				}
				msg, err := template.BuildApnsMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))
			})

			It("Should build message correctly with no metadata", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := template.BuildApnsMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1}},"push_expiry":0}`))
			})

			It("Should not build message with invalid json", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := template.BuildApnsMsg(req, `{"alert": "message", "badge": 1`)
				Expect(err).NotTo(BeNil())
				Expect(msg).To(Equal(""))
			})
		})

		Describe("MessageBuilder.buildGcmMsg", func() {
			It("Should build message correctly with metadata", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
				}
				msg, err := template.BuildGcmMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"to":"token","data":{"alert":"message","badge":1,"m":{"meta":"data"}},"push_expiry":0}`))
			})

			It("Should build message correctly with no metadata", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := template.BuildGcmMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"to":"token","data":{"alert":"message","badge":1},"push_expiry":0}`))
			})

			It("Should not build message with invalid json", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := template.BuildGcmMsg(req, `{"alert": "message", "badge": 1`)
				Expect(err).NotTo(BeNil())
				Expect(msg).To(Equal(""))
			})
		})

		Describe("MessageBuilder.BuildMessage", func() {
			It("Should build apns message", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "apns",
					Message:    `{"alert": "{{param1}}, {{param2}}", "badge": 1}`,
					Params:     map[string]string{"param1": "Hello", "param2": "world"},
				}
				msg, err := template.BuildMessage(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello, world","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))

			})

			It("Should build apns message with no params", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "apns",
					Message:    `{"alert": "Hello", "badge": 1}`,
				}
				msg, err := template.BuildMessage(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))

			})

			It("Should build gcm message", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "gcm",
					Message:    `{"title": "{{param1}}, {{param2}}", "subtitle": "{{param1}}"}`,
					Params:     map[string]string{"param1": "Hello", "param2": "world"},
				}
				msg, err := template.BuildMessage(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`))
			})

			It("Should build gcm message with no params", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "gcm",
					Message:    `{"title": "Hello, world", "subtitle": "Hello"}`,
				}
				msg, err := template.BuildMessage(req)
				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`))
			})

			It("Should not build message of unknown type", func() {
				req := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "unknown",
					Message:    `{"alert": "{{param1}}, {{param2}}", "badge": 1}`,
					Params:     map[string]string{"param1": "Hello", "param2": "world"},
				}
				msg, err := template.BuildMessage(req)

				Expect(err).NotTo(BeNil())
				Expect(msg).To(BeNil())
			})
		})

		Describe("MessageBuilder.replaceTemplate", func() {
			It("Should replace variables", func() {
				reqGcm := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "gcm",
					Message:    `{"title": "Hello, world", "subtitle": "Hello"}`,
				}
				reqApns := &messages.RequestMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]string{"meta": "data"},
					App:        "colorfy",
					Type:       "apns",
					Message:    `{"alert": "{{param1}}, {{param2}}", "badge": 1}`,
					Params:     map[string]string{"param1": "Hello", "param2": "world"},
				}

				inChan := make(chan *messages.RequestMessage, 1)
				outChan := make(chan *messages.KafkaMessage, 1)
				go template.MessageBuilder(inChan, outChan)

				go func() {
					inChan <- reqGcm
					inChan <- reqApns
				}()

				out1 := <-outChan
				out2 := <-outChan

				Expect(out1.Message).To(Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`))
				Expect(out1.Topic == "colorfy")

				Expect(out2.Message).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello, world","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))
				Expect(out2.Topic).To(Equal("colorfy"))
			})
		})
	})
})
