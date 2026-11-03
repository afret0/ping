package ping

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func Test_uidOfMember(t *testing.T) {
	tests := []struct {
		name   string
		member string
		want   string
	}{
		{name: "新版成员就是裸 uid", member: "u1@x.com", want: "u1@x.com"},
		{name: "老版成员是 Event json, 取里面的 uid", member: `{"uid":"u1@x.com","retryCount":0}`, want: "u1@x.com"},
		{name: "重试副本也算老版编码", member: `{"uid":"u1@x.com","retryCount":3}`, want: "u1@x.com"},
		{name: "空成员取不到 uid", member: "", want: ""},
		{name: "空白成员取不到 uid", member: "   ", want: ""},
		{name: "空 json 取不到 uid", member: `{"uid":"","retryCount":0}`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uidOfMember(tt.member); got != tt.want {
				t.Errorf("uidOfMember(%q) = %q, want %q", tt.member, got, tt.want)
			}
		})
	}
}

func Test_staleEnoughToFire(t *testing.T) {
	const ttl = int64(15)
	now := int64(1700000000000)

	tests := []struct {
		name       string
		lastPingAt int64
		want       bool
	}{
		{name: "刚好到 TTL 边界, 判下线", lastPingAt: now - ttl*1000, want: true},
		{name: "超过 TTL, 判下线", lastPingAt: now - (ttl+5)*1000, want: true},
		{name: "还差 1 毫秒到 TTL, 不判下线", lastPingAt: now - ttl*1000 + 1, want: false},
		{name: "刚刚心跳过, 不判下线", lastPingAt: now - 1000, want: false},
		{name: "心跳时间在未来(时钟漂移), 不判下线", lastPingAt: now + 1000, want: false},
		{name: "队列里查不到该 uid 的心跳, 判下线", lastPingAt: 0, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := staleEnoughToFire(tt.lastPingAt, now, ttl); got != tt.want {
				t.Errorf("staleEnoughToFire(%d, %d, %d) = %v, want %v", tt.lastPingAt, now, ttl, got, tt.want)
			}
		})
	}
}

// 复现线上故障：心跳队列里留下一条 retryCount>0 的老副本，副本分数不再被真实心跳刷新，
// 老化后必须被「新鲜度回查」拦住，不能把还在心跳的 uid 判成下线；
// 等真的不再心跳时，仍要照常判下线。
// 需要真实 redis，默认跳过：PING_TEST_REDIS_ADDR=host:port [PING_TEST_REDIS_USER/PASSWORD/PREFIX] go test -run Test_offlineNotFiredForLiveUid
func Test_offlineNotFiredForLiveUid(t *testing.T) {
	addr := os.Getenv("PING_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set PING_TEST_REDIS_ADDR to run this integration test")
	}

	ctx := context.Background()
	rc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{addr},
		Username: os.Getenv("PING_TEST_REDIS_USER"),
		Password: os.Getenv("PING_TEST_REDIS_PASSWORD"),
	})
	if err := rc.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}

	prefix := os.Getenv("PING_TEST_REDIS_PREFIX")
	if prefix == "" {
		prefix = fmt.Sprintf("ping-it-%d", time.Now().UnixNano())
	}

	offlineL := make(chan string, 16)
	svr := NewService(rc, &Option{Prefix: prefix, OfflineTTL: 4})
	svr.RegisterOfflineHandle(func(uid string) error {
		offlineL <- uid
		return nil
	})
	svr.StartTick()

	uid := fmt.Sprintf("it-%d", time.Now().UnixNano())
	queueKey := fmt.Sprintf("%s:pingService", prefix)
	unAckKey := fmt.Sprintf("%s:pingService:unAck", prefix)
	defer func() {
		rc.Del(ctx, queueKey, unAckKey)
		rc.Del(ctx, fmt.Sprintf("%s:pingService:lock:%s", prefix, uid))
	}()

	staleCopy := `{"uid":"` + uid + `","retryCount":1}`
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := svr.Ping(ctx, uid); err != nil {
					t.Errorf("Ping failed: %v", err)
					return
				}
				if err := rc.ZAdd(ctx, queueKey, redis.Z{
					Score:  float64(time.Now().Add(-time.Hour).UnixMilli()),
					Member: staleCopy,
				}).Err(); err != nil {
					t.Errorf("seed stale copy failed: %v", err)
					return
				}
			}
		}
	}()

	time.Sleep(20 * time.Second)

	if n := countOffline(offlineL, uid, time.Second); n != 0 {
		close(stop)
		<-done
		t.Fatalf("还在心跳的 uid 被判成下线 %d 次, 判定: 新鲜度回查没有拦住过期副本", n)
	}

	close(stop)
	<-done

	if countOffline(offlineL, uid, 40*time.Second) == 0 {
		t.Fatalf("停止心跳后 40s 内没有判下线, 判定: 正常下线路径被改坏了")
	}

	left, err := rc.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		t.Fatalf("read queue failed: %v", err)
	}
	for _, m := range left {
		if uidOfMember(m) == uid {
			t.Fatalf("判下线后队列里还留着该 uid 的成员: %q", m)
		}
	}
}

// countOffline 在 window 时间内统计该 uid 被判下线的次数。
func countOffline(ch <-chan string, uid string, window time.Duration) int {
	n := 0
	deadline := time.After(window)

	for {
		select {
		case got := <-ch:
			if got == uid {
				n++
			}
		case <-deadline:
			return n
		}
	}
}
