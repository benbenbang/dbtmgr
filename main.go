package main

import (
	"os"
	"statemgr/cmd"
	"statemgr/internal/logging"
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
