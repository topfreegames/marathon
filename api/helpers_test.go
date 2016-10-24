package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"git.topfreegames.com/topfreegames/marathon/api"
	"git.topfreegames.com/topfreegames/marathon/log"
	"git.topfreegames.com/topfreegames/marathon/models"
	mt "git.topfreegames.com/topfreegames/marathon/testing"
	"github.com/labstack/echo/engine/standard"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// GetTestDB returns a connection to the test database
func GetTestDB(l zap.Logger) (*models.DB, error) {
	return models.GetDB(l, "localhost", "marathon", 9910, "disable", "marathon", "")
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB(l zap.Logger) *models.DB {
	faultyDb, _ := models.InitDb(l, "localhost", "marathon_tet", 9910, "disable", "marathon", "")
	return faultyDb
}

func getConfig() (*viper.Viper, error) {
	var config = viper.New()
	config.SetConfigFile("./../config/test.yaml")
	config.SetEnvPrefix("marathon")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetDefaultTestApp returns a new Marathon API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp(config *viper.Viper) *api.Application {
	l := mt.NewMockLogger()
	if config != nil {
		application := api.GetApplication("0.0.0.0", 8888, config, true, true, l)
		return application
	} else {
		cfg, err := getConfig()
		if err != nil {
			log.P(l, "Could not load config", func(cm log.CM) {
				cm.Write(zap.Object("config", config))
			})
		}
		application := api.GetApplication("0.0.0.0", 8888, cfg, true, true, l)
		return application
	}
}

// Get from server
func Get(app *api.Application, url string) (int, string) {
	return doRequest(app, "GET", url, "")
}

// Post to server
func Post(app *api.Application, url, body string) (int, string) {
	return doRequest(app, "POST", url, body)
}

// PostJSON to server
func PostJSON(app *api.Application, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Post(app, url, string(result))
}

// Put to server
func Put(app *api.Application, url, body string) (int, string) {
	return doRequest(app, "PUT", url, body)
}

// PutJSON to server
func PutJSON(app *api.Application, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Put(app, url, string(result))
}

// Delete from server
func Delete(app *api.Application, url string) (int, string) {
	return doRequest(app, "DELETE", url, "")
}

// GinkgoReporter implements tests for httpexpect
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

var client *http.Client
var transport *http.Transport

func initClient() {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
}

func doRequest(application *api.Application, method, url, body string) (int, string) {
	initClient()
	defer transport.CloseIdleConnections()
	application.Engine.SetHandler(application.Application)
	ts := httptest.NewServer(application.Engine.(*standard.Server))

	var bodyBuff io.Reader
	if body != "" {
		bodyBuff = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", ts.URL, url), bodyBuff)
	req.Header.Set("Connection", "close")
	req.Close = true
	Expect(err).NotTo(HaveOccurred())

	res, err := client.Do(req)
	ts.Close()
	//Wait for port of httptest to be reclaimed by OS
	time.Sleep(50 * time.Millisecond)
	Expect(err).NotTo(HaveOccurred())

	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	Expect(err).NotTo(HaveOccurred())

	return res.StatusCode, string(b)
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
