package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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

	router.GET("/ping", func(ctx *gin.Context) {
		sb := strings.Builder{}
		sb.WriteString(common.NowString())
		sb.WriteString(" order sf: ")
		sb.WriteString(snowGen.Next().String())
		sb.WriteString(" redis: ")
		pong, err := rdb.Ping(ctx).Result()
		if err != nil {
			sb.WriteString(err.Error())
		} else {
			sb.WriteString(pong)
		}
		ctx.String(http.StatusOK, sb.String())
	})

	router.DELETE("/drop-database", func(ctx *gin.Context) {
		rdb.FlushDB(ctx)
		ctx.Status(http.StatusOK)
	})

	_ = router.Run("0.0.0.0:5000")
}
