package payment

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"strconv"
	"wdm/common"
)

func createUser(ctx *gin.Context) {
	userId := snowGen.Next().String()
	shardKey := userId
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "createUser: error shard key %v", shardKey)
		return
	}
	rdb.Set(ctx, keyCredit(userId), 0, 0)
	ctx.JSON(http.StatusOK, common.CreateUserResponse{UserId: userId})
}

func findUser(ctx *gin.Context) {
	userId := ctx.Param("user_id")

	shardKey := userId
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "findUser: error shard key %v", shardKey)
		return
	}

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

	shardKey := userId
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "addCredit: error shard key %v", shardKey)
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

	shardKey := userId
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "removeCredit: error shard key %v", shardKey)
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

// so-called internal api, seems unused by the test suit
// also not used by us
// very slow
func cancelPayment(ctx *gin.Context) {
	// userId := ctx.Param("user_id")
	orderId := ctx.Param("order_id")
	valid, resp := getOrderFromRemote(ctx, orderId)
	if !valid {
		return
	}
	shardKey := resp.UserId
	rdb := srdb.Route(shardKey)
	if rdb == nil {
		ctx.String(http.StatusPreconditionFailed, "cancelPayment: error shard key %v", shardKey)
		return
	}
	rdb.IncrByIfGe0XX(ctx, keyCredit(resp.UserId), resp.TotalCost)
	ctx.Status(http.StatusOK)
}

// seems unused, thus slow impl
func paymentStatus(ctx *gin.Context) {
	// userId := ctx.Param("user_id")
	orderId := ctx.Param("order_id")
	valid, resp := getOrderFromRemote(ctx, orderId)
	if !valid {
		return
	}
	ctx.JSON(http.StatusOK, common.PaymentStatusResponse{Paid: resp.Paid})
}

func getOrderFromRemote(ctx *gin.Context, orderId string) (valid bool, data common.FindOrderResponse) {
	valid = false
	resp, err := http.Get(orderServiceUrl + "find/" + orderId)
	if err != nil {
		ctx.String(http.StatusTeapot, "getOrderFromRemote: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		ctx.String(http.StatusTeapot, "getOrderFromRemote: http %v", resp.Status)
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.String(http.StatusTeapot, "getOrderFromRemote: read %v", err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.String(http.StatusTeapot, "getOrderFromRemote: unmarshal %v", err)
		return
	}
	valid = true
	return
}
