package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(application *Application) func(c echo.Context) error {
	return func(c echo.Context) error {
		workingString := application.Config.GetString("healthcheck.workingText")
		_, err := application.Db.SelectInt("SELECT COUNT(1)")
		if err != nil {
			return fmt.Errorf("Error connecting to database: %s", err)
		}

		res, err := application.RedisClient.Client.Ping().Result()
		if err != nil || res != "PONG" {
			return fmt.Errorf("Error connecting to redis: %s", err)
		}

		workingString = strings.TrimSpace(workingString)
		return c.String(http.StatusOK, workingString)
	}
}
