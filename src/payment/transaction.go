package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"wdm/common"
)

// todo:
//  record tx state due to at-least-one
//  save tx data locally

func prepareCkTx(ctx *gin.Context) {
	var req common.CreditTxPrpRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "prepareCkTx: %v", err)
		return
	}

	if _, err := rdb.IncrByIfGe0XX(ctx, keyCredit(req.Payer.Id), -req.Payer.Amount).Result(); err == redis.Nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "prepareCkTx: %v", err)
		return
	}

	ctx.Status(http.StatusOK)
}

func commitCkTx(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func abortCkTx(ctx *gin.Context) {
	var req common.CreditTxPrpRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "abortCkTx: %v", err)
		return
	}

	if _, err := rdb.IncrByIfGe0XX(ctx, keyCredit(req.Payer.Id), req.Payer.Amount).Result(); err == redis.Nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "abortCkTx: %v", err)
		return
	}

	ctx.Status(http.StatusOK)
}
