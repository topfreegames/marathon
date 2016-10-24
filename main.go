package main

import (
	"git.topfreegames.com/topfreegames/marathon/cmd"
	"git.topfreegames.com/topfreegames/marathon/log"
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

	log.D(cmdL, "Executing root cmd...")
	cmd.Execute(cmd.RootCmd, cmdL)
}
