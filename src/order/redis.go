package order

import (
	"context"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	*redis.Client
	rsHDecrIfGe0XX    *redis.Script
	rsPrepareCkTx     *redis.Script
	rsAcknowledgeCkTx *redis.Script
	rsCommitCkTx      *redis.Script
	rsAbortCkTx       *redis.Script
}

func newRedisDB(addr, password string, db int) *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisDB{
		Client:            rdb,
		rsHDecrIfGe0XX:    redis.NewScript(luaHDecrIfGe0XX),
		rsPrepareCkTx:     redis.NewScript(luaPrepareCkTx),
		rsAcknowledgeCkTx: redis.NewScript(luaAcknowledgeCkTx),
		rsCommitCkTx:      redis.NewScript(luaCommitCkTx),
		rsAbortCkTx:       redis.NewScript(luaAbortCkTx),
	}
}

func (rdb *redisDB) CacheAllScripts(ctx context.Context) error {
	pipe := rdb.Pipeline()
	rdb.rsHDecrIfGe0XX.Load(ctx, pipe)
	rdb.rsPrepareCkTx.Load(ctx, pipe)
	rdb.rsAcknowledgeCkTx.Load(ctx, pipe)
	rdb.rsCommitCkTx.Load(ctx, pipe)
	rdb.rsAbortCkTx.Load(ctx, pipe)
	_, err := pipe.Exec(ctx)
	return err
}

func (rdb *redisDB) HDecrIfGe0XX(ctx context.Context, key, field string) *redis.Cmd {
	return rdb.rsHDecrIfGe0XX.Run(ctx, rdb.Client, []string{key}, field)
}

func (rdb *redisDB) PrepareCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsPrepareCkTx.Run(
		ctx, rdb.Client,
		[]string{keyUserId(orderId), keyPaid(orderId), keyCart(orderId), keyCkTxId(orderId), common.KeyTxState(txId)},
		txId, string(common.TxPreparing),
	)
}

func (rdb *redisDB) AcknowledgeCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsAcknowledgeCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTxId(orderId), common.KeyTxState(txId)},
		txId, string(common.TxAcknowledged),
	)
}

func (rdb *redisDB) CommitCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsCommitCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTxId(orderId), common.KeyTxState(txId), keyPaid(orderId)},
		txId, string(common.TxCommitted),
	)
}

func (rdb *redisDB) AbortCkTx(ctx context.Context, txId, orderId string) *redis.Cmd {
	return rdb.rsAbortCkTx.Run(
		ctx, rdb.Client,
		[]string{keyCkTxId(orderId), common.KeyTxState(txId)},
		txId, string(common.TxAborted),
	)
}
