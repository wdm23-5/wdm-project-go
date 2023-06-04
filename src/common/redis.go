package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type RedisCmdRequest struct {
	Args []string `json:"args"`
}

type RedisCmdResponse struct {
	Val string `json:"val"`
	Err string `json:"err"`
}

func RedisCmdHandler(ctx *gin.Context, rdb redis.UniversalClient) {
	var req RedisCmdRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "RedisCmdHandler: %v", err)
		return
	}
	tmp := make([]any, len(req.Args))
	for i, a := range req.Args {
		tmp[i] = a
	}
	val, err := rdb.Do(ctx, tmp...).Result()
	if err != nil {
		ctx.JSON(http.StatusOK, RedisCmdResponse{Val: fmt.Sprintf("%v", val), Err: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, RedisCmdResponse{Val: fmt.Sprintf("%v", val), Err: ""})
}
