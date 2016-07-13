package cmd

import (
	"git.topfreegames.com/topfreegames/marathon/api"
	"github.com/spf13/cobra"
)

var host string
var port int
var debug bool

// startCmd represents the start command
var startServerCmd = &cobra.Command{
	Use:   "start-server",
	Short: "starts the marathon API server",
	Long:  `Starts marathon server with the specified arguments. You can use environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := api.GetApplication(host, port, ConfigFile, debug)
		app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startServerCmd)

	startServerCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind marathon to")
	startServerCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind marathon to")
	startServerCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
