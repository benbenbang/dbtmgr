package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Set the log output as os.Stderr
	log.SetOutput(os.Stderr)

	// Set the log level
	logLevel, ok := os.LookupEnv("DBTMGR_LOG_LEVEL")
	if !ok {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.WithField("log_level", logLevel).Error("Invalid log level; using 'info' as default")
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
}

func GetLogger() *logrus.Logger {
	return log
}
