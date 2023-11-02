package main

import (
	"dbtmgr/cmd"
	"dbtmgr/internal/logging"
	"os"
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
