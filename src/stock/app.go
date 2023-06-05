package stock

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wdm/common"
)

var depId string
var snowGen *common.SnowflakeGenerator
var srdb *common.ShardedRedisDB[redisDB]

func init() {
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

	router.POST("/item/create/:price", createItem)
	router.GET("/find/:item_id", findItem)
	router.POST("/add/:item_id/:amount", addStock)
	router.POST("/subtract/:item_id/:amount", removeStock)

	router.POST("/tx/checkout/prepare/:tx_id/:shard_key", prepareCkTx)
	router.POST("/tx/checkout/commit/:tx_id/:shard_key", commitCkTx)
	router.POST("/tx/checkout/abort/:tx_id/:shard_key", abortCkTx)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	router.GET("/ping", func(ctx *gin.Context) {
		common.GinPingHandler(ctx, "stock", snowGen, srdb.Select(0))
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

func keyPrice(itemId string) string {
	return "item_" + itemId + ":price"
}

func keyStock(itemId string) string {
	return "item_" + itemId + ":stock"
}
