package cmd

import (
	"os"

	"github.com/uber-go/zap"
)

func getLogLevel() zap.Level {
	var level = zap.WarnLevel
	var environment = os.Getenv("ENV")
	if environment == "test" {
		level = zap.FatalLevel
	}
	return level
}

// Logger is the template fetcher logger
var Logger = zap.NewJSON(getLogLevel())
