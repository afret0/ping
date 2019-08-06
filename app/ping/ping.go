package ping

import (
	"github.com/gin-gonic/gin"
	"goFrame/config"
)

func Ping(ctx *gin.Context) {
	conf := config.GetConf()
	ctx.JSON(200, conf.Ping)
}
