package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
	"wdm/common"
)

func findOrder(ctx *gin.Context) {
	orderId := ctx.Param("order_id")

	info, err := loadOrderInfo(ctx, orderId)
	if err == redis.Nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "findOrder: loadOrderInfo: %v", err)
		return
	}

	var itemsFlat []string
	priceCh := make(chan int)
	errCh := make(chan string)
	nThread := 0
	var abort atomic.Bool
	for itemId, amount := range info.cart {
		if abort.Load() {
			break
		}
		if amount <= 0 {
			continue
		}
		for i := 0; i < amount; i++ {
			itemsFlat = append(itemsFlat, itemId)
		}
		nThread++
		go func(itemId string, amount int) {
			if abort.Load() {
				errCh <- "abort"
				return
			}
			price, err := getItemPrice(ctx, itemId)
			if err != nil {
				abort.Store(true)
				errCh <- err.Error()
				return
			}
			priceCh <- price * amount
		}(itemId, amount)
	}

	totalPrice := 0
	for i := 0; i < nThread; i++ {
		select {
		case price := <-priceCh:
			totalPrice += price
		case errStr := <-errCh:
			ctx.String(http.StatusNotFound, errStr)
			go func(i int) {
				// clean up
				for i++; i < nThread; i++ {
					select {
					case <-priceCh:
					case <-errCh:
					}
				}
			}(i)
			return
		}
	}

	ctx.JSON(http.StatusOK, common.FindOrderResponse{
		OrderId:   orderId,
		Paid:      info.paid,
		Items:     itemsFlat,
		UserId:    info.userId,
		TotalCost: totalPrice,
	})
}

func loadOrderInfo(ctx *gin.Context, orderId string) (info orderInfo, err error) {
	pipe := rdb.TxPipeline()
	userIdCmd := pipe.Get(ctx, keyUserId(orderId))
	paidCmd := pipe.Get(ctx, keyPaid(orderId))
	cartCmd := pipe.HGetAll(ctx, keyCart(orderId))
	if _, err = pipe.Exec(ctx); err != nil {
		return
	}

	if info.userId, err = userIdCmd.Result(); err != nil {
		return
	}

	paidInt, err := paidCmd.Int()
	if err != nil {
		return
	}
	info.paid = paidInt != 0

	cart, err := cartCmd.Result()
	if err != nil {
		return
	}
	info.cart = make(map[string]int, len(cart))
	for itemId, amountStr := range cart {
		amount, errA := strconv.Atoi(amountStr)
		if errA != nil {
			err = errA
			return
		}
		info.cart[itemId] = amount
	}

	return
}

// prices for hot products are cached
func getItemPrice(ctx context.Context, itemId string) (price int, err error) {
	key := "item_" + itemId + ":price"
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// no such item in cache
		item, err := getItemFromRemote(itemId)
		if err != nil {
			return 0, err
		}
		// todo: limit size growth with lfu
		rdb.Set(ctx, key, item.Price, 5*time.Minute)
		return item.Price, nil
	} else if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func getItemFromRemote(itemId string) (data common.FindItemResponse, err error) {
	// todo: make use of mId
	resp, err := http.Post(gatewayUrl+"/stock/find/"+itemId, "text/plain", nil)
	if err != nil {
		err = fmt.Errorf("getItemFromRemote: post %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("getItemFromRemote: http %v", resp.Status)
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("getItemFromRemote: read %v", err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		err = fmt.Errorf("getItemFromRemote: unmarshal %v", err)
	}
	return
}
