package config

import (
	"os"
	"statectl/internal/logging"

	"github.com/spf13/viper"
)

func init() {
	Initialize()
}

// Version will be injected by the build process
var Version string

// Initialize the logger
var log = logging.GetLogger()

func Initialize() {
	// 1. From the current path (last priority, where the binary is executed)
	viper.AddConfigPath(".")
	viper.SetConfigName("config") // no need to include file extension

	// 2. From the home directory (second priority)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorln("Unable to find home directory:", err)
		return
	}
	viper.AddConfigPath(home)

	// Attempt to read the config
	err = viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Debugf("Fatal error config file: %s\n", err)
	}

	// 3. Read environment variables that match (highest priority)
	viper.AutomaticEnv()
}
