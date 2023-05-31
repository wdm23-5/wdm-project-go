package order

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client: *rdb,
	}
}
