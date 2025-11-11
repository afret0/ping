package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (s *Service) startTick() {
	lg := GetLogger()
	lg.Infof("start tick")
	defer lg.Infof("stop tick")

	for range time.Tick(time.Second) {
		if s.debug() {
			lg.Infof("tick")
		}

		go s.tick()
		go s.tickUnAck()
	}
}

func (s *Service) tick() {
	EL, err := s.redis.ZRangeByScore(context.Background(), s.pingQK, &redis.ZRangeBy{Min: "-inf", Max: fmt.Sprintf("%d", time.Now().Add(-time.Duration(s.offlineTTL)*time.Second).UnixMilli())}).Result()
	if err != nil {
		GetLogger().Errorf("tick error: %v", err)
		return
	}

	for _, e := range EL {
		e := e
		go s.handleOffline(e)
	}
}

func (s *Service) tickUnAck() {
	EL, err := s.redis.ZRevRangeByScore(context.Background(), s.pingUnAckQK, &redis.ZRangeBy{Min: "-inf", Max: fmt.Sprintf("%d", time.Now().UnixMilli())}).Result()
	if err != nil {
		GetLogger().Errorf("tick unAck error: %v", err)
		return
	}

	for _, e := range EL {
		e := e
		ctx := context.WithValue(context.Background(), "opId", strings.ReplaceAll(uuid.New().String(), "-", ""))

		E := new(Event)
		err = json.Unmarshal([]byte(e), E)
		if err != nil {
			GetLogger().Errorf("tick unAck error: %v", err)
			continue
		}

		if E.RetryCount >= 3 {
			s.redis.ZRem(ctx, s.pingUnAckQK, e)
			continue
		}

		E.RetryCount++
		EB, err := json.Marshal(E)
		if err != nil {
			GetLogger().Errorf("tick unAck error: %v", err)
			continue
		}

		s.redis.ZAdd(ctx, s.pingQK, redis.Z{Score: float64(time.Now().UnixMilli()), Member: string(EB)})
		s.redis.ZRem(ctx, s.pingUnAckQK, e)
	}
}
