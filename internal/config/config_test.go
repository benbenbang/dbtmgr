package config_test

import (
	"os"
	"statectl/internal/config"
	"testing"

	"github.com/spf13/viper"
)

func TestInitializeConfig(t *testing.T) {
	config.Initialize()
	os.Setenv("STATECTL_TEST", "test")

	if viper.GetString("STATECTL_TEST") != "test" {
		t.Errorf("error initializing config")
	}
	viper.Reset()
}
