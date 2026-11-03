package ping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// 下线判定与真正执行之间存在时间差（tick 每秒选一次，handler 是异步 goroutine，
// 重试副本也会在队列里躺着），所以把成员挪进 unAck 时给一个宽限期：
// 这段时间内 tickUnAck 不会把它当「没人处理」重新塞回心跳队列，
// 否则同一个 uid 会在 handler 还在跑的时候被重复排队、并派生出新的队列成员。
const unAckGrace = 30 * time.Second

func (s *Service) handleOffline(member string) {
	ctx := context.WithValue(context.Background(), "opId", strings.ReplaceAll(uuid.New().String(), "-", ""))
	lg := CtxLogger(ctx).WithFields(logrus.Fields{"member": member})

	uid := uidOfMember(member)
	if uid == "" {
		lg.Errorf("handle offline error: 队列成员取不到 uid, 判定: 直接丢弃这条成员, member: %s", member)
		s.redis.ZRem(ctx, s.pingQK, member)
		return
	}

	lastPingAt, err := s.redis.ZScore(ctx, s.pingQK, uid).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		lg.Errorf("handle offline error: 心跳新鲜度回查失败, 判定: 本轮不判下线, uid: %s, err: %v", uid, err)
		return
	}

	if !staleEnoughToFire(int64(lastPingAt), time.Now().UnixMilli(), s.offlineTTL) {
		drop := s.redis.Pipeline()
		drop.ZRem(ctx, s.pingQK, member)
		drop.ZRem(ctx, s.pingUnAckQK, member)
		if _, errDrop := drop.Exec(ctx); errDrop != nil {
			lg.Errorf("handle offline error: 过期成员清理失败, uid: %s, err: %v", uid, errDrop)
		}

		lg.Infof("skip, 该 uid 还在心跳, 判断为队列里的过期副本, 丢弃且不判下线, uid: %s, lastPingAt: %d, offlineTTL: %d",
			uid, int64(lastPingAt), s.offlineTTL)

		return
	}

	pipe := s.redis.Pipeline()
	pipe.ZRem(ctx, s.pingQK, member)
	pipe.ZAdd(ctx, s.pingUnAckQK, redis.Z{Score: float64(time.Now().Add(unAckGrace).UnixMilli()), Member: member})
	_, err = pipe.Exec(ctx)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}

	LT := 3 * time.Minute
	ob, err := s.lock.Obtain(ctx, s.lockKey(uid), LT, nil)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			lg.Infof("未获取到锁, 判断为重复执行, uid: %s", uid)
			return
		}
		lg.Errorf("obtain lock failed, 不执行, err: %s, uid: %s", err, uid)
		return
	}
	// 锁按 uid 维度：处理完必须释放，否则该 uid 在锁 TTL 内再次真下线会被吞掉。
	defer ob.Release(ctx)

	err = s.offlineHandle(uid)
	if err != nil {
		if errors.Is(err, ErrRetry) {
			s.requeueRetry(ctx, lg, member, uid)
			return
		}
		lg.Errorf("handle offline error: %v", err)
		return
	}

	p1 := s.redis.Pipeline()
	p1.ZRem(ctx, s.pingQK, member)
	p1.ZRem(ctx, s.pingQK, uid)
	p1.ZRem(ctx, s.pingUnAckQK, member)
	_, err = p1.Exec(ctx)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}
}

func (s *Service) requeueRetry(ctx context.Context, lg *logrus.Entry, member string, uid string) {
	e1 := new(Event)

	if err := json.Unmarshal([]byte(member), e1); err != nil || e1.Uid == "" {
		e1 = &Event{Uid: uid}
	}

	if e1.RetryCount > 3 {
		p := s.redis.Pipeline()
		p.ZRem(ctx, s.pingQK, member)
		p.ZRem(ctx, s.pingUnAckQK, member)
		if _, err := p.Exec(ctx); err != nil {
			lg.Errorf("handle offline error: 超重试次数后清理失败, uid: %s, err: %v", uid, err)
		}

		lg.Errorf("handle offline error: retry count exceed, uid: %s", uid)

		return
	}

	e1.RetryCount++

	e1B, err := json.Marshal(e1)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}

	if err := s.redis.ZRem(ctx, s.pingUnAckQK, member).Err(); err != nil {
		lg.Errorf("handle offline error: 重试前清理 unAck 失败, uid: %s, err: %v", uid, err)
	}

	score := float64(time.Now().Add(time.Duration(e1.RetryCount) * time.Second).UnixMilli())
	if err := s.redis.ZAdd(ctx, s.pingQK, redis.Z{Score: score, Member: string(e1B)}).Err(); err != nil {
		lg.Errorf("handle offline error: 重试排队失败, uid: %s, err: %v", uid, err)
	}

	lg.Infof("offline handle 要求重试, uid: %s, retryCount: %d", uid, e1.RetryCount)
}

func (s *Service) lockKey(uid string) string {
	return fmt.Sprintf("%s:pingService:lock:%s", s.Prefix, uid)
}
