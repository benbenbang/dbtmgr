package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var AWS_DEFAULT_REGION = "eu-west-1"
var DBT_STATE_BUCKET string
var DBT_STATE_KEY string
var DBT_LOCK_KEY string
var Version string

func Initialize() {
	// 1. From the current path (last priority, where the binary is executed)
	viper.AddConfigPath(".")
	viper.SetConfigName("config") // no need to include file extension

	// 2. From the home directory (second priority)
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Unable to find home directory:", err)
		return
	}
	viper.AddConfigPath(home)

	// Attempt to read the config
	err = viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		fmt.Printf("Fatal error config file: %s \n", err)
	}

	// 3. Read environment variables that match (highest priority)
	viper.AutomaticEnv()
}
