package helper

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger(logFile string, level logrus.Level) {
	Logger = logrus.New()

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Logger.SetLevel(level)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		Logger.SetOutput(file)
	} else {
		Logger.SetOutput(os.Stdout)
	}
}
