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

func (s *Service) handleOffline(eventS string) {
	ctx := context.WithValue(context.Background(), "opId", strings.ReplaceAll(uuid.New().String(), "-", ""))
	lg := CtxLogger(ctx).WithFields(logrus.Fields{"event": eventS})

	pipe := s.redis.Pipeline()
	pipe.ZRem(ctx, s.pingQK, eventS)
	pipe.ZAdd(ctx, s.pingUnAckQK, redis.Z{Score: float64(time.Now().UnixMilli()), Member: eventS})
	_, err := pipe.Exec(ctx)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}

	e := new(Event)
	err = json.Unmarshal([]byte(eventS), e)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}

	LT := 3 * time.Minute
	_, err = s.lock.Obtain(ctx, fmt.Sprintf("%s:pingService:lock:%s", s.Prefix, eventS), LT, nil)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			lg.Infof("未获取到锁, 判断为重复执行, event: %s", eventS)
			return
		}
		lg.Errorf("obtain lock failed, 不执行, err: %s, event: %s", err, eventS)
		return
	}

	err = s.offlineHandle(e)
	if err != nil {
		if errors.Is(err, ErrRetry) {
			e1 := new(Event)
			err = json.Unmarshal([]byte(eventS), e1)
			if err != nil {
				lg.Errorf("handle offline error: %v", err)
				return
			}

			if e1.RetryCount > 3 {
				s.redis.ZRem(ctx, s.pingQK, eventS)
				s.redis.ZRem(ctx, s.pingUnAckQK, eventS)
				lg.Errorf("handle offline error: retry count exceed")
				return
			}

			e1.RetryCount++
			e1B, err := json.Marshal(e1)
			if err != nil {
				lg.Errorf("handle offline error: %v", err)
				return
			}

			s.redis.ZAdd(ctx, s.pingQK, redis.Z{Score: float64(time.Now().Add(time.Duration(e1.RetryCount*1) * time.Second).UnixMilli()), Member: string(e1B)})
			return
		}
		lg.Errorf("handle offline error: %v", err)
		return
	}

	p1 := s.redis.Pipeline()
	p1.ZRem(ctx, s.pingQK, eventS)
	p1.ZRem(ctx, s.pingUnAckQK, eventS)
	_, err = p1.Exec(ctx)
	if err != nil {
		lg.Errorf("handle offline error: %v", err)
		return
	}
}
