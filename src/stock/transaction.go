package stock

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func prepareCkTx(ctx *gin.Context) {
	ctx.String(http.StatusTeapot, "prepareCkTx")
}

func commitCkTx(ctx *gin.Context) {
	ctx.String(http.StatusTeapot, "commitCkTx")
}

func abortCkTx(ctx *gin.Context) {
	ctx.String(http.StatusTeapot, "abortCkTx")
}
