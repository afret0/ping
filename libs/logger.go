package libs

import (
	"github.com/sirupsen/logrus"
	"os"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	env := os.Getenv("ENV")
	env = "pro"
	if env == "pro" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		//logName := viper.GetString("app") + ".log"
		//file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		//if err != nil {
		//	fmt.Println(err)
		//} else {
		//	logger.SetOutput(file)
		//logger.Out = file
		//}

		logger.SetFormatter(&logrus.TextFormatter{DisableColors: false,
			FullTimestamp: false})
	}
	logger.SetLevel(logrus.InfoLevel)
	logger.SetReportCaller(true)
}

// GetLogger ...
func GetLogger() *logrus.Logger {
	return logger
}
