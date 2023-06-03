package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wdm/common"
)

var gatewayUrl string
var snowGen *common.SnowflakeGenerator
var rdb *redisDB

func Main() {
	gatewayUrl = common.MustGetEnv("GATEWAY_URL")
	snowGen = common.NewSnowFlakeGenerator(common.MustGetEnv("MACHINE_ID"))
	rdb = newRedisDB()

	router := gin.New()
	common.DEffect(func() { router.Use(common.GinLogger()) })

	router.POST("/create/:user_id", createOrder)
	router.DELETE("/remove/:order_id", removeOrder)
	router.POST("/addItem/:order_id/:item_id", addItem)
	router.DELETE("/removeItem/:order_id/:item_id", removeItem)
	router.GET("/find/:order_id", findOrder)
	router.POST("/checkout/:order_id", checkoutOrder)

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "order", snowGen, rdb)
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		rdb.FlushDB(ctx)
		ctx.Status(http.StatusOK)
	})

	_ = router.Run("0.0.0.0:5000")
}

type orderInfo struct {
	userId string
	paid   bool
	cart   map[string]int
}

func keyUserId(orderId string) string {
	return "order_" + orderId + ":user_id"
}

func keyPaid(orderId string) string {
	return "order_" + orderId + ":paid"
}

func keyCart(orderId string) string {
	return "order_" + orderId + ":item_id:amount"
}

func keyCkTxId(orderId string) string {
	return "order_" + orderId + ":tx_id"
}
