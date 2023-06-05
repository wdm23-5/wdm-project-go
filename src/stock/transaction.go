package stock

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wdm/common"
)

func prepareCkTx(ctx *gin.Context) {
	var req common.ItemTxPrpRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "prepareCkTx: %v", err)
		return
	}

	if txId := ctx.Param("tx_id"); txId != req.TxId {
		ctx.String(http.StatusBadRequest, "prepareCkTx: txId mismatch")
		return
	}

	shardKey := ctx.Param("shard_key")
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "prepareCkTx: error shard key %v", shardKey)
		return
	}

	val, err := rdb.PrepareCkTx(ctx, req.TxId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTx %v", err)
		return
	}
	stateStr, ok := val.(string)
	if !ok {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTx not string %v", val)
		return
	}
	state := common.TxState(stateStr)

	// results of multiple calls should be consistent
	// if concurrent prepare, wait
	waitedTime := time.Duration(0)
	for key := common.KeyTxState(req.TxId); state == common.TxPreparing; {
		if waitedTime > 60*1000 {
			// 1 min passed
			ctx.String(http.StatusInternalServerError, "prepareCkTx: wait timeout")
			return
		}
		t := common.TxRand3A()
		waitedTime += t
		time.Sleep(t * time.Millisecond)
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: wait %v", err)
			return
		}
		state = common.TxState(val)
	}
	switch state {
	case "":
		// new tx, pass
	case common.TxAcknowledged, common.TxCommitted:
		priceStr, err := rdb.HGet(ctx, common.KeyTxLocked(req.TxId), "price").Result()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: init HGET %v", err)
			return
		}
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: init atoi %v", err)
			return
		}
		ctx.JSON(http.StatusOK, common.ItemTxPrpResponse{TotalCost: price})
		return
	case common.TxAborted:
		ctx.Status(http.StatusNotAcceptable)
		return
	default:
		ctx.String(http.StatusInternalServerError, "prepareCkTx: init invalid state %v", state)
		return
	}

	// new tx, sub stock
	for i := range req.Items {
		itemId := req.Items[i].Id
		// check if all item share the same machineId i.e. shard key
		if common.SnowflakeIDPickMachineIdFast(itemId) != common.SnowflakeIDPickMachineIdFast(shardKey) {
			// just panic, errr
			ctx.String(http.StatusInternalServerError, "prepareCkTx: mId(shardKey) != mId(itemId)")
			return
		}
		amount := req.Items[i].Amount
		val, err := rdb.PrepareCkTxMove(ctx, req.TxId, itemId, amount).Result()
		if err == redis.Nil {
			// out of stock
			err := abtThenRollback(ctx, rdb, req.TxId)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTxMove %v", err)
				return
			}
			ctx.String(http.StatusNotAcceptable, "prepareCkTx: PrepareCkTxMove out of stock")
			return
		}
		if err != nil {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTxMove %v", err)
			return
		}
		stateStr, ok := val.(string)
		if !ok {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTxMove not string %v", val)
			return
		}
		switch state := common.TxState(stateStr); state {
		case common.TxPreparing:
			// normal
			continue
		case common.TxAborted:
			// fast abort, rollback by the abort consumer
			ctx.String(http.StatusNotAcceptable, "prepareCkTx: PrepareCkTxMove abort")
			return
		default:
			ctx.String(http.StatusInternalServerError, "prepareCkTx: PrepareCkTxMove invalid state %v", state)
			return
		}
	}

	// all pass, ack
	val, err = rdb.AcknowledgeCkTx(ctx, req.TxId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx %v", err)
		return
	}
	arr, ok := val.([]interface{})
	if !ok {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx not an array of any")
		return
	}
	if len(arr) != 2 {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx array length error")
		return
	}

	if stateStr, ok := arr[0].(string); !ok {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx array[0] not a string")
		return
	} else {
		switch state := common.TxState(stateStr); state {
		case common.TxPreparing:
			// all good
		case common.TxAborted:
			// fast abort, rollback by the abort consumer
			ctx.String(http.StatusNotAcceptable, "prepareCkTx: AcknowledgeCkTx abort")
			return
		default:
			ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx invalid state %v", state)
			return
		}
	}

	if priceStr, ok := arr[1].(string); !ok {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx array[1] not a string")
		return
	} else if price, err := strconv.Atoi(priceStr); err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: AcknowledgeCkTx array[1] not an int %v", err)
		return
	} else {
		ctx.JSON(http.StatusOK, common.ItemTxPrpResponse{TotalCost: price})
	}
}

func commitCkTx(ctx *gin.Context) {
	txId := ctx.Param("tx_id")
	shardKey := ctx.Param("shard_key")

	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "commitCkTx: error shard key %v", shardKey)
		return
	}

	val, err := rdb.CommitCkTx(ctx, txId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "commitCkTx: %v", err)
		return
	}

	stateStr, ok := val.(string)
	if !ok {
		ctx.String(http.StatusInternalServerError, "commitCkTx: not string %v", val)
		return
	}
	switch state := common.TxState(stateStr); state {
	case common.TxAcknowledged, common.TxCommitted:
		ctx.Status(http.StatusOK)
		return
	default:
		ctx.String(http.StatusInternalServerError, "commitCkTx: invalid state %v", state)
		return
	}
}

func abortCkTx(ctx *gin.Context) {
	txId := ctx.Param("tx_id")
	shardKey := ctx.Param("shard_key")

	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "abortCkTx: error shard key %v", shardKey)
		return
	}

	err := abtThenRollback(ctx, rdb, txId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: %v", err)
		return
	}
	ctx.Status(http.StatusOK)
}

// the error MUST be waited for even in eventual consistency
func abtThenRollback(ctx context.Context, rdb *redisDB, txId string) error {
	val, err := rdb.AbortCkTx(ctx, txId).Result()
	if err != nil {
		return fmt.Errorf("abtThenRollback: %v", err)
	}

	stateStr, ok := val.(string)
	if !ok {
		return fmt.Errorf("abtThenRollback: not string %v", val)
	}
	switch state := common.TxState(stateStr); state {
	case common.TxPreparing, common.TxAcknowledged:
		// roll back
		strictConsistency := true
		//goland:noinspection GoBoolExpressions
		if strictConsistency {
			ch := make(chan string, 1)
			rollbackStock(ctx, rdb, txId, ch)
			errStr := <-ch
			if errStr != "OK" {
				return fmt.Errorf("abtThenRollback: rollbackStock %v", errStr)
			}
			return nil
		}
		go rollbackStock(context.Background(), rdb, txId, nil)
		return nil
	case "", common.TxAborted:
		return nil
	default:
		return fmt.Errorf("abtThenRollback: invalid state %v", state)
	}
}

// called internally by abtThenRollback
func rollbackStock(ctx context.Context, rdb *redisDB, txId string, ch chan string) {
	keys := make([]string, 0)
	cursor := uint64(0)
	var err error
	for {
		keys, cursor, err = rdb.HScan(ctx, common.KeyTxLocked(txId), cursor, "", 0).Result()
		if err != nil {
			if ch != nil {
				ch <- err.Error()
			}
			return
		}
		if len(keys) > 0 {
			pipe := rdb.Pipeline()
			for i := 0; i < len(keys); i += 2 {
				itemId := strings.TrimPrefix(keys[i], "item_")
				rdb.AppendAbortCkTxRollback(ctx, pipe, txId, itemId)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				if ch != nil {
					ch <- err.Error()
				}
				return
			}
		}
		if cursor == 0 {
			break
		}
	}
	if ch != nil {
		ch <- "OK"
	}
}
