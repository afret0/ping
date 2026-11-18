package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ErrRetry = fmt.Errorf("retry")

type Option struct {
	Prefix     string
	OfflineTTL int64
	//OfflineHandle func(uid string) error
}

type Service struct {
	pingQK        string
	pingUnAckQK   string
	redis         redis.UniversalClient
	offlineHandle func(uid string) error
	offlineTTL    int64
	Prefix        string

	lock *redislock.Client
}

type Event struct {
	//Id         string `json:"id"`
	Uid        string `json:"uid"`
	RetryCount int    `json:"retryCount"`
}

func (e *Event) Marshall() []byte {
	b, _ := json.Marshal(e)
	return b
}

func NewService(redisClient redis.UniversalClient, opt *Option) *Service {
	if opt.Prefix == "" {
		panic("prefix is required")
	}
	if opt.OfflineTTL <= 3 {
		panic("offline ttl must be greater than 3")
	}
	//if opt.OfflineHandle == nil {
	//	panic("offline handle is required")
	//}

	svr := &Service{
		pingQK:      fmt.Sprintf("%s:pingService", opt.Prefix),
		pingUnAckQK: fmt.Sprintf("%s:pingService:unAck", opt.Prefix),
		redis:       redisClient,
		offlineTTL:  opt.OfflineTTL,
		Prefix:      opt.Prefix,
		lock:        redislock.New(redisClient),
		//offlineHandle: opt.OfflineHandle,
	}

	return svr
}

// 心跳队列的成员就是 uid：同一个 uid 永远只有一条，每次心跳刷新它的分数。
// 早期版本把 Event(json, 含 retryCount) 当成员，重试路径会为同一个 uid 追加
// retryCount 不同的副本，副本的分数不会被真实心跳刷新，老化后就把还在心跳的
// uid 判成下线（调用方表现为健康连接被踢、客户端重连抖动）。
func (s *Service) Ping(ctx context.Context, uid string) error {
	lg := CtxLogger(ctx).WithFields(logrus.Fields{"uid": uid})

	err := s.redis.ZAdd(ctx, s.pingQK, redis.Z{Score: float64(time.Now().UnixMilli()), Member: uid}).Err()
	if err != nil {
		lg.WithError(err).Error("failed to add ping")
		return err
	}

	return nil
}

func (s *Service) PingMulti(ctx context.Context, uidL []string) error {
	lg := CtxLogger(ctx).WithFields(logrus.Fields{})

	if len(uidL) == 0 {
		return nil
	}

	now := float64(time.Now().UnixMilli())

	pipe := s.redis.Pipeline()
	for _, uid := range uidL {
		pipe.ZAdd(ctx, s.pingQK, redis.Z{Score: now, Member: uid})
	}

	if _, err := pipe.Exec(ctx); err != nil {
		lg.Errorf("fail, 批量心跳写入失败, count: %d, err: %s", len(uidL), err)
		return err
	}

	return nil
}

// uidOfMember 兼容两种成员编码：新版是裸 uid，老版（含重试副本）是 Event 的 json。
func uidOfMember(member string) string {
	if strings.TrimSpace(member) == "" {
		return ""
	}

	if strings.HasPrefix(strings.TrimSpace(member), "{") {
		e := new(Event)
		if err := json.Unmarshal([]byte(member), e); err != nil {
			return ""
		}

		if strings.TrimSpace(e.Uid) == "" {
			return ""
		}

		return e.Uid
	}

	return member
}

// staleEnoughToFire 在「真正要判下线的那一刻」重新核对新鲜度：
// tick 选出来的成员可能已经过期（tick 每秒跑、handler 是异步的、重试副本也在队里），
// 只要该 uid 在 TTL 内还有心跳，就不能判它下线。
func staleEnoughToFire(lastPingAt int64, now int64, offlineTTL int64) bool {
	if lastPingAt <= 0 {
		return true
	}

	return now-lastPingAt >= offlineTTL*int64(time.Second/time.Millisecond)
}

func (s *Service) RegisterOfflineHandle(f func(uid string) error) {
	s.offlineHandle = f
}

func (s *Service) StartTick() {

	if s.offlineHandle == nil {
		panic("offline handle is required")
	}

	go s.startTick()
}

func (s *Service) debug() bool {
	env := os.Getenv("PING_DEBUG")
	if env == "yes" || env == "true" || env == "TRUE" || env == "1" {
		return true
	}
	return false
}
