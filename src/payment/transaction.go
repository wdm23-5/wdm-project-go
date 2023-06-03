package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"wdm/common"
)

func prepareCkTx(ctx *gin.Context) {
	var req common.CreditTxPrpRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "prepareCkTx: %v", err)
		return
	}

	val, err := rdb.PrpThenAckAbtCkTx(ctx, req.TxId, req.Payer.Id, req.Payer.Amount).Result()
	if err == redis.Nil {
		// not enough credit
		// impl different from stock
		ctx.String(http.StatusNotAcceptable, "prepareCkTx: PrpThenAckAbtCkTx not enough credit")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: PrpThenAckAbtCkTx %v", err)
		return
	}
	stateStr, ok := val.(string)
	if !ok {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: PrpThenAckAbtCkTx not string %v", val)
		return
	}
	state := common.TxState(stateStr)

	// results of multiple calls should be consistent
	switch state {
	case "":
		// new tx, pass
	case common.TxAcknowledged, common.TxCommitted:
		ctx.Status(http.StatusOK)
		return
	case common.TxAborted:
		ctx.Status(http.StatusNotAcceptable)
		return
	default:
		ctx.String(http.StatusInternalServerError, "prepareCkTx: invalid state %v", state)
		return
	}

	// already changed to TxAcknowledged
	ctx.Status(http.StatusOK)
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
		return
	default:
		ctx.String(http.StatusInternalServerError, "commitCkTx: invalid state %v", state)
		return
	}
}

func abortCkTx(ctx *gin.Context) {
	txId := ctx.Param("tx_id")

	// only one field, should be fast
	fieldValues, err := rdb.HGetAll(ctx, common.KeyTxLocked(txId)).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: HGETALL %v", err)
		return
	}
	if len(fieldValues) != 1 {
		ctx.String(http.StatusInternalServerError, "abortCkTx: HGETALL length error")
		return
	}
	userId := ""
	// *typical go*
	for userId = range fieldValues {
		break
	}

	val, err := rdb.AbtThenRollbackCkTx(ctx, txId, userId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: AbtThenRollbackCkTx %v", err)
		return
	}
	stateStr, ok := val.(string)
	if !ok {
		ctx.String(http.StatusInternalServerError, "abortCkTx: AbtThenRollbackCkTx not string %v", val)
		return
	}
	state := common.TxState(stateStr)
	switch state {
	case "", common.TxPreparing, common.TxAborted:
		// already rolled back
		ctx.Status(http.StatusOK)
		return
	default:
		ctx.String(http.StatusInternalServerError, "abortCkTx: AbtThenRollbackCkTx invalid state %v", state)
		return
	}
}
