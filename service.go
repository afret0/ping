package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ErrRetry = fmt.Errorf("retry")

type Option struct {
	Prefix        string
	OfflineTTL    int64
	OfflineHandle func(uid string) error
}

type Service struct {
	pingQK        string
	pingUnAckQK   string
	redis         redis.UniversalClient
	debug         bool
	offlineHandle func(uid string) error
	offlineTTL    int64
	Prefix        string

	lock *redislock.Client
}

type Event struct {
	Id         string `json:"id"`
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
	if opt.OfflineHandle == nil {
		panic("offline handle is required")
	}

	svr := &Service{
		pingQK:        fmt.Sprintf("%s:pingService", opt.Prefix),
		pingUnAckQK:   fmt.Sprintf("%s:pingService:unAck", opt.Prefix),
		redis:         redisClient,
		offlineTTL:    opt.OfflineTTL,
		Prefix:        opt.Prefix,
		lock:          redislock.New(redisClient),
		offlineHandle: opt.OfflineHandle,
	}

	svr.StartTick()

	return svr
}

func (s *Service) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Service) Ping(ctx context.Context, uid string) error {
	lg := CtxLogger(ctx).WithFields(logrus.Fields{"uid": uid})

	e := &Event{
		Id:  strings.ReplaceAll(uuid.New().String(), "-", ""),
		Uid: uid,
	}

	err := s.redis.ZAdd(ctx, s.pingQK, redis.Z{Score: float64(time.Now().UnixMilli()), Member: string(e.Marshall())}).Err()
	if err != nil {
		lg.WithError(err).Error("failed to add ping")
		return err
	}

	return nil
}

//func (s *Service) RegisterOfflineHandle(f func(e *Event) error) {
//	s.offlineHandle = f
//}

func (s *Service) StartTick() {

	if s.offlineHandle == nil {
		panic("offline handle is required")
	}

	go s.startTick()
}
