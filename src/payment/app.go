package payment

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wdm/common"
)

var orderServiceUrl string
var depId string
var snowGen *common.SnowflakeGenerator
var srdb *common.ShardedRedisDB[redisDB]

func init() {
	orderServiceUrl = common.MustGetEnv("ORDER_SERVICE_URL")

	mId := strings.Split(common.MustGetEnv("MACHINE_ID"), "/")
	if len(mId) != 2 {
		panic("error MACHINE_ID")
	}
	depId = mId[0]

	snowGen = common.NewSnowFlakeGenerator(depId)

	redisAddrs := strings.Split(common.MustGetEnv("REDIS_ADDRS"), ",")
	if len(redisAddrs) != common.MustS2I(mId[1]) {
		panic("error REDIS_ADDRS")
	}
	pwd := common.MustGetEnv("REDIS_PASSWORD")
	idb := common.MustS2I(common.MustGetEnv("REDIS_DB"))
	rdbs := make([]*redisDB, len(redisAddrs))
	for i, addr := range redisAddrs {
		rdb := newRedisDB(addr, pwd, idb)
		go func() {
			time.Sleep(2 * time.Second)
			if err := rdb.CacheAllScripts(context.Background()); err != nil {
				panic("load lua script: " + err.Error())
			}
		}()
		rdbs[i] = rdb
	}
	srdb = common.NewShardedRedisDB(rdbs...)
}

func Main() {
	router := gin.New()
	common.DEffect(func() { router.Use(common.GinLogger()) })

	router.POST("/create_user", createUser)
	router.GET("/find_user/:user_id", findUser)
	router.POST("/add_funds/:user_id/:amount", addCredit)
	router.POST("/pay/:user_id/:order_id/:amount", removeCredit)
	router.POST("/cancel/:user_id/:order_id", cancelPayment)
	router.GET("/status/:user_id/:order_id", paymentStatus)

	router.POST("/tx/checkout/prepare/:tx_id/:shard_key", prepareCkTx)
	router.POST("/tx/checkout/commit/:tx_id/:shard_key", commitCkTx)
	router.POST("/tx/checkout/abort/:tx_id/:shard_key", abortCkTx)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "payment", snowGen, srdb.Select(0))
	})

	router.POST("/redis-exec/:idx", func(ctx *gin.Context) {
		str := ctx.Param("idx")
		idx, err := strconv.Atoi(str)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		common.RedisCmdHandler(ctx, srdb.Select(idx))
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		srdb.ForEach(func(r *redisDB) { r.FlushAll(ctx) })
		ctx.Status(http.StatusOK)
	})

	_ = router.Run("0.0.0.0:5000")
}

func keyCredit(userId string) string {
	return "user_" + userId + ":credit"
}
