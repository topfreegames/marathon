package api

import (
	"fmt"
	"strings"

	"github.com/kataras/iris"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(application *Application) func(c *iris.Context) {
	return func(c *iris.Context) {
		workingString := application.Config.GetString("healthcheck.workingText")
		_, err := application.Db.SelectInt("select count 1")
		if err != nil {
			c.Write(fmt.Sprintf("Error connecting to database: %s", err))
			c.SetStatusCode(500)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		workingString = strings.TrimSpace(workingString)
		c.Write(workingString)
		c.SetHeader("MARATHON-VERSION", VERSION)
	}
}
