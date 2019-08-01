package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	//logLevel:=viper.GetString("logLevel")
	log.SetLevel(log.ErrorLevel)

}
