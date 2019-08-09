package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"goFrame/config"
	"goFrame/libs"
	"goFrame/router"
	"os"
)

func init() {
	switch os.Getenv("ENV") {
	case "pro":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	defer libs.CloseRedis()
}

func main() {
	logger := libs.GetLogger()
	logger.Info(viper.GetString("root"))

	conf := config.GetConf()
	//app := gin.New()
	app := gin.Default()
	//app.Use(gin.Logger())

	router.RegisteredRoot(app)
	router.RegisteredHello(app)

	err := app.Run(":" + conf.Port)

	if err != nil {
		logger.Error(err)
	}
}
