package stock

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"wdm/common"
)

var snowGen *common.SnowflakeGenerator
var rdb *redisDB

func Main() {
	snowGen = common.NewSnowFlakeGenerator(common.MustGetEnv("MACHINE_ID"))
	rdb = newRedisDB()
	go func() {
		time.Sleep(time.Second)
		if err := rdb.CacheAllScripts(context.Background()); err != nil {
			panic("load lua script: " + err.Error())
		}
	}()

	router := gin.New()
	common.DEffect(func() { router.Use(common.GinLogger()) })

	router.POST("/item/create/:price", createItem)
	router.GET("/find/:item_id", findItem)
	router.POST("/add/:item_id/:amount", addStock)
	router.POST("/subtract/:item_id/:amount", removeStock)

	router.POST("/tx/checkout/prepare/:tx_id", prepareCkTx)
	router.POST("/tx/checkout/commit/:tx_id", commitCkTx)
	router.POST("/tx/checkout/abort/:tx_id", abortCkTx)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "stock", snowGen, rdb)
	})

	router.POST("/redis-exec", func(ctx *gin.Context) {
		common.RedisCmdHandler(ctx, rdb)
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		rdb.FlushAll(ctx)
		ctx.Status(http.StatusOK)
	})

	_ = router.Run("0.0.0.0:5000")
}

func keyPrice(itemId string) string {
	return "item_" + itemId + ":price"
}

func keyStock(itemId string) string {
	return "item_" + itemId + ":stock"
}
