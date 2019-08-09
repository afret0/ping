package libs

import (
	"fmt"
	"github.com/go-redis/redis"
	"goFrame/config"
	"strconv"
)

type redisClientStoreStruct struct {
	_redis map[int]*redis.Client
}

var redisClientStore *redisClientStoreStruct

func init() {
	redisClientStore = new(redisClientStoreStruct)
	redisClientStore._redis = make(map[int]*redis.Client)
	db := 0
	_ = GetRedis(&db)
}

func GetRedis(db *int) *redis.Client {
	if redisClient, ok := redisClientStore._redis[*db]; ok {
		return redisClient
	} else {
		redisClient = newRedis(db)
		redisClientStore._redis[*db] = redisClient
		return redisClient
	}
}

func CloseRedis() {
	for _, v := range redisClientStore._redis {
		_ = v.Close()
	}
}

func newRedis(db *int) *redis.Client {
	conf := config.GetConf()

	redisOptions := new(redis.Options)
	redisOptions.Addr = conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port)
	redisOptions.Password = conf.Redis.Pwd
	redisOptions.DB = *db

	client := redis.NewClient(redisOptions)
	_, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("redis 链接失败, %s", err))
	}
	return client
}
