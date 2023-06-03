package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"wdm/common"
)

func createUser(ctx *gin.Context) {
	userId := snowGen.Next().String()
	rdb.Set(ctx, keyCredit(userId), 0, 0)
	ctx.JSON(http.StatusOK, common.CreateUserResponse{UserId: userId})
}

func findUser(ctx *gin.Context) {
	userId := ctx.Param("user_id")

	creditStr, err := rdb.Get(ctx, keyCredit(userId)).Result()
	if err == redis.Nil {
		ctx.String(http.StatusNotFound, userId)
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "findUser: %v", err)
		return
	}
	credit, err := strconv.Atoi(creditStr)
	if err != nil {
		ctx.String(http.StatusNotAcceptable, "findUser: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, common.FindUserResponse{
		UserId: userId,
		Credit: credit,
	})
}

func addCredit(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	amountStr := ctx.Param("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		ctx.String(http.StatusMethodNotAllowed, "addCredit: %v", err)
		return
	}

	_, err = rdb.IncrByIfGe0XX(ctx, keyCredit(userId), amount).Result()
	if err == redis.Nil {
		// special
		ctx.JSON(http.StatusOK, common.AddFundsResponse{Done: false})
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "addCredit: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, common.AddFundsResponse{Done: true})
}

func removeCredit(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	amountStr := ctx.Param("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		ctx.String(http.StatusMethodNotAllowed, "removeCredit: %v", err)
		return
	}

	_, err = rdb.IncrByIfGe0XX(ctx, keyCredit(userId), -amount).Result()
	if err == redis.Nil {
		// special
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "removeCredit: %v", err)
		return
	}

	ctx.Status(http.StatusOK)
}

// todo
// weird api, seems unused by the test suit
func cancelPayment(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	orderId := ctx.Param("order_id")
	_, _ = userId, orderId
	ctx.Status(http.StatusTeapot)
}

// todo
func paymentStatus(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	orderId := ctx.Param("order_id")
	_, _ = userId, orderId
	ctx.JSON(http.StatusTeapot, common.PaymentStatusResponse{Paid: false})
}
