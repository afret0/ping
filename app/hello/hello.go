package hello

import (
	"github.com/gin-gonic/gin"
	"goFrame/config"
)

func Hello(ctx *gin.Context) {
	conf := config.GetConf()
	ctx.JSON(200, conf.Ping)
}
