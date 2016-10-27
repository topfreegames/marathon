package templates_test

import (
	"time"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"
	"git.topfreegames.com/topfreegames/marathon/templates"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/imdario/mergo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Fetcher", func() {
	var (
		db *models.DB
		l  zap.Logger
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
		_db, dbErr := models.GetTestDB(l)
		Expect(dbErr).To(BeNil())
		Expect(_db).NotTo(BeNil())

		db = _db
	})

	Describe("FetchTemplate", func() {
		It("should fetch a template for a given name and default locale", func() {
			tc := templates.CreateTemplateCache(10)
			template := &models.Template{
				Name:     "test_template1",
				Locale:   "en",
				Defaults: map[string]interface{}{"param1": "value1", "param2": "value2"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "en",
				Template:   "test_template1",
				Params:     map[string]interface{}{"param1": "inputvalue1", "param3": "inputvalue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			fetched, errFetch := templates.FetchTemplate(l, input, db, tc)
			Expect(errFetch).To(BeNil())
			Expect(fetched.App).To(Equal(input.App))
			Expect(fetched.Token).To(Equal(input.Token))
			Expect(fetched.Service).To(Equal(input.Service))
			Expect(fetched.PushExpiry).To(Equal(input.PushExpiry))
			Expect(fetched.Locale).To(Equal(input.Locale))
			Expect(fetched.Metadata).To(Equal(input.Metadata))
			Expect(fetched.Message).To(Equal(template.Body))

			// Merge input and template params
			mergo.Merge(&input.Params, template.Defaults)
			Expect(fetched.Params).To(Equal(input.Params))
		})

		It("should fetch a template for a given and name non-default locale", func() {
			tc := templates.CreateTemplateCache(10)
			template := &models.Template{
				Name:     "test_template2",
				Locale:   "pt",
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "pt",
				Template:   "test_template2",
				Params:     map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			fetched, errFetch := templates.FetchTemplate(l, input, db, tc)
			Expect(errFetch).To(BeNil())
			Expect(fetched.App).To(Equal(input.App))
			Expect(fetched.Token).To(Equal(input.Token))
			Expect(fetched.Service).To(Equal(input.Service))
			Expect(fetched.PushExpiry).To(Equal(input.PushExpiry))
			Expect(fetched.Metadata).To(Equal(input.Metadata))
			Expect(fetched.Message).To(Equal(template.Body))

			// Not the default, but the requested
			Expect(fetched.Locale).To(Equal(input.Locale))

			// Merge input and template params
			mergo.Merge(&input.Params, template.Defaults)
			Expect(fetched.Params).To(Equal(input.Params))
		})

		It("should fetch a template for a given name and unexistent non-default locale", func() {
			tc := templates.CreateTemplateCache(10)
			template := &models.Template{
				Name:     "test_template3",
				Locale:   "en",
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "de",
				Template:   "test_template3",
				Params:     map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			fetched, errFetch := templates.FetchTemplate(l, input, db, tc)
			Expect(errFetch).To(BeNil())
			Expect(fetched.App).To(Equal(input.App))
			Expect(fetched.Token).To(Equal(input.Token))
			Expect(fetched.Service).To(Equal(input.Service))
			Expect(fetched.PushExpiry).To(Equal(input.PushExpiry))
			Expect(fetched.Metadata).To(Equal(input.Metadata))
			Expect(fetched.Message).To(Equal(template.Body))

			// Not the one requestedm but the default
			Expect(fetched.Locale).To(Equal(template.Locale))

			// Merge input and template params
			mergo.Merge(&input.Params, template.Defaults)
			Expect(fetched.Params).To(Equal(input.Params))
		})

		It("should not fetch an unexistent template", func() {
			tc := templates.CreateTemplateCache(10)
			template := &models.Template{
				Name:     "test_template4",
				Locale:   "en",
				Defaults: map[string]interface{}{"param2": "templateValue2", "param3": "templateValue3"},
				Body:     map[string]interface{}{"alert": "%{param1}, %{param2}, %{param3}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "de",
				Template:   "test_templatee4",
				Params:     map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			_, errFetch := templates.FetchTemplate(l, input, db, tc)
			Expect(errFetch).NotTo(BeNil())
		})
	})

	Describe("", func() {
		It("should fetch template for input messages and output the result templated message", func() {
			template := &models.Template{
				Name:     "test_template5",
				Locale:   "en",
				Defaults: map[string]interface{}{"param1": "default1", "param2": "default2"},
				Body:     map[string]interface{}{"alert": "%{value1}, %{value2}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			inChan := make(chan *messages.InputMessage, 10)
			outChan := make(chan *messages.TemplatedMessage, 10)
			doneChan := make(chan struct{}, 2)
			defer close(doneChan)

			go templates.Fetcher(l, inChan, outChan, doneChan, db)

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "de",
				Template:   "test_template5",
				Params:     map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			inChan <- input

			templatedMessage := <-outChan
			Expect(templatedMessage.App).To(Equal(input.App))
			Expect(templatedMessage.Token).To(Equal(input.Token))
			Expect(templatedMessage.Service).To(Equal(input.Service))
			Expect(templatedMessage.PushExpiry).To(Equal(input.PushExpiry))
			Expect(templatedMessage.Metadata).To(Equal(input.Metadata))
			Expect(templatedMessage.Message).To(Equal(template.Body))

			// Not the one requestedm but the default
			Expect(templatedMessage.Locale).To(Equal(template.Locale))

			// Merge input and template params
			mergo.Merge(&input.Params, template.Defaults)
			Expect(templatedMessage.Params).To(Equal(input.Params))
		})

		It("should not fetch unexistent template", func() {
			template := &models.Template{
				Name:     "test_template6",
				Locale:   "en",
				Defaults: map[string]interface{}{"param1": "default1", "param2": "default2"},
				Body:     map[string]interface{}{"alert": "%{value1}, %{value2}"},
			}
			errInsert := db.Insert(template)
			Expect(errInsert).To(BeNil())

			inChan := make(chan *messages.InputMessage, 10)
			outChan := make(chan *messages.TemplatedMessage, 10)
			doneChan := make(chan struct{}, 2)
			defer close(doneChan)

			go templates.Fetcher(l, inChan, outChan, doneChan, db)

			input := &messages.InputMessage{
				App:        "colorfy",
				Token:      "token",
				Service:    "gcm",
				PushExpiry: 0,
				Locale:     "de",
				Template:   "test_templatee5",
				Params:     map[string]interface{}{"param1": "inputValue1", "param3": "inputValue3"},
				Message:    map[string]interface{}{},
				Metadata:   map[string]interface{}{"meta": "data"},
			}

			inChan <- input

			time.Sleep(500 * time.Millisecond)
			Expect(len(outChan)).To(Equal(0))
		})
	})
})
