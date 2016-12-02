package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/marathon/api"
	. "github.com/topfreegames/marathon/testing"
	"github.com/uber-go/zap"
)

var _ = Describe("API Application", func() {
	var logger zap.Logger
	BeforeEach(func() {
		logger = zap.New(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			zap.FatalLevel,
		)
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app := api.GetApplication("127.0.0.1", 9999, true, logger, GetConfPath())
			Expect(app).NotTo(BeNil())
			Expect(app.Host).To(Equal("127.0.0.1"))
			Expect(app.Port).To(Equal(9999))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App authentication", func() {
		var app *api.Application

		BeforeEach(func() {
			app = api.GetApplication("127.0.0.1", 9999, false, logger, GetConfPath())
		})

		It("should fail if no user", func() {
			status, _ := Get(app, "/healthcheck", "")
			Expect(status).To(Equal(http.StatusUnauthorized))
		})

		It("should work if valid user", func() {
			status, _ := Get(app, "/healthcheck", "test@test.com")
			Expect(status).To(Equal(200))
		})
	})

	Describe("Error Handler", func() {
		var sink *TestBuffer
		BeforeEach(func() {
			sink = &TestBuffer{}
			logger = zap.New(
				zap.NewJSONEncoder(zap.NoTime()),
				zap.Output(sink),
				zap.WarnLevel,
			)
		})

		It("should handle errors when they are error objects", func() {
			app := api.GetApplication("127.0.0.1", 9999, false, logger, GetConfPath())

			app.OnErrorHandler(fmt.Errorf("some other error occurred"), []byte("stack"))
			result := sink.Buffer.String()
			var obj map[string]interface{}
			err := json.Unmarshal([]byte(result), &obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj["level"]).To(Equal("error"))
			Expect(obj["msg"]).To(Equal("Panic occurred."))
			Expect(obj["operation"]).To(Equal("OnErrorHandler"))
			Expect(obj["stack"]).To(Equal("stack"))
		})
	})
})
