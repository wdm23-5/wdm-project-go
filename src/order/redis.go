package order

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsHDecrIfGe0XX *redis.Script
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client:         *rdb,
		rsHDecrIfGe0XX: redis.NewScript(luaHDecrIfGe0XX),
	}
}

func (rdb *redisDB) HDecrIfGe0XX(ctx context.Context, key, field string) *redis.Cmd {
	return rdb.rsHDecrIfGe0XX.Run(ctx, rdb.Client, []string{key}, field)
}
