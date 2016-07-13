package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
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
		Logger.Error("Error", zap.Error(err))
		os.Exit(-1)
	}
}

func init() {
	var environment = os.Getenv("ENV")
	if environment == "" {
		environment = "development"
	}
	filename := "./config/" + environment + ".yaml"
	RootCmd.PersistentFlags().StringVarP(
		// cobra.OnInitialize(initConfig)
		&ConfigFile, "config", "c", filename,
		"config file uses ENV to set `./config/<ENV>.yaml`",
	)
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if ConfigFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(ConfigFile)
	} else {
		var environment = os.Getenv("ENV")
		if environment == "" {
			environment = "development"
		}
		ConfigFile = "./config/" + environment + ".yml"
		viper.SetConfigFile(ConfigFile)
	}
	viper.SetEnvPrefix("marathon")
	viper.SetConfigName(".marathon") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Logger.Error(
			"Using config file",
			zap.String("Config file", viper.ConfigFileUsed()),
		)
	}
}
