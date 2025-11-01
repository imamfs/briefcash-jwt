package helper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger(logFile string, level logrus.Level) {
	Logger = logrus.New()

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Logger.SetLevel(level)

	dir := filepath.Dir(logFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
			Logger.SetOutput(os.Stdout)
			return
		}
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		Logger.SetOutput(os.Stdout)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger.SetOutput(multiWriter)

	Logger.Infof("Logger initiate. Log file: %s", logFile)
}
