package payment

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"wdm/common"
)

type redisDB struct {
	redis.Client
	rsIncrByIfGe0XX       *redis.Script
	rsPrpThenAckAbtCkTx   *redis.Script
	rsCommitCkTx          *redis.Script
	rsAbtThenRollbackCkTx *redis.Script
}

func newRedisDB() *redisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", common.MustGetEnv("REDIS_HOST"), common.MustGetEnv("REDIS_PORT")),
		Password: common.MustGetEnv("REDIS_PASSWORD"),
		DB:       common.MustS2I(common.MustGetEnv("REDIS_DB")),
	})

	return &redisDB{
		Client:                *rdb,
		rsIncrByIfGe0XX:       redis.NewScript(common.LuaIncrByIfGe0XX),
		rsPrpThenAckAbtCkTx:   redis.NewScript(luaPrpThenAckAbtCkTx),
		rsCommitCkTx:          redis.NewScript(luaCommitCkTx),
		rsAbtThenRollbackCkTx: redis.NewScript(luaAbtThenRollbackCkTx),
	}
}

func (rdb *redisDB) IncrByIfGe0XX(ctx context.Context, key string, value int) *redis.Cmd {
	return rdb.rsIncrByIfGe0XX.Run(ctx, rdb.Client, []string{key}, value)
}

func (rdb *redisDB) CacheAllScripts(ctx context.Context) error {
	pipe := rdb.Pipeline()
	rdb.rsIncrByIfGe0XX.Load(ctx, pipe)
	rdb.rsPrpThenAckAbtCkTx.Load(ctx, pipe)
	rdb.rsCommitCkTx.Load(ctx, pipe)
	rdb.rsAbtThenRollbackCkTx.Load(ctx, pipe)
	_, err := pipe.Exec(ctx)
	return err
}

func (rdb *redisDB) PrpThenAckAbtCkTx(ctx context.Context, txId, userId string, amount int) *redis.Cmd {
	return rdb.rsPrpThenAckAbtCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId), keyCredit(userId)},
		string(common.TxAcknowledged), string(common.TxAborted), amount, userId,
	)
}

func (rdb *redisDB) CommitCkTx(ctx context.Context, txId string) *redis.Cmd {
	return rdb.rsCommitCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId)},
		string(common.TxAcknowledged), string(common.TxCommitted),
	)
}

func (rdb *redisDB) AbtThenRollbackCkTx(ctx context.Context, txId, userId string) *redis.Cmd {
	return rdb.rsAbtThenRollbackCkTx.Run(
		ctx, rdb.Client,
		[]string{common.KeyTxState(txId), common.KeyTxLocked(txId), keyCredit(userId)},
		string(common.TxAcknowledged), string(common.TxAborted), userId,
	)
}
