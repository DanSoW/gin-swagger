package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var LogInfo = &logrus.Logger{
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.InfoLevel,
}

func InitLogInfo(file *os.File) {
	file, err := os.OpenFile("./logs/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		LogInfo.SetOutput(file)
	} else {
		LogInfo.Error("Failed to log to file, using default stderr")
	}
}
