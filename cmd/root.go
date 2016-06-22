package cmd

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigFile is the file containing all config
var ConfigFile string

// RootCmd is the root command for marathon CLI application
var RootCmd = &cobra.Command{
	Use:   "marathon",
	Short: "Marathon sends push-notifications",
	Long:  `Use marathon to send push-notifications to apple (apns) and google (gcm).`,
}

// Execute runs RootCmd to initialize marathon CLI application
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		glog.Error(err)
		os.Exit(-1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(
		&ConfigFile, "config", "c", "./config/local.yaml",
		"config file (default is ./config/local.yaml",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if ConfigFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(ConfigFile)
	}
	viper.SetEnvPrefix("marathon")
	viper.SetConfigName(".marathon") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		glog.Infof("Using config file:", viper.ConfigFileUsed())
	}
}
