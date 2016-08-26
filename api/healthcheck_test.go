package api_test

import (
	"net/http"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var _ = Describe("Marathon API Handler", func() {
	var (
		l      zap.Logger
		config *viper.Viper
	)
	BeforeEach(func() {
		l = mt.NewMockLogger()
		_db, err := models.GetTestDB(l)
		Expect(err).NotTo(HaveOccurred())
		Expect(_db).NotTo(BeNil())

		config = viper.New()
		config.SetConfigFile("./../config/test.yaml")
		config.SetEnvPrefix("marathon")
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()
		err = config.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Notification Handler", func() {
		It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp(config)
			res := Get(a, "/healthcheck")

			res.Status(http.StatusOK).Body().Equal("working")
		})
		It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp(config)
			a.Config.Set("healthcheck.workingText", "OTHERWORKING")
			res := Get(a, "/healthcheck")

			res.Status(http.StatusOK).Body().Equal("OTHERWORKING")
		})
	})
})
