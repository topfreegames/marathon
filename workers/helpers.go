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
	var ll zap.Option
	if logProps.Level == "" {
		logProps.Level = config.GetString("workers.logger.level")
	}
	switch logProps.Level {
	case "debug":
		ll = zap.DebugLevel
	case "info":
		ll = zap.InfoLevel
	case "warn":
		ll = zap.WarnLevel
	case "error":
		ll = zap.ErrorLevel
	case "panic":
		ll = zap.PanicLevel
	case "fatal":
		ll = zap.FatalLevel
	default:
		ll = zap.InfoLevel
	}
	return zap.New(
		zap.NewJSONEncoder(), // drop timestamps in tests
		ll,
		zap.AddCaller(),
	)
}
