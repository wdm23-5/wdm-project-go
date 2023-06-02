package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"wdm/common"
)

func prepareCkTxLocal(ctx *gin.Context, txId, orderId string) (locked bool, info orderInfo, err error) {
	val, err := rdb.PrepareCkTx(ctx, txId, orderId).Result()
	if err == redis.Nil {
		return
	}
	if err != nil {
		err = fmt.Errorf("prepareCkTx: %v", err)
		return
	}

	// to array
	arr, ok := val.([]interface{})
	if !ok {
		err = errors.New("prepareCkTx: not an array of any")
		return
	}
	if len(arr) != 4 {
		err = errors.New("prepareCkTx: array length error")
		return
	}

	// array[0] to locked
	if lockedStr, ok := arr[0].(string); !ok {
		err = errors.New("prepareCkTx: array[0] not a string")
		return
	} else if lockedInt, errA := strconv.Atoi(lockedStr); errA != nil {
		err = fmt.Errorf("prepareCkTx: array[0] not an int (%v)", errA)
		return
	} else {
		locked = lockedInt != 0
	}

	// array[1] to userId
	if info.userId, ok = arr[1].(string); !ok {
		err = errors.New("prepareCkTx: array[1] not a string")
		return
	}

	// array[2] to paid
	if paidStr, ok := arr[2].(string); !ok {
		err = errors.New("prepareCkTx: array[2] not a string")
		return
	} else if paidInt, errA := strconv.Atoi(paidStr); errA != nil {
		err = fmt.Errorf("prepareCkTx: array[2] not an int (%v)", errA)
		return
	} else {
		info.paid = paidInt != 0
	}

	// array[3] to cart
	cart, ok := arr[3].([]interface{})
	if !ok {
		err = errors.New("prepareCkTx: array[3] not an array of any")
		return
	}
	cartLen := len(cart)
	if cartLen == 0 {
		info.cart = make(map[string]int, 0)
		goto skipCart
	}
	if cartLen&1 != 0 {
		err = fmt.Errorf("prepareCkTx: array[3] array length error (%v)", cartLen)
		return
	}
	info.cart = make(map[string]int, cartLen>>1)
	for i := 0; i < cartLen; i += 2 {
		itemId, ok := cart[i].(string)
		if !ok {
			err = fmt.Errorf("prepareCkTx: array[3][%v] not a string", i)
			return
		}
		amountStr, ok := cart[i+1].(string)
		if !ok {
			err = fmt.Errorf("prepareCkTx: array[3][%v] not a string", i+1)
			return
		}
		amount, errA := strconv.Atoi(amountStr)
		if errA != nil {
			err = fmt.Errorf("prepareCkTx: array[3][%v] not an int (%v)", i+1, amountStr)
			return
		}
		info.cart[itemId] = amount
	}

skipCart:
	if err != nil {
		panic(fmt.Sprintf("prepareCkTx: error state %v", err))
	}
	return
}

// ------- stock -------

func prepareCkTxStock(txId string, cart map[string]int) (price int, err error) {
	if len(cart) == 0 {
		return 0, nil
	}

	requests := make(map[string]*common.ItemTxPrpAbtRequest, 4)
	for itemId, amount := range cart {
		if amount <= 0 {
			continue
		}
		// todo: group & send by id
		// mId := common.SnowflakeIDPickMachineIdFast(itemId)
		mId := "1"
		req := requests[mId]
		if req == nil {
			req = &common.ItemTxPrpAbtRequest{TxId: txId, Items: make([]common.IdAmountPair, 0, 8)}
			requests[mId] = req
		}
		req.Items = append(req.Items, common.IdAmountPair{Id: itemId, Amount: amount})
	}

	errCh := make(chan string)
	go prepareCkTxStockSendRequests(txId, requests, errCh)
	errStr := <-errCh
	if strings.HasPrefix(errStr, "OK ") {
		price, err = strconv.Atoi(errStr[3:])
		if err != nil {
			panic("prepareCkTxStock: atoi")
		}
		return
	}
	return 0, errors.New("prepareCkTxStock: " + errStr)
}

// go this function and receive on ch
func prepareCkTxStockSendRequests(txId string, requests map[string]*common.ItemTxPrpAbtRequest, ch chan string) {
	okCh := make(chan int)
	errCh := make(chan string)
	count := 0
	var abort atomic.Bool
	for machineId, request := range requests {
		if abort.Load() {
			break
		}
		count++
		go func(mId string, req *common.ItemTxPrpAbtRequest) {
			if abort.Load() {
				errCh <- "abort"
				return
			}
			payload, err := json.Marshal(*req)
			if err != nil {
				abort.Store(true)
				errCh <- err.Error()
				return
			}
			// todo: make use of mId
			url := gatewayUrl + "stock/tx/checkout/prepare/" + txId
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
			if abort.Load() {
				errCh <- "abort"
				return
			}
			if err != nil {
				abort.Store(true)
				errCh <- "post " + err.Error()
				return
			}
			if resp.StatusCode != http.StatusOK {
				abort.Store(true)
				errCh <- "http " + resp.Status
				return
			}
			//goland:noinspection GoUnhandledErrorResult
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				abort.Store(true)
				errCh <- "read " + err.Error()
				return
			}
			var data common.ItemTxPrpResponse
			err = json.Unmarshal(body, &data)
			if err != nil {
				abort.Store(true)
				errCh <- "unmarshal " + err.Error()
				return
			}
			okCh <- data.TotalCost
		}(machineId, request)
	}

	totalPrice := 0
	hasErr := false
	for i := 0; i < count; i++ {
		select {
		case price := <-okCh:
			totalPrice += price
		case err := <-errCh:
			if !hasErr {
				ch <- "prepareCkTxStockSendRequests: " + err
				hasErr = true
			}
		}
	}
	if !hasErr {
		ch <- "OK " + strconv.Itoa(totalPrice)
	}
}

// ------- payment -------

func prepareCkTxPayment(txId, userId string, price int) error {
	req := common.CreditTxPrpAbtRequest{TxId: txId, Pay: common.IdAmountPair{Id: userId, Amount: price}}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("prepareCkTxPayment: %v", err)
	}
	// todo: make use of mId
	// mId := common.SnowflakeIDPickMachineIdFast(userId)
	url := gatewayUrl + "payment/tx/checkout/prepare/" + txId
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("prepareCkTxPayment: post %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prepareCkTxPayment: http %v", resp.StatusCode)
	}
	return nil
}
