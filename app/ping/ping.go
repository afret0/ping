package ping

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goFrame/libs"
)

func Ping(ctx *gin.Context) {
	logger := libs.GetLogger()
	db := 0
	redis := libs.GetRedis(&db)
	afreto, err := redis.Get("afreto").Result()
	if err != nil {
		logger.Info(err)
	}
	fmt.Println(afreto)

	dbName := "uki"
	col := "order"
	orderCol := libs.GetMongoCol(&dbName, &col)
	query := make(map[string]string)
	query["owner"] = "74"
	var orderINfo *interface{}
	err = orderCol.Find(query).One(&orderINfo)
	if err != nil {
		logger.Error(err)
	}
	logger.Warn(*orderINfo)
	ctx.JSON(200, &afreto)
}
