package config_test

import (
	"statectl/internal/config"
	"testing"
)

func TestColor(t *testing.T) {
	config.Blue("blue")
	config.Green("green")
	config.Red("red")
	config.Yellow("yellow")
	config.Magenta("magenta")
	config.Cyan("cyan")
	config.White("white")
	config.Bold("bold")
}
