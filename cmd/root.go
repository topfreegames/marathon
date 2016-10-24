package cmd

import (
	"os"
	"strings"

	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// RootCmd is the root command for marathon CLI application
var RootCmd = &cobra.Command{
	Use:   "marathon",
	Short: "Marathon sends push-notifications",
	Long:  `Use marathon to send push-notifications to apple (apns) and google (gcm).`,
}

// Execute runs RootCmd to initialize marathon CLI application
func Execute(cmd *cobra.Command, l zap.Logger) {
	if err := cmd.Execute(); err != nil {
		log.F(l, "Error", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		os.Exit(-1)
	}
}

func init() {}

// InitConfig reads in config file and ENV variables if set.
func InitConfig(l zap.Logger, configFile string) *viper.Viper {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.F(l, "Config file does not exist", func(cm log.CM) {
			cm.Write(zap.String("configFile", configFile))
		})
		os.Exit(-1)
	}
	var config = viper.New()
	config.SetConfigFile(configFile)

	config.SetConfigType("yaml")
	config.SetEnvPrefix("marathon")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AddConfigPath(".") // optionally look for config in the working directory
	config.AutomaticEnv()     // read in environment variables that match

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		log.F(l, "Error", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		os.Exit(-1)
	}
	log.I(l, "Using config file", func(cm log.CM) {
		cm.Write(zap.String("Config file", viper.ConfigFileUsed()))
	})
	return config
}
