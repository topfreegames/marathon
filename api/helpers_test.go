package api_test

import (
	"fmt"
	"net/http"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/api"
	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/gavv/httpexpect"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp"
)

// GetTestDB returns a connection to the test database
func GetTestDB(l zap.Logger) (*models.DB, error) {
	return models.GetDB(l, "localhost", "marathon_test", 5432, "disable", "marathon_test", "")
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB(l zap.Logger) *models.DB {
	faultyDb, _ := models.InitDb(l, "localhost", "marathon_tet", 5432, "disable", "marathon_test", "")
	return faultyDb
}

// GetDefaultTestApp returns a new Marathon API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *api.Application {
	l := mt.NewMockLogger()
	// l := zap.NewJSON(zap.InfoLevel, zap.AddCaller())
	application := api.GetApplication("0.0.0.0", 8888, "../config/test.yaml", true, l)
	application.Configure()
	return application
}

// Get returns a test request against specified URL
func Get(application *api.Application, url string, queryString ...map[string]interface{}) *httpexpect.Response {
	req := SendRequest(application, "GET", url)
	if len(queryString) == 1 {
		for k, v := range queryString[0] {
			req = req.WithQuery(k, v)
		}
	}
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(application *api.Application, url string, payload string) *httpexpect.Response {
	return sendBody(application, "POST", url, payload)
}

// PutBody returns a test request against specified URL
func PutBody(application *api.Application, url string, payload string) *httpexpect.Response {
	return sendBody(application, "PUT", url, payload)
}

func sendBody(application *api.Application, method string, url string, payload string) *httpexpect.Response {
	req := SendRequest(application, method, url)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(application *api.Application, url string, payload map[string]interface{}) *httpexpect.Response {
	return SendJSON(application, "POST", url, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(application *api.Application, url string, payload map[string]interface{}) *httpexpect.Response {
	return SendJSON(application, "PUT", url, payload)
}

// SendJSON sends a json with a given payload to a given url through a gien method
func SendJSON(application *api.Application, method, url string, payload map[string]interface{}) *httpexpect.Response {
	req := SendRequest(application, method, url)
	return req.WithJSON(payload).Expect()
}

// Delete returns a test request against specified URL
func Delete(application *api.Application, url string) *httpexpect.Response {
	req := SendRequest(application, "DELETE", url)
	return req.Expect()
}

//GinkgoReporter implements tests for httpexpect
type GinkgoReporter struct {
}

// Errorf implements Reporter.Errorf.
func (g *GinkgoReporter) Errorf(message string, args ...interface{}) {
	Expect(false).To(BeTrue(), fmt.Sprintf(message, args...))
}

//GinkgoPrinter reports errors to stdout
type GinkgoPrinter struct{}

//Logf reports to stdout
func (g *GinkgoPrinter) Logf(source string, args ...interface{}) {
	fmt.Printf(source, args...)
}

// SendRequest sends a request to a given url through a given method
func SendRequest(application *api.Application, method, url string) *httpexpect.Request {
	api := application.Application
	srv := api.Servers.Main()

	if srv == nil { // maybe the user called this after .Listen/ListenTLS/ListenUNIX, the tester can be used as standalone (with no running iris instance) or inside a running instance/application
		srv = api.ListenVirtual(api.Config.Tester.ListeningAddr)
	}

	opened := api.Servers.GetAllOpened()
	h := srv.Handler
	baseURL := srv.FullHost()
	if len(opened) > 1 {
		baseURL = ""
		//we have more than one server, so we will create a handler here and redirect by registered listening addresses
		h = func(reqCtx *fasthttp.RequestCtx) {
			for _, s := range opened {
				if strings.HasPrefix(reqCtx.URI().String(), s.FullHost()) { // yes on :80 should be passed :80 also, this is inneed for multiserver testing
					s.Handler(reqCtx)
					break
				}
			}
		}
	}

	if api.Config.Tester.ExplicitURL {
		baseURL = ""
	}

	testConfiguration := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(h),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
	}
	if api.Config.Tester.Debug {
		testConfiguration.Printers = []httpexpect.Printer{
			httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		}
	}

	return httpexpect.WithConfig(testConfiguration).Request(method, url)
}

// GetGameRoute returns a clan route for the given game id.
func GetGameRoute(gameID, route string) string {
	return fmt.Sprintf("/games/%s/%s", gameID, route)
}

// CreateMembershipRoute returns a create membership route for the given game and clan id.
func CreateMembershipRoute(gameID, clanPublicID, route string) string {
	return fmt.Sprintf("/games/%s/clans/%s/memberships/%s", gameID, clanPublicID, route)
}

func str(value interface{}) string {
	return fmt.Sprintf("%v", value)
}
