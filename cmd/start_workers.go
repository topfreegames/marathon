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
		worker := workers.GetContinuousWorker(ConfigFile)
		worker.StartWorker()
		sigsKill := make(chan os.Signal, 1)
		signal.Notify(sigsKill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-sigsKill
		Logger.Info("Closing workers", zap.String("signal", fmt.Sprintf("%+v", sig)))
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
