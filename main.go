package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"goFrame/app/ping"
	_ "goFrame/conf"
	"goFrame/log"
	"os"
)

func init() {
	switch os.Getenv("ENV") {
	case "pro":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func main() {
	logger := log.GetLogger()
	logger.Info(viper.GetString("root"))
	root := gin.Default()
	root.Group("/")
	root.Handle("GET", "ping", ping.Ping)
	err := root.Run(":" + viper.GetString("port"))
	if err != nil {
		logger.Error(err)
	}
}
