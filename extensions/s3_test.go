// /*
// * Copyright (c) 2016 TFG Co <backend@tfgco.com>
// * Author: TFG Co <backend@tfgco.com>
// *
// * Permission is hereby granted, free of charge, to any person obtaining a copy of
// * this software and associated documentation files (the "Software"), to deal in
// * the Software without restriction, including without limitation the rights to
// * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// * the Software, and to permit persons to whom the Software is furnished to do so,
// * subject to the following conditions:
// *
// * The above copyright notice and this permission notice shall be included in all
// * copies or substantial portions of the Software.
// *
// * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// */
package extensions_test

import (
	"bytes"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/extensions"
	"github.com/uber-go/zap"
)

var _ = Describe("S3 Extension", func() {
	var logs *bytes.Buffer
	var logger zap.Logger

	BeforeEach(func() {
		logs = &bytes.Buffer{}
		logger = zap.New(zap.NewJSONEncoder(zap.NoTime()), zap.Output(zap.AddSync(logs)))
	})

	newConf := func(accessKey, secretAccessKey string) *viper.Viper {
		conf := viper.New()
		conf.Set("s3.region", "us-east-1")
		conf.Set("s3.accessKey", accessKey)
		conf.Set("s3.secretAccessKey", secretAccessKey)
		return conf
	}

	It("should use static credentials when both keys are set", func() {
		s, err := extensions.NewS3(newConf("AKIAFAKESTATIC", "fake-secret"), logger)
		Expect(err).NotTo(HaveOccurred())
		Expect(logs.String()).To(ContainSubstring(`"credentialsSource":"static"`))

		u, err := s.PutObjectRequest("marathon-test-bucket/jobs/job.csv")
		Expect(err).NotTo(HaveOccurred())
		q, err := url.Parse(u)
		Expect(err).NotTo(HaveOccurred())
		Expect(q.Query().Get("X-Amz-Credential")).To(HavePrefix("AKIAFAKESTATIC/"))
	})

	It("should use the default credential chain when a key is missing", func() {
		_, err := extensions.NewS3(newConf("AKIAFAKESTATIC", ""), logger)
		Expect(err).NotTo(HaveOccurred())
		Expect(logs.String()).To(ContainSubstring(`"credentialsSource":"default-chain"`))
	})
})
