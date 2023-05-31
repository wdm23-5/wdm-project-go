package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wdm/common"
)

func createOrder(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	orderId := snowGen.Next().String()

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
	rdb.Del(ctx, keyUserId(orderId), keyPaid(orderId), keyCart(orderId))
	ctx.Status(http.StatusOK)
}

func addItem(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	itemId := ctx.Param("item_id")
	rdb.HIncrBy(ctx, keyCart(orderId), itemId, 1)
	ctx.Status(http.StatusOK)
}

func removeItem(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	itemId := ctx.Param("item_id")
	rdb.HDecrIfGe0XX(ctx, keyCart(orderId), itemId)
	ctx.Status(http.StatusOK)
}
