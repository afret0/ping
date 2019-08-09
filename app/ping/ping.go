package ping

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goFrame/libs"
)

func Ping(ctx *gin.Context) {
	db := 0
	redis := libs.GetRedis(&db)
	afreto, _ := redis.Get("afreto").Result()
	fmt.Println(afreto)
	//conf := config.GetConf()
	ctx.JSON(200, &afreto)
}
