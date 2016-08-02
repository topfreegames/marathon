package templates_test

import (
	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/templates"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Describe("Builder", func() {
		Describe("Replace", func() {
			It("Should replace variables", func() {
				params := map[string]interface{}{
					"param1":  "hello",
					"someone": "banduk",
				}

				message := "{{param1}}, {{someone}} is awesome"

				resp, tplErr := templates.Replace(message, params)
				Expect(tplErr).To(BeNil())
				Expect(resp).To(Equal("hello, banduk is awesome"))
			})

			It("Should replace multiple variables", func() {
				params := map[string]interface{}{
					"param1": "hello",
				}
				message := "{{param1}}, {{param1}}, {{param1}}"
				resp, rplTplErr := templates.Replace(message, params)
				Expect(rplTplErr).To(BeNil())
				Expect(resp).To(Equal("hello, hello, hello"))
			})

			It("Should replace multiple nested variables", func() {
				params := map[string]interface{}{
					"param1": map[string]interface{}{
						"param1_1": "value1_1",
						"param1_2": "value1_2",
					},
					"param2": map[string]interface{}{
						"param2_1": "value2_1",
						"param2_2": "value2_2",
					},
				}
				message := "{{param1.param1_1}}, {{param1.param1_2}}, {{param2.param2_2}}"
				resp, rplTplErr := templates.Replace(message, params)
				Expect(rplTplErr).To(BeNil())
				Expect(resp).To(Equal("value1_1, value1_2, value2_2"))
			})

			It("Should not replace variables if template has syntax error", func() {
				params := map[string]interface{}{
					"param1": "hello",
				}
				message := "{{param1}"
				resp, rplTplErr := templates.Replace(message, params)
				Expect(rplTplErr).NotTo(BeNil())
				Expect(resp).To(Equal(""))
			})
		})

		Describe("ApnsMsg", func() {
			It("Should build message correctly with metadata", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
				}
				msg, err := templates.ApnsMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))
			})

			It("Should build message correctly with no metadata", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := templates.ApnsMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1}},"push_expiry":0}`))
			})

			It("Should not build message with invalid json", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := templates.ApnsMsg(req, `{"alert": "message", "badge": 1`)
				Expect(err).NotTo(BeNil())
				Expect(msg).To(Equal(""))
			})

			It("Should not build message correctly with incalid content", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				_, err := templates.ApnsMsg(req, `{"alert": "message", "badge": 1`)
				Expect(err).NotTo(BeNil())
			})
		})

		Describe("GcmMsg", func() {
			It("Should build message correctly with metadata", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
				}
				msg, err := templates.GcmMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"to":"token","data":{"alert":"message","badge":1,"m":{"meta":"data"}},"push_expiry":0}`))
			})

			It("Should build message correctly with no metadata", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				msg, err := templates.GcmMsg(req, `{"alert": "message", "badge": 1}`)
				Expect(err).To(BeNil())
				Expect(msg).To(Equal(`{"to":"token","data":{"alert":"message","badge":1},"push_expiry":0}`))
			})

			It("Should not build message correctly with incalid content", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
				}
				_, err := templates.GcmMsg(req, `{"alert": "message", "badge": 1`)
				Expect(err).NotTo(BeNil())
			})
		})

		Describe("Builder.Build", func() {
			It("Should build apns message", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "apns",
					Message:    map[string]interface{}{"alert": "{{param1}}, {{param2}}", "badge": 1},
					Params:     map[string]interface{}{"param1": "Hello", "param2": "world"},
				}
				msg, err := templates.Build(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello, world","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))

			})

			It("Should build apns message with no params", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "apns",
					Message:    map[string]interface{}{"alert": "Hello", "badge": 1},
				}
				msg, err := templates.Build(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello","badge":1},"m":{"meta":"data"}},"push_expiry":0}`))

			})

			It("Should build gcm message", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "gcm",
					Message:    map[string]interface{}{"title": "{{param1}}, {{param2}}", "subtitle": "{{param1}}"},
					Params:     map[string]interface{}{"param1": "Hello", "param2": "world"},
				}
				msg, err := templates.Build(req)

				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`))
			})

			It("Should build gcm message with no params", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "gcm",
					Message:    map[string]interface{}{"title": "Hello, world", "subtitle": "Hello"},
				}
				msg, err := templates.Build(req)
				Expect(err).To(BeNil())
				Expect(msg.Message).To(Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`))
			})

			It("Should not build message of unknown type", func() {
				req := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "unknown",
					Message:    map[string]interface{}{"alert": "{{param1}}, {{param2}}", "badge": 1},
					Params:     map[string]interface{}{"param1": "Hello", "param2": "world"},
				}
				msg, err := templates.Build(req)

				Expect(err).NotTo(BeNil())
				Expect(msg).To(BeNil())
			})
		})

		Describe("Builder.Replace", func() {
			It("Should replace variables", func() {
				reqGcm := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "gcm",
					Message:    map[string]interface{}{"title": "Hello, world", "subtitle": "Hello"},
				}
				reqApns := &messages.TemplatedMessage{
					Token:      "token",
					PushExpiry: 0,
					Metadata:   map[string]interface{}{"meta": "data"},
					App:        "colorfy",
					Service:    "apns",
					Message:    map[string]interface{}{"alert": "{{param1}}, {{param2}}", "badge": 1},
					Params:     map[string]interface{}{"param1": "Hello", "param2": "world"},
				}

				inChan := make(chan *messages.TemplatedMessage, 1)
				outChan := make(chan *messages.KafkaMessage, 1)
				doneChan := make(chan struct{}, 1)
				defer close(doneChan)
				go templates.Builder(inChan, outChan, doneChan)

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
