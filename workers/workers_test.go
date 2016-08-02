package workers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWorkers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workers")
}
