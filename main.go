package main

import (
	"os"
	"statectl/cmd"
	"statectl/internal/logging"
)

func main() {
	log := logging.GetLogger()

	if err := cmd.DefaultCmd.Execute(); err != nil {
		if err != nil {
			log.Errorln(err)
			os.Exit(1)
		}
	}
}
