package ping

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func offlineHandle(e *Event) error {
	GetLogger().Infof("offline handle: %+v", e)
	return ErrRetry
	return nil
}

func Test_ping(t *testing.T) {
	RC := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{"120.27.235.209:6379"},
		Password: "Qiyiguo2303",
		Username: "default",
	})

	svr := NewService(RC, &Option{
		Prefix:     "test",
		OfflineTTL: 15,
	})
	svr.RegisterOfflineHandle(offlineHandle)
	svr.StartTick()

	for i := 0; i <= 100; i++ {
		_ = svr.Ping(context.Background(), fmt.Sprintf("%d", i))
		time.Sleep(5 * time.Second)
	}

	time.Sleep(1 * time.Hour)
}
