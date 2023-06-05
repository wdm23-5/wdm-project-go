package stock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsIncrByIfGe0XX     *redis.Script
	rsPrepareCkTx       *redis.Script
	rsPrepareCkTxMove   *redis.Script
	rsAcknowledgeCkTx   *redis.Script
	rsCommitCkTx        *redis.Script
	rsAbortCkTx         *redis.Script
	rsAbortCkTxRollback *redis.Script
}

func newRedisDB(addr, password string, db int) *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisDB{
		Client:              *rdb,
		rsIncrByIfGe0XX:     redis.NewScript(common.LuaIncrByIfGe0XX),
		rsPrepareCkTx:       redis.NewScript(luaPrepareCkTx),
		rsPrepareCkTxMove:   redis.NewScript(luaPrepareCkTxMove),
		rsAcknowledgeCkTx:   redis.NewScript(luaAcknowledgeCkTx),
		rsCommitCkTx:        redis.NewScript(luaCommitCkTx),
		rsAbortCkTx:         redis.NewScript(luaAbortCkTx),
		rsAbortCkTxRollback: redis.NewScript(luaAbortCkTxRollback),
	}
}

func (rdb *redisDB) CacheAllScripts(ctx context.Context) error {
	pipe := rdb.Pipeline()
	rdb.rsIncrByIfGe0XX.Load(ctx, pipe)
	rdb.rsPrepareCkTx.Load(ctx, pipe)
	rdb.rsPrepareCkTxMove.Load(ctx, pipe)
	rdb.rsAcknowledgeCkTx.Load(ctx, pipe)
	rdb.rsCommitCkTx.Load(ctx, pipe)
	rdb.rsAbortCkTx.Load(ctx, pipe)
	rdb.rsAbortCkTxRollback.Load(ctx, pipe)
	_, err := pipe.Exec(ctx)
	return err
}

func (rdb *redisDB) IncrByIfGe0XX(ctx context.Context, key string, value int) *redis.Cmd {
	return rdb.rsIncrByIfGe0XX.Run(ctx, rdb.Client, []string{key}, value)
}

func (rdb *redisDB) PrepareCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsPrepareCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		string(common.TxPreparing), "price",
	)
}

func (rdb *redisDB) PrepareCkTxMove(ctx context.Context, txId, itemId string, amount int) *redis.Cmd {
	return rdb.rsPrepareCkTxMove.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), keyStock(itemId), keyPrice(itemId), common.KeyTxLocked(txId)},
		string(common.TxPreparing), amount, "item_"+itemId, "price",
	)
}

func (rdb *redisDB) AcknowledgeCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsAcknowledgeCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		string(common.TxPreparing), string(common.TxAcknowledged), "price",
	)
}

func (rdb *redisDB) CommitCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsCommitCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		string(common.TxAcknowledged), string(common.TxCommitted), "price",
	)
}

func (rdb *redisDB) AbortCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsAbortCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		string(common.TxPreparing), string(common.TxAcknowledged), string(common.TxAborted), "price",
	)
}

func (rdb *redisDB) AppendAbortCkTxRollback(ctx context.Context, pipe redis.Pipeliner, txId, itemId string) *redis.Cmd {
	return rdb.rsAbortCkTxRollback.Run(
		ctx, pipe,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId), keyStock(itemId)},
		string(common.TxAborted), "item_"+itemId,
	)
}
