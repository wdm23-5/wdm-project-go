package stock

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"time"
	"wdm/common"
)

// todo:
//  record tx state due to at-least-one
//  save tx data locally

// PRP -> ABT
//  v
// ACK -> ABT
//  v
// CMT

func prepareCkTx(ctx *gin.Context) {
	var req common.ItemTxPrpAbtRequest
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
			continue
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
			ctx.String(http.StatusBadRequest, "prepareCkTx: abort")
			return
		default:
			ctx.String(http.StatusInternalServerError, "prepareCkTx: move invalid state %v", state)
			return
		}
	}

	// all pass, ack
	// todo: change state
	priceStr, err := rdb.HGet(ctx, common.KeyTxLocked(req.TxId), "price").Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: ack HGET %v", err)
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: ack atoi %v", err)
		return
	}
	ctx.JSON(http.StatusOK, common.ItemTxPrpResponse{TotalCost: price})
}

func commitCkTx(ctx *gin.Context) {
	txId := ctx.Param("tx_id")

	val, err := rdb.CommitCkTx(ctx, txId).Result()
	if err == redis.Nil {
		ctx.Status(http.StatusNotFound)
		return
	}
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
	// todo: roll back
	ctx.String(http.StatusTeapot, "abortCkTx")
}
