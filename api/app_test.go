package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("API Application", func() {
	BeforeEach(func() {
		var err error
		testDb, err := GetTestDB()
		Expect(err).NotTo(HaveOccurred())
	})
})
