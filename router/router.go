package router

import "github.com/gin-gonic/gin"
import "goFrame/app/ping"

func InitRouter() *gin.Engine {
	r := gin.Default()
	index := r.Group("/")
	index.Handle("GET", "/", ping.Ping)
	return r
}
