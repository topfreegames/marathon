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
		l := zap.NewJSON(ll, zap.AddCaller())

		cmdL := l.With(
			zap.String("source", "startContinuousWorkersCmd"),
			zap.String("operation", "Run"),
		)

		cmdL.Debug("Get continuous workers...")
		worker := workers.GetContinuousWorker(ConfigFile)

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
	// startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind marathon to")
	// startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind marathon to")
	// startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
