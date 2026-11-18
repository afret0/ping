package ping

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestPingMulti(t *testing.T) {
	mr := miniredis.RunT(t)
	rc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
	svr := NewService(rc, &Option{Prefix: "test-multi", OfflineTTL: 15})
	ctx := context.Background()

	cases := []struct {
		name  string
		uidL  []string
		wantN int
	}{
		{name: "批量写入三个 uid", uidL: []string{"u1", "u2", "u3"}, wantN: 3},
		{name: "重复成员不追加副本只刷新分数", uidL: []string{"u1", "u2", "u3"}, wantN: 3},
		{name: "空批次直接返回", uidL: nil, wantN: 3},
		{name: "追加新 uid", uidL: []string{"u4"}, wantN: 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := time.Now().UnixMilli()

			if err := svr.PingMulti(ctx, c.uidL); err != nil {
				t.Fatalf("PingMulti err: %s", err)
			}

			members, err := mr.ZMembers("test-multi:pingService")
			if err != nil {
				t.Fatalf("ZMembers err: %s", err)
			}

			if len(members) != c.wantN {
				t.Fatalf("成员数不符, want: %d, got: %d, members: %+v", c.wantN, len(members), members)
			}

			for _, uid := range c.uidL {
				score, err := mr.ZScore("test-multi:pingService", uid)
				if err != nil {
					t.Fatalf("uid 不在心跳队列, uid: %s, err: %s", uid, err)
				}

				if int64(score) < before {
					t.Fatalf("分数没有刷新到本次批量时刻, uid: %s, score: %d, before: %d", uid, int64(score), before)
				}
			}
		})
	}
}

func TestPingMulti与单条Ping同队列(t *testing.T) {
	mr := miniredis.RunT(t)
	rc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
	svr := NewService(rc, &Option{Prefix: "test-mix", OfflineTTL: 15})
	ctx := context.Background()

	if err := svr.Ping(ctx, "u1"); err != nil {
		t.Fatalf("Ping err: %s", err)
	}

	if err := svr.PingMulti(ctx, []string{"u1", "u2"}); err != nil {
		t.Fatalf("PingMulti err: %s", err)
	}

	members, err := mr.ZMembers("test-mix:pingService")
	if err != nil {
		t.Fatalf("ZMembers err: %s", err)
	}

	if len(members) != 2 {
		t.Fatalf("单条与批量必须落同一个队列且同 uid 只有一条, want: 2, got: %d, members: %+v", len(members), members)
	}
}
