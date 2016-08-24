package main

import (
	"git.topfreegames.com/topfreegames/marathon/cmd"
	"github.com/uber-go/zap"
)

func main() {
	ll := zap.InfoLevel
	l := zap.New(
		zap.NewJSONEncoder(), // drop timestamps in tests
		ll,
		zap.AddCaller(),
	)

	cmdL := l.With(
		zap.String("source", "main"),
		zap.String("operation", "Execute"),
	)

	cmdL.Debug("Executing root cmd...")
	cmd.Execute(cmd.RootCmd, cmdL)
}
