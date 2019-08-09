package router

import (
	"github.com/gin-gonic/gin"
	"goFrame/app/ping"
)

// RegisteredRoot ...
func RegisteredRoot(app *gin.Engine) {
	index := app.Group("/")
	index.Handle("GET", "/ping", ping.Ping)
}
