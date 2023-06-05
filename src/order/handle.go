package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wdm/common"
)

func createOrder(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	orderId := snowGen.Next().String()

	rdb := srdb.Route(orderId)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "createOrder: error shard key %v", orderId)
		return
	}

	pipe := rdb.TxPipeline()
	pipe.Set(ctx, keyUserId(orderId), userId, 0)
	pipe.Set(ctx, keyPaid(orderId), 0, 0)
	_, err := pipe.Exec(ctx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "createOrder: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, common.CreateOrderResponse{OrderId: orderId})
}

func removeOrder(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	rdb := srdb.Route(orderId)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "removeOrder: error shard key %v", orderId)
		return
	}
	rdb.Del(ctx, keyUserId(orderId), keyPaid(orderId), keyCart(orderId))
	ctx.Status(http.StatusOK)
}

func addItem(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	itemId := ctx.Param("item_id")
	rdb := srdb.Route(orderId)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "addItem: error shard key %v", orderId)
		return
	}
	rdb.HIncrBy(ctx, keyCart(orderId), itemId, 1)
	ctx.Status(http.StatusOK)
}

func removeItem(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	itemId := ctx.Param("item_id")
	rdb := srdb.Route(orderId)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "removeItem: error shard key %v", orderId)
		return
	}
	rdb.HDecrIfGe0XX(ctx, keyCart(orderId), itemId)
	ctx.Status(http.StatusOK)
}
