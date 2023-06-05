package order

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func checkoutOrder(ctx *gin.Context) {
	// todo: use message queue to limit rate

	orderId := ctx.Param("order_id")

	// NB: In fact, the machine id part of a tx id should be the id of the issuing service which
	// is already true in our implementation. If, however, future developers decide to decouple
	// the machine id of the generator and that of the service (which is how it should be), then
	// this line of code should be changed accordingly.
	txId := snowGen.Next().String()

	// In current impl, we can only do orderId -> txId but not
	// the reverse unless we scan the whole sharded database.
	// In fact, we can save the information in KeyTxLocked since
	// it is not used yet.
	// todo: save txId -> (orderId, userId)
	rdb := srdb.Route(orderId)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "checkoutOrder: error shard key %v", orderId)
		return
	}

	locked, info, err := prepareCkTxLocal(ctx, rdb, txId, orderId)
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
	price, err := prepareCkTxStock(txId, info.cart)
	if err != nil {
		strictConsistency := true
		//goland:noinspection GoBoolExpressions
		if strictConsistency {
			abortCkTxStock(txId, info.cart)
		} else {
			// todo: use message queue
			go abortCkTxStock(txId, info.cart)
		}
		_, errA := rdb.AbortCkTx(ctx, txId, orderId).Result()
		if errA != nil {
			ctx.String(http.StatusInternalServerError, "checkoutOrder: AbortCkTx: %v; %v", errA, err)
		} else {
			ctx.String(http.StatusBadRequest, "checkoutOrder: %v", err)
		}
		return
	}

	// ask payment
	err = prepareCkTxPayment(txId, info.userId, price)
	if err != nil {
		strictConsistency := true
		//goland:noinspection GoBoolExpressions
		if strictConsistency {
			abortCkTxPayment(txId, info.userId)
			abortCkTxStock(txId, info.cart)
		} else {
			// todo: use message queue
			go abortCkTxPayment(txId, info.userId)
			go abortCkTxStock(txId, info.cart)
		}
		_, errA := rdb.AbortCkTx(ctx, txId, orderId).Result()
		if errA != nil {
			ctx.String(http.StatusInternalServerError, "checkoutOrder: AbortCkTx: %v; %v", errA, err)
		} else {
			ctx.String(http.StatusNotAcceptable, "checkoutOrder: %v", err)
		}
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
