package workers

import (
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Log defines a Log property
type Log struct {
	Level string
}

// ConfigureLogger defines a Log property
func ConfigureLogger(logProps Log, config *viper.Viper) zap.Logger {
	var level zap.Option
	if logProps.Level == "" {
		logProps.Level = config.GetString("workers.logger.level")
	}
	switch logProps.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}
	return zap.NewJSON(level)
}
