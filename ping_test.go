package ping

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func offlineHandle(uid string) error {
	GetLogger().Infof("offline handle: %s", uid)
	//return ErrRetry
	return nil
}

func Test_ping(t *testing.T) {
	RC := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{"r-bp1kvud328x48r9xp6pd.redis.rds.aliyuncs.com:6379"},
		Password: "Qiyiguo0621",
		Username: "kiwi0621",
	})

	svr := NewService(RC, &Option{
		Prefix:     "test",
		OfflineTTL: 5,
	})

	svr.RegisterOfflineHandle(offlineHandle)
	svr.StartTick()

	for i := 0; i <= 1000; i++ {
		//_ = svr.Ping(context.Background(), fmt.Sprintf("%d", i))
		if i < 20 {
			_ = svr.Ping(context.Background(), fmt.Sprintf("%d", 1))
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Hour)
}
