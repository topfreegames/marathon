package template

import (
	"testing"

	"git.topfreegames.com/topfreegames/marathon/messages"

	. "github.com/franela/goblin"
)

func TestMessageBuilder(t *testing.T) {
	g := Goblin(t)

	g.Describe("MessageBuilder.replaceTemplate", func() {
		g.It("Should replace variables", func() {
			params := map[string]string{
				"param1":  "hello",
				"someone": "banduk",
			}

			message := "{{param1}}, {{someone}} is awesome"
			resp := replaceTemplate(message, &params)

			g.Assert(resp).Equal("hello, banduk is awesome")
		})

		g.It("Should replace multiple variables", func() {
			params := map[string]string{
				"param1":  "hello",
				"someone": "banduk",
			}
			message := "{{param1}}, {{param1}}, {{param1}}"
			resp := replaceTemplate(message, &params)
			g.Assert(resp).Equal("hello, hello, hello")
		})
	})

	g.Describe("MessageBuilder.buildApnsMsg", func() {
		g.It("Should build message correctly with metadata", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
			}
			msg := buildApnsMsg(req, `{"alert": "message", "badge": 1}`)

			g.Assert(msg).Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1},"m":{"meta":"data"}},"push_expiry":0}`)
		})

		g.It("Should build message correctly with no metadata", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
			}
			msg := buildApnsMsg(req, `{"alert": "message", "badge": 1}`)
			g.Assert(msg).Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"message","badge":1}},"push_expiry":0}`)
		})
	})

	g.Describe("MessageBuilder.buildGcmMsg", func() {
		g.It("Should build message correctly with metadata", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
			}
			msg := buildGcmMsg(req, `{"alert": "message", "badge": 1}`)
			g.Assert(msg).Equal(`{"to":"token","data":{"alert":"message","badge":1,"m":{"meta":"data"}},"push_expiry":0}`)
		})

		g.It("Should build message correctly with no metadata", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
			}
			msg := buildGcmMsg(req, `{"alert": "message", "badge": 1}`)
			g.Assert(msg).Equal(`{"to":"token","data":{"alert":"message","badge":1},"push_expiry":0}`)
		})
	})

	g.Describe("MessageBuilder.BuildMessage", func() {
		g.It("Should build apns message", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
				App:        "colorfy",
				Type:       "apns",
				Message:    `{"alert": "{{param1}}, {{param2}}", "badge": 1}`,
				Params:     map[string]string{"param1": "Hello", "param2": "world"},
			}
			msg, err := BuildMessage(req)

			g.Assert(err).Equal(nil)
			g.Assert(msg.Message).Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello, world","badge":1},"m":{"meta":"data"}},"push_expiry":0}`)

		})

		g.It("Should build gcm message", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
				App:        "colorfy",
				Type:       "gcm",
				Message:    `{"title": "{{param1}}, {{param2}}", "subtitle": "{{param1}}"}`,
				Params:     map[string]string{"param1": "Hello", "param2": "world"},
			}
			msg, err := BuildMessage(req)

			g.Assert(err).Equal(nil)
			g.Assert(msg.Message).Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`)
		})

		g.It("Should build gcm message with no params", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
				App:        "colorfy",
				Type:       "gcm",
				Message:    `{"title": "Hello, world", "subtitle": "Hello"}`,
			}
			msg, err := BuildMessage(req)
			g.Assert(err).Equal(nil)
			g.Assert(msg.Message).Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`)
		})

		g.It("Should not build message of unknown type", func() {
			req := &messages.RequestMessage{
				Token:      "token",
				PushExpiry: 0,
				Metadata:   map[string]string{"meta": "data"},
				App:        "colorfy",
				Type:       "unknown",
				Message:    `{"alert": "{{param1}}, {{param2}}", "badge": 1}`,
				Params:     map[string]string{"param1": "Hello", "param2": "world"},
			}
			msg, err := BuildMessage(req)

			g.Assert(err == nil).IsFalse()
			g.Assert(msg == nil).IsTrue()
		})
	})

	g.Describe("MessageBuilder.replaceTemplate", func() {
		g.It("Should replace variables", func() {
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
			go MessageBuilder(inChan, outChan)

			go func() {
				inChan <- reqGcm
				inChan <- reqApns
			}()

			out1 := <-outChan
			out2 := <-outChan

			g.Assert(out1.Message).Equal(`{"to":"token","data":{"m":{"meta":"data"},"subtitle":"Hello","title":"Hello, world"},"push_expiry":0}`)
			g.Assert(out1.Topic == "colorfy")

			g.Assert(out2.Message).Equal(`{"DeviceToken":"token","Payload":{"aps":{"alert":"Hello, world","badge":1},"m":{"meta":"data"}},"push_expiry":0}`)
			g.Assert(out2.Topic).Equal("colorfy")
		})
	})
}
