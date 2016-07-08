package api

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

// GetDefaultTestApplication returns a new Marathon API Applicationlication bound to 0.0.0.0:8888 for test
func GetDefaultTestApplication() *Application {
	return GetApplication("0.0.0.0", 8888, "../config/test.yaml", true)
}

// Get returns a test request against specified URL
func Get(application *Application, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(application, "GET", url, t)
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(application *Application, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(application, "POST", url, t, payload)
}

// PutBody returns a test request against specified URL
func PutBody(application *Application, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(application, "PUT", url, t, payload)
}

func sendBody(application *Application, method string, url string, t *testing.T, payload string) *httpexpect.Response {
	req := sendRequest(application, method, url, t)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(application *Application, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(application, "POST", url, t, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(application *Application, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(application, "PUT", url, t, payload)
}

func sendJSON(application *Application, method, url string, t *testing.T, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(application, method, url, t)
	return req.WithJSON(payload).Expect()
}

func sendRequest(application *Application, method, url string, t *testing.T) *httpexpect.Request {
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
