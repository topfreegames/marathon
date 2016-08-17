package api_test

import (
	"net/http"

	"git.topfreegames.com/topfreegames/marathon/models"
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
		_db, err := models.GetTestDB(l)
		Expect(err).NotTo(HaveOccurred())
		Expect(_db).NotTo(BeNil())
	})

	Describe("Create Notification Handler", func() {
		It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp()
			res := Get(a, "/healthcheck")

			res.Status(http.StatusOK).Body().Equal("working")
		})
		It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp()
			a.Config.Set("healthcheck.workingText", "OTHERWORKING")
			res := Get(a, "/healthcheck")

			res.Status(http.StatusOK).Body().Equal("OTHERWORKING")
		})
	})
})
