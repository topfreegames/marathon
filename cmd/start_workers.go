package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.topfreegames.com/topfreegames/marathon/workers"

	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
)

var app, service string

// startCmd represents the start command
var startContinuousWorkersCmd = &cobra.Command{
	Use:   "start-workers",
	Short: "starts the marathon workers",
	Long:  `Starts marathon workers with the specified arguments. You can use environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		l := zap.New(
			zap.NewJSONEncoder(), // drop timestamps in tests
			ll,
			zap.AddCaller(),
		)

		cmdL := l.With(
			zap.String("source", "startContinuousWorkersCmd"),
			zap.String("operation", "Run"),
		)

		continuousWorkerConf := &workers.ContinuousWorker{
			App:        app,
			Service:    service,
			ConfigPath: "./../config/test.yaml",
		}
		cmdL.Debug("Get continuous workers...")
		worker, err := workers.GetContinuousWorker(continuousWorkerConf)
		if err != nil {
			cmdL.Error("Error", zap.Error(err))
			os.Exit(-1)
		}

		cmdL.Debug("Starting continuous workers...")
		worker.StartWorker()
		sigsKill := make(chan os.Signal, 1)
		signal.Notify(sigsKill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-sigsKill

		cmdL.Info("Closing workers", zap.String("signal", fmt.Sprintf("%+v", sig)))
		worker.Close()
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(startContinuousWorkersCmd)
	startContinuousWorkersCmd.Flags().StringVarP(&app, "app", "a", "", "app to consume")
	startContinuousWorkersCmd.Flags().StringVarP(&service, "service", "s", "", "service to consume")
	// startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind marathon to")
	// startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
