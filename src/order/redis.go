package order

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsHDecrIfGe0XX    *redis.Script
	rsPrepareCkTx     *redis.Script
	rsAcknowledgeCkTx *redis.Script
	rsCommitCkTx      *redis.Script
	rsAbortCkTx       *redis.Script
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client:            *rdb,
		rsHDecrIfGe0XX:    redis.NewScript(luaHDecrIfGe0XX),
		rsPrepareCkTx:     redis.NewScript(luaPrepareCkTx),
		rsAcknowledgeCkTx: redis.NewScript(luaAcknowledgeCkTx),
		rsCommitCkTx:      redis.NewScript(luaCommitCkTx),
		rsAbortCkTx:       redis.NewScript(luaAbortCkTx),
	}
}

func (rdb *redisDB) HDecrIfGe0XX(ctx context.Context, key, field string) *redis.Cmd {
	return rdb.rsHDecrIfGe0XX.Run(ctx, rdb.Client, []string{key}, field)
}

func (rdb *redisDB) PrepareCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsPrepareCkTx.Run(
		ctx, rdb.Client,
		[]string{keyUserId(orderId), keyPaid(orderId), keyCart(orderId), keyCkTx(orderId), common.KeyTxState(txId)},
		txId, common.TxPreparing,
	)
}

func (rdb *redisDB) AcknowledgeCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsAcknowledgeCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTx(orderId), common.KeyTxState(txId)},
		txId, common.TxAcknowledged,
	)
}

func (rdb *redisDB) CommitCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsCommitCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTx(orderId), common.KeyTxState(txId), keyPaid(orderId)},
		txId, common.TxCommitted,
	)
}

func (rdb *redisDB) AbortCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsAbortCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTx(orderId), common.KeyTxState(txId)},
		txId, common.TxAborted,
	)
}
