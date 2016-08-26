package api_test

import (
	// "encoding/json"
	// "net/http"

	mt "git.topfreegames.com/topfreegames/marathon/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Marathon API Handler", func() {
	var (
		l zap.Logger
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
	})

	Describe("Create upload handler", func() {
		It("Should create a Notification (will fail without the AWS credentials", func() {
			a := GetDefaultTestApp()

			res := Get(a, "/uploadurl")

			// Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			Expect(res.Raw().StatusCode).To(Equal(400))
			// var result map[string]interface{}
			// json.Unmarshal([]byte(res.Body().Raw()), &result)
			// Expect(result["success"]).To(BeTrue())
		})
	})
})
