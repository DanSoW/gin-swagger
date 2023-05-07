package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var LogFatal = &logrus.Logger{
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.FatalLevel,
}

func InitLogFatal(file *os.File) {
	file, err := os.OpenFile("./logs/fatal.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		LogFatal.SetOutput(file)
	} else {
		LogFatal.Error("Failed to log to file, using default stderr")
	}
}
