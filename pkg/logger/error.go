package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var LogError = &logrus.Logger{
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.ErrorLevel,
}

func InitLogError(file *os.File) {
	file, err := os.OpenFile("./logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		LogError.SetOutput(file)
	} else {
		LogError.Error("Failed to log to file, using default stderr")
	}
}
