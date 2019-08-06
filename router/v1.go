package router

import (
	"github.com/gin-gonic/gin"
	"goFrame/app/hello"
)

func RegisteredHello(app *gin.Engine) {
	helloRouter := app.Group("/v1")
	helloRouter.Handle("GET", "/hello", hello.Hello)
}
