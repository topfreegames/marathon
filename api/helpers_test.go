package api_test

import (
	"net/http"
	"testing"

	"git.topfreegames.com/topfreegames/marathon/api"
	"github.com/gavv/httpexpect"
)

// GetDefaultTestApplication returns a new Marathon API Applicationlication bound to 0.0.0.0:8888 for test
func GetDefaultTestApplication() *api.Application {
	return api.GetApplication("0.0.0.0", 8888, "../config/test.yaml", true)
}

// Get returns a test request against specified URL
func Get(application *api.Application, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(application, "GET", url, t)
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(application *api.Application, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(application, "POST", url, t, payload)
}

// PutBody returns a test request against specified URL
func PutBody(application *api.Application, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(application, "PUT", url, t, payload)
}

func sendBody(application *api.Application, method string, url string, t *testing.T, payload string) *httpexpect.Response {
	req := sendRequest(application, method, url, t)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(application *api.Application, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(application, "POST", url, t, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(application *api.Application, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(application, "PUT", url, t, payload)
}

func sendJSON(application *api.Application, method, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(application, method, url, t)
	return req.WithJSON(payload).Expect()
}

func sendRequest(application *api.Application, method, url string, t *testing.T) *httpexpect.Request {
	handler := application.Application.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	return e.Request(method, url)
}
