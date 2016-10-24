package api_test

// import (
// 	"testing"
//
// 	. "github.com/franela/goblin"
// 	. "github.com/onsi/gomega"
// )
//
// func Test(t *testing.T) {
// 	g := Goblin(t)
//
// 	// special hook for gomega
// 	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
//
// 	g.Describe("App Struct", func() {
// 		g.It("should create app with custom arguments", func() {
// 			application := GetApplication("127.0.0.1", 9999, "../config/test.yaml", false, false)
// 			Expect(application.Port).To(Equal(9999))
// 			Expect(application.Host).To(Equal("127.0.0.1"))
// 		})
// 	})
// }
