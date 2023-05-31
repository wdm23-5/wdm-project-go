package stock

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"wdm/common"
)

func createItem(ctx *gin.Context) {
	priceStr := ctx.Param("price")
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		ctx.String(http.StatusMethodNotAllowed, "createItem: %v", err)
		return
	}
	itemId := snowGen.Next().String()

	pipe := rdb.TxPipeline()
	pipe.Set(ctx, keyPrice(itemId), price, 0)
	pipe.Set(ctx, keyStock(itemId), 0, 0)
	_, err = pipe.Exec(ctx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "createItem: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, common.CreateItemResponse{ItemId: itemId})
}

func findItem(ctx *gin.Context) {
	itemId := ctx.Param("item_id")

	pipe := rdb.TxPipeline()
	priceCmd := pipe.Get(ctx, keyPrice(itemId))
	stockCmd := pipe.Get(ctx, keyStock(itemId))
	_, err := pipe.Exec(ctx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "findItem: %v", err)
		return
	}

	priceStr, err := priceCmd.Result()
	if err == redis.Nil {
		ctx.String(http.StatusNotFound, itemId)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "findItem: %v", err)
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		ctx.String(http.StatusNotAcceptable, "findItem: %v", err)
		return
	}

	stockStr, err := stockCmd.Result()
	if err == redis.Nil {
		ctx.String(http.StatusNotFound, itemId)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "findItem: %v", err)
		return
	}
	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		ctx.String(http.StatusNotAcceptable, "findItem: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, common.FindItemResponse{
		Price: price,
		Stock: stock,
	})
}

func addStock(ctx *gin.Context) {
	itemId := ctx.Param("item_id")
	amountStr := ctx.Param("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		ctx.String(http.StatusMethodNotAllowed, "addStock: %v", err)
		return
	}

	_, err = rdb.IncrByIfGe0XX(ctx, keyStock(itemId), amount).Result()
	if err == redis.Nil {
		ctx.String(http.StatusNotFound, itemId)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "addStock: %v", err)
		return
	}

	ctx.Status(http.StatusOK)
}

func removeStock(ctx *gin.Context) {
	itemId := ctx.Param("item_id")
	amountStr := ctx.Param("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		ctx.String(http.StatusMethodNotAllowed, "removeStock: %v", err)
		return
	}

	_, err = rdb.IncrByIfGe0XX(ctx, keyStock(itemId), -amount).Result()
	if err == redis.Nil {
		ctx.String(http.StatusNotFound, itemId)
		return
	} else if err != nil {
		ctx.String(http.StatusInternalServerError, "removeStock: %v", err)
		return
	}

	ctx.Status(http.StatusOK)
}
