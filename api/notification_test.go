package api_test

//
// import (
// 	"encoding/json"
// 	"net/http"
//
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )
//
// var _ = Describe("Marathon API Handler", func() {
// 	BeforeEach(func() {})
//
// 	Describe("Create Notification Handler", func() {
// 		FIt("Should create a Notification", func() {
// 			a := GetDefaultTestApp()
// 			// type notificationPayload struct {
// 			// 	App      string
// 			// 	Service  string
// 			// 	Filters  map[string]interface{}
// 			// 	Template string
// 			// 	Params   map[string]interface{}
// 			// 	Message  map[string]interface{}
// 			// 	Metadata map[string]interface{}
// 			// }
//
// 			payload := map[string]interface{}{
// 				"app":     "app",
// 				"service": "service",
// 				"filters": map[string]interface{}{
// 					"a": "b",
// 				},
// 			}
// 			res := PostJSON(a, "/apps/appName/users/notifications", payload)
//
// 			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
// 			var result map[string]interface{}
// 			json.Unmarshal([]byte(res.Body().Raw()), &result)
// 			Expect(result["success"]).To(BeTrue())
// 		})
// 	})
// })
