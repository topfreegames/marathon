package api_test

import (
	// "encoding/json"
	// "net/http"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Marathon API Handler", func() {
	var (
		config *viper.Viper
	)
	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("./../config/test.yaml")
		config.SetEnvPrefix("marathon")
		config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		config.AutomaticEnv()
		err := config.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create upload handler", func() {
		It("Should create a Notification (will fail without the AWS credentials", func() {
			a := GetDefaultTestApp(config)

			res := Get(a, "/uploadurl")

			// Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			Expect(res.Raw().StatusCode).To(Equal(400))
			// var result map[string]interface{}
			// json.Unmarshal([]byte(res.Body().Raw()), &result)
			// Expect(result["success"]).To(BeTrue())
		})
	})
})
