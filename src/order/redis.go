package order

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	rdb *redis.Client
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{rdb: rdb}
}

func (rdb *redisDB) ping(ctx context.Context) *redis.StatusCmd {
	return rdb.rdb.Ping(ctx)
}

func (rdb *redisDB) flushDB(ctx context.Context) {
	rdb.rdb.FlushDB(ctx)
}
