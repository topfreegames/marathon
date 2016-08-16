package templates_test

import (
	"encoding/json"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/templates"

	mt "git.topfreegames.com/topfreegames/marathon/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Template", func() {
	var (
		l zap.Logger
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
	})
	Describe("Parser", func() {
		Describe("Parse", func() {
			It("Should parse an input message correctly", func() {
				input := &messages.InputMessage{
					App:        "colorfy",
					Token:      "token",
					Service:    "gcm",
					PushExpiry: int64(0),
					Locale:     "en",
					Template:   "test_template1",
					Params:     map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					Message:    map[string]interface{}{},
					Metadata:   map[string]interface{}{"meta": "data"},
				}
				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				output, err := templates.Parse(l, string(inputString), true)
				Expect(err).NotTo(HaveOccurred())
				Expect(input.App).To(Equal(output.App))
				Expect(input.Token).To(Equal(output.Token))
				Expect(input.Service).To(Equal(output.Service))
				Expect(input.PushExpiry).To(Equal(output.PushExpiry))
				Expect(input.Locale).To(Equal(output.Locale))
				Expect(input.Template).To(Equal(output.Template))
				Expect(input.Params).To(Equal(output.Params))
				Expect(input.Message).To(Equal(output.Message))
				Expect(input.Metadata).To(Equal(output.Metadata))
			})

			It("Should parse a plain text message correctly", func() {
				input := &messages.InputMessage{
					App:        "colorfy",
					Token:      "token",
					Service:    "gcm",
					PushExpiry: int64(0),
					Locale:     "en",
					Template:   "",
					Params:     map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					Message:    map[string]interface{}{"aps": "hey"},
					Metadata:   map[string]interface{}{"meta": "data"},
				}
				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				output, err := templates.Parse(l, string(inputString), true)
				Expect(err).NotTo(HaveOccurred())
				Expect(input.App).To(Equal(output.App))
				Expect(input.Token).To(Equal(output.Token))
				Expect(input.Service).To(Equal(output.Service))
				Expect(input.PushExpiry).To(Equal(output.PushExpiry))
				Expect(input.Locale).To(Equal(output.Locale))
				Expect(input.Template).To(Equal(output.Template))
				Expect(input.Params).To(Equal(output.Params))
				Expect(input.Message).To(Equal(output.Message))
				Expect(input.Metadata).To(Equal(output.Metadata))
			})

			It("Should not parse an invalid message (both message and template defined)", func() {
				input := &messages.InputMessage{
					App:        "colorfy",
					Token:      "token",
					Service:    "gcm",
					PushExpiry: int64(0),
					Locale:     "en",
					Template:   "template",
					Params:     map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					Message:    map[string]interface{}{"aps": "hey"},
					Metadata:   map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				_, err := templates.Parse(l, string(inputString), true)
				Expect(err).To(HaveOccurred())
			})

			It("Should not parse a message without template and message", func() {
				input := map[string]interface{}{
					"app":         "colorfy",
					"token":       "token",
					"service":     "gcm",
					"push_expiry": int64(0),
					"locale":      "en",
					"template":    "",
					"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					"message":     map[string]interface{}{},
					"metadata":    map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				_, err := templates.Parse(l, string(inputString), true)
				Expect(err).To(HaveOccurred())
			})

			It("Should not parse a message which is not a json", func() {
				input := map[string]interface{}{
					"app":         "colorfy",
					"token":       "token",
					"service":     "gcm",
					"push_expiry": int64(0),
					"locale":      "en",
					"template":    "template",
					"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					"message":     map[string]interface{}{},
					"metadata":    map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				message := string(inputString)
				_, err := templates.Parse(l, message[:len(message)-1], true)
				Expect(err).To(HaveOccurred())
			})

			It("Should not parse a message with no App", func() {
				input := map[string]interface{}{
					"token":       "token",
					"service":     "gcm",
					"push_expiry": int64(0),
					"locale":      "en",
					"template":    "template",
					"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					"message":     map[string]interface{}{},
					"metadata":    map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				_, err := templates.Parse(l, string(inputString), true)
				Expect(err).To(HaveOccurred())
			})

			It("Should not parse a message with no Token", func() {
				input := map[string]interface{}{
					"app":         "colorfy",
					"service":     "gcm",
					"push_expiry": int64(0),
					"locale":      "en",
					"Service":     "gcm",
					"template":    "template",
					"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					"message":     map[string]interface{}{},
					"metadata":    map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				_, err := templates.Parse(l, string(inputString), true)
				Expect(err).To(HaveOccurred())
			})

			It("Should not parse a message with no Service", func() {
				input := map[string]interface{}{
					"app":        "colorfy",
					"token":      "token",
					"pushExpiry": int64(0),
					"locale":     "en",
					"template":   "template",
					"params":     map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
					"message":    map[string]interface{}{},
					"metadata":   map[string]interface{}{"meta": "data"},
				}

				inputString, marshalErr := json.Marshal(input)
				Expect(marshalErr).To(BeNil())

				_, err := templates.Parse(l, string(inputString), true)
				Expect(err).To(HaveOccurred())
			})

			Describe("", func() {
				It("Should parse messages correctly", func() {
					// Both message and template
					input1 := map[string]interface{}{
						"app":         "colorfy1",
						"token":       "token",
						"push_expiry": int64(0),
						"locale":      "en",
						"service":     "gcm",
						"template":    "template",
						"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
						"message":     map[string]interface{}{"alert": "{{param1}}"},
						"metadata":    map[string]interface{}{"meta": "data"},
					}

					inputString1, marshalErr1 := json.Marshal(input1)
					Expect(marshalErr1).To(BeNil())
					message1 := string(inputString1)

					// No message
					input2 := map[string]interface{}{
						"app":         "colorfy2",
						"token":       "token",
						"push_expiry": int64(0),
						"locale":      "en",
						"service":     "gcm",
						"template":    "template",
						"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
						"metadata":    map[string]interface{}{"meta": "data"},
					}
					inputString2, marshalErr2 := json.Marshal(input2)
					Expect(marshalErr2).To(BeNil())
					message2 := string(inputString2)

					// No template
					input3 := map[string]interface{}{
						"app":         "colorfy3",
						"token":       "token",
						"push_expiry": int64(0),
						"locale":      "en",
						"service":     "gcm",
						"params":      map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
						"message":     map[string]interface{}{"alert": "{{param1}}"},
						"metadata":    map[string]interface{}{"meta": "data"},
					}
					inputString3, marshalErr3 := json.Marshal(input3)
					Expect(marshalErr3).To(BeNil())
					message3 := string(inputString3)

					// No message or template
					input4 := map[string]interface{}{
						"app":         "colorfy4",
						"token":       "token",
						"push_expiry": int64(0),
						"locale":      "en",
						"service":     "gcm",
					}
					inputString4, marshalErr4 := json.Marshal(input4)
					Expect(marshalErr4).To(BeNil())
					message4 := string(inputString4)

					inChan := make(chan string, 10)
					outChan := make(chan *messages.InputMessage, 10)
					doneChan := make(chan struct{}, 1)
					defer close(doneChan)

					requireToken := true
					go templates.Parser(l, requireToken, inChan, outChan, doneChan)

					inChan <- message1
					inChan <- message2
					inChan <- message3
					inChan <- message4

					output1 := <-outChan
					output2 := <-outChan

					Expect(output1.App).To(Equal(input2["app"]))
					Expect(output1.Token).To(Equal(input2["token"]))
					Expect(output1.Service).To(Equal(input2["service"]))
					Expect(output1.PushExpiry).To(Equal(input2["push_expiry"]))
					Expect(output1.Locale).To(Equal(input2["locale"]))
					Expect(output1.Template).To(Equal(input2["template"]))
					Expect(output1.Params).To(Equal(input2["params"]))
					Expect(output1.Message).To(Equal(map[string]interface{}{}))
					Expect(output1.Metadata).To(Equal(input2["metadata"]))

					Expect(output2.App).To(Equal(input3["app"]))
					Expect(output2.Token).To(Equal(input3["token"]))
					Expect(output2.Service).To(Equal(input3["service"]))
					Expect(output2.PushExpiry).To(Equal(input3["push_expiry"]))
					Expect(output2.Locale).To(Equal(input3["locale"]))
					Expect(output2.Template).To(Equal(""))
					Expect(output2.Params).To(Equal(input3["params"]))
					Expect(output2.Message).To(Equal(input3["message"]))
					Expect(output2.Metadata).To(Equal(input3["metadata"]))
				})
			})
		})
	})
})
