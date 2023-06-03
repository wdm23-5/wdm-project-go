package stock

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wdm/common"
)

// PRP -> ABT
//  v
// ACK -> ABT
//  v
// CMT

func prepareCkTx(ctx *gin.Context) {
	var req common.ItemTxPrpRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "prepareCkTx: %v", err)
		return
	}

	val, err := rdb.PrepareCkTx(ctx, req.TxId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: %v", err)
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
	for key := common.KeyTxState(req.TxId); state == common.TxPreparing; {
		time.Sleep(10 * time.Millisecond)
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "prepareCkTx: wait %v", err)
			return
		}
		state = common.TxState(val)
	}
	switch state {
	case "":
		// new tx
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
		amount := req.Items[i].Amount
		val, err := rdb.PrepareCkTxMove(ctx, req.TxId, itemId, amount).Result()
		if err == redis.Nil {
			// out of stock
			// todo
			abortCkTx(ctx)
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
			ctx.String(http.StatusBadRequest, "prepareCkTx: PrepareCkTxMove abort")
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
			ctx.String(http.StatusBadRequest, "prepareCkTx: AcknowledgeCkTx abort")
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
	default:
		ctx.String(http.StatusInternalServerError, "commitCkTx: invalid state %v", state)
	}
}

func abortCkTx(ctx *gin.Context) {
	txId := ctx.Param("tx_id")

	val, err := rdb.AbortCkTx(ctx, txId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: %v", err)
		return
	}

	stateStr, ok := val.(string)
	if !ok {
		ctx.String(http.StatusInternalServerError, "abortCkTx: not string %v", val)
		return
	}
	switch state := common.TxState(stateStr); state {
	case common.TxPreparing, common.TxAcknowledged:
		// roll back
		go func() {
			ctx := context.Background()
			keys := make([]string, 0)
			cursor := uint64(0)
			for cursor != 0 {
				keys, cursor, err = rdb.HScan(ctx, common.KeyTxLocked(txId), cursor, "", 0).Result()
				if err != nil {
					return
				}
				pipe := rdb.Pipeline()
				for i := 0; i < len(keys); i += 2 {
					itemId := strings.TrimPrefix(keys[i], "item_")
					rdb.AppendAbortCkTxRollback(ctx, pipe, txId, itemId)
				}
				_, err := pipe.Exec(ctx)
				if err != nil {
					return
				}
			}
		}()
		ctx.Status(http.StatusOK)
	case common.TxAborted:
		ctx.Status(http.StatusOK)
	default:
		ctx.String(http.StatusInternalServerError, "abortCkTx: invalid state %v", state)
	}
}
