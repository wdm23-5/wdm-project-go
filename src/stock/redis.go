package stock

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsIncrByIfGe0XX *redis.Script
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client:          *rdb,
		rsIncrByIfGe0XX: redis.NewScript(common.LuaIncrByIfGe0XX),
	}
}

func (rdb *redisDB) IncrByIfGe0XX(ctx context.Context, key string, value int) *redis.Cmd {
	return rdb.rsIncrByIfGe0XX.Run(ctx, rdb.Client, []string{key}, value)
}
