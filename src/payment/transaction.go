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

	if txId := ctx.Param("tx_id"); txId != req.TxId {
		ctx.String(http.StatusBadRequest, "prepareCkTx: txId mismatch")
		return
	}

	shardKey := ctx.Param("shard_key")
	// For now, the shard key is userId.
	// If future developer leverage other algo than
	// identity mapping, this check should also be
	// updated. So do the other checks under /tx.
	if shardKey != req.Payer.Id {
		ctx.String(http.StatusInternalServerError, "abortCkTx: shardKey != req.Payer.Id")
		return
	}

	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "prepareCkTx: error shard key %v", shardKey)
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

	// only one field, should be fast
	fieldValues, err := rdb.HGetAll(ctx, common.KeyTxLocked(txId)).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: HGETALL %v", err)
		return
	}
	if lfv := len(fieldValues); lfv == 0 {
		// todo: atomic check and set
		rdb.Set(ctx, common.KeyTxState(txId), common.TxAborted, 0)
		ctx.String(http.StatusOK, "abortCkTx: HGETALL length zero")
		return
	} else if lfv != 1 {
		ctx.String(http.StatusInternalServerError, "abortCkTx: HGETALL length error")
		return
	}
	userId := ""
	// *typical go*
	for userId = range fieldValues {
		break
	}

	if shardKey != userId {
		ctx.String(http.StatusInternalServerError, "abortCkTx: shardKey != userId")
		return
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
