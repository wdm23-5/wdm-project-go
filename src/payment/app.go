package payment

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

	router.POST("/create_user", createUser)
	router.GET("/find_user/:user_id", findUser)
	router.POST("/add_funds/:user_id/:amount", addCredit)
	router.POST("/pay/:user_id/:order_id/:amount", removeCredit)
	router.POST("/cancel/:user_id/:order_id", cancelPayment)
	router.GET("/status/:user_id/:order_id", paymentStatus)

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "payment", snowGen, rdb)
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		rdb.FlushDB(ctx)
		ctx.Status(http.StatusOK)
	})

	_ = router.Run("0.0.0.0:5000")
}

func keyCredit(userId string) string {
	return "user_" + userId + ":credit"
}
