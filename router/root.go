package router

import "github.com/gin-gonic/gin"
import "goFrame/app/ping"

var app *gin.Engine

func init() {
	app = gin.Default()
	index := app.Group("/")
	index.Handle("GET", "/ping", ping.Ping)
}

func GetRouter() *gin.Engine {
	return app
}
