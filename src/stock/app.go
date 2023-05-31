package stock

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wdm/common"
)

var snowGen *common.SnowflakeGenerator
var rdb *redisDB

func Main() {
	snowGen = common.NewSnowFlakeGenerator(common.MustGetEnv("MACHINE_ID"))
	rdb = newRedisDB()

	router := gin.New()
	common.DEffect(func() { router.Use(common.GinLogger()) })

	router.POST("/item/create/:price", createItem)
	router.GET("/find/:item_id", findItem)
	router.POST("/add/:item_id/:amount", addStock)
	router.POST("/subtract/:item_id/:amount", removeStock)

	router.POST("/checkout/tx/prepare/:tx_id", prepareCkTx)
	router.POST("/checkout/tx/commit/:tx_id", commitCkTx)
	router.POST("/checkout/tx/abort/:tx_id", abortCkTx)

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "stock", snowGen, rdb)
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		rdb.FlushDB(ctx)
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
