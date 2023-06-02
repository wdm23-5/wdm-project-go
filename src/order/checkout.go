package order

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func checkoutOrder(ctx *gin.Context) {
	// todo: use message queue to limit rate

	orderId := ctx.Param("order_id")
	txId := snowGen.Next().String()

	locked, info, err := prepareCkTxLocal(ctx, txId, orderId)
	if err == redis.Nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "checkoutOrder: %v", err)
		return
	}

	if !locked {
		if info.paid {
			// already paid
			ctx.Status(http.StatusOK)
		} else {
			// concurrent checkout on same order
			ctx.Status(http.StatusTooManyRequests)
		}
		return
	}

	// tx preparing

	// empty cart, commit directly
	if len(info.cart) == 0 {
		_, err = rdb.CommitCkTx(ctx, txId, orderId).Result()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "checkoutOrder: CommitCkTx: %v", err)
		} else {
			ctx.Status(http.StatusOK)
		}
		return
	}

	// ask stock
	price, err := prepareCkTxStock(txId, info)
	if err != nil {
		ctx.String(http.StatusBadRequest, "checkoutOrder: %v", err)
		// todo: abort
		return
	}

	// ask payment
	err = prepareCkTxPayment(txId, info.userId, price)
	if err != nil {
		ctx.String(http.StatusBadRequest, "checkoutOrder: %v", err)
		// todo: abort
		return
	}

	_, err = rdb.AcknowledgeCkTx(ctx, txId, orderId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "checkoutOrder: AcknowledgeCkTx: %v", err)
		return
	}

	// tx acked, start committing

	// todo: use message queue to guarantee delivery. return directly after message queue ok
	go commitCkTxRemote(txId, info)

	_, err = rdb.CommitCkTx(ctx, txId, orderId).Result()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "checkoutOrder: CommitCkTx: %v", err)
	} else {
		ctx.Status(http.StatusOK)
	}
}
