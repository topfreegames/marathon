package cmd

import "github.com/spf13/cobra"

var host string
var port int
var debug bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the marathon API server",
	Long: `Starts marathon server with the specified arguments. You can use
environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		// app := api.GetApp(
		// 	host,
		// 	port,
		// 	ConfigFile,
		// 	debug,
		// )
		//
		// app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind marathon to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8888, "Port to bind marathon to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
