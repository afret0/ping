package main

import (
	"github.com/gin-gonic/gin"
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
}

func main() {
	//defer redis.CloseRedis()
	//defer mongo.CloseMongoSession()

	logger := libs.GetLogger()
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
