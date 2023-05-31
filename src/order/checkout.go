package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func checkoutOrder(ctx *gin.Context) {
	// todo: use message queue to limit rate
	ctx.String(http.StatusTeapot, "checkoutOrder")
}
