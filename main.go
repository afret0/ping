package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goFrame/config"
	"goFrame/libs"
	"goFrame/routers"
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
	fmt.Println(conf.Ping)

	app := gin.New()
	app.Use(gin.Logger())
	//app.Use(gin.Recovery())

	routers.RegisteredRoot(app)

	//err := app.Run(":" + conf.Port)
	err := app.Run(":10110")

	if err != nil {
		logger.Error(err)
	}
}
