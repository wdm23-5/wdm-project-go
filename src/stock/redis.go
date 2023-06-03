package stock

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsIncrByIfGe0XX   *redis.Script
	rsPrepareCkTx     *redis.Script
	rsPrepareCkTxMove *redis.Script
	rsCommitCkTx      *redis.Script
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client:            *rdb,
		rsIncrByIfGe0XX:   redis.NewScript(common.LuaIncrByIfGe0XX),
		rsPrepareCkTx:     redis.NewScript(luaPrepareCkTx),
		rsPrepareCkTxMove: redis.NewScript(luaPrepareCkTxMove),
		rsCommitCkTx:      redis.NewScript(luaCommitCkTx),
	}
}

func (rdb *redisDB) IncrByIfGe0XX(ctx context.Context, key string, value int) *redis.Cmd {
	return rdb.rsIncrByIfGe0XX.Run(ctx, rdb.Client, []string{key}, value)
}

func (rdb *redisDB) PrepareCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsPrepareCkTxMove.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		common.TxPreparing,
	)
}

func (rdb *redisDB) PrepareCkTxMove(ctx context.Context, txId, itemId string, amount int) *redis.Cmd {
	return rdb.rsPrepareCkTxMove.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), keyStock(itemId), keyPrice(itemId), common.KeyTxLocked(txId)},
		common.TxPreparing, amount, "item_"+itemId, "price",
	)
}

func (rdb *redisDB) CommitCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsPrepareCkTxMove.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		common.TxAcknowledged, common.TxCommitted, "price",
	)
}
