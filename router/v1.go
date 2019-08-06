package router

import (
	"github.com/gin-gonic/gin"
	"goFrame/app/hello"
)

func RegisteredHello(app *gin.Engine) {
	helloRouter := app.Group("/hello")
	helloRouter.Handle("GET", "/", hello.Hello)
}
