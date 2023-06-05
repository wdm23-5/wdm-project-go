package order

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wdm/common"
)

var stockServiceUrl string   // append suffix WITHOUT stock and WITHOUT initial slash
var paymentServiceUrl string // append suffix WITHOUT payment and WITHOUT initial slash
var depId string
var snowGen *common.SnowflakeGenerator
var srdb *common.ShardedRedisDB[redisDB]

func init() {
	stockServiceUrl = common.MustGetEnv("STOCK_SERVICE_URL")
	paymentServiceUrl = common.MustGetEnv("PAYMENT_SERVICE_URL")

	mId := strings.Split(common.MustGetEnv("MACHINE_ID"), "/")
	if len(mId) != 2 {
		panic("error MACHINE_ID")
	}
	depId = mId[0]

	snowGen = common.NewSnowFlakeGenerator(depId) // for simplicity. see readme

	redisAddrs := strings.Split(common.MustGetEnv("REDIS_ADDRS"), ",")
	if len(redisAddrs) != common.MustS2I(mId[1]) {
		// panic due to identity mapping. see readme
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

	router.POST("/create/:user_id", createOrder)
	router.DELETE("/remove/:order_id", removeOrder)
	router.POST("/addItem/:order_id/:item_id", addItem)
	router.DELETE("/removeItem/:order_id/:item_id", removeItem)
	router.GET("/find/:order_id", findOrder)
	router.POST("/checkout/:order_id", checkoutOrder)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "order", snowGen, srdb.Select(0))
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

// the tx that is checking out / has checked out this order
func keyCkTxId(orderId string) string {
	return "order_" + orderId + ":tx_id"
}
