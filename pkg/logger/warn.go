package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var LogWarn = &logrus.Logger{
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.WarnLevel,
}

func InitLogWarn(file *os.File) {
	file, err := os.OpenFile("./logs/warn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		LogWarn.SetOutput(file)
	} else {
		LogWarn.Error("Failed to log to file, using default stderr")
	}
}
