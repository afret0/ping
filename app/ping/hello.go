package ping

import "github.com/gin-gonic/gin"

func Hello(ctx *gin.Context) {
	kitty := "kitty"
	ctx.JSON(200, kitty)
}
