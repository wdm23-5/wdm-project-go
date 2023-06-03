package common

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
)

type CmdPingAble interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

func GinPingHandler(ctx *gin.Context, name string, snow *SnowflakeGenerator, rdb CmdPingAble) {
	sb := strings.Builder{}
	sb.WriteString(NowString())
	sb.WriteString(" ")
	sb.WriteString(name)
	sb.WriteString(" sf: ")
	sb.WriteString(snow.Next().String())
	sb.WriteString(" redis: ")
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		sb.WriteString(err.Error())
	} else {
		sb.WriteString(pong)
	}
	ctx.String(http.StatusOK, sb.String())
}
