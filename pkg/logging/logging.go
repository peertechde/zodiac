package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultLogger is the default base logger.
var Logger = defaultLogger()

func defaultLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	}
	logger.SetLevel(logrus.InfoLevel)
	return logger
}
