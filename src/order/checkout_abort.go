package order

import (
	"net/http"
	"sync"
)

func abortCkTxStock(txId string, cart map[string]int) {
	requests := make(map[string]struct{}, 4)
	for itemId, amount := range cart {
		if amount <= 0 {
			continue
		}
		// todo: group & send by id
		// mId := common.SnowflakeIDPickMachineIdFast(itemId)
		_ = itemId
		mId := "1"
		requests[mId] = struct{}{}
	}

	wg := sync.WaitGroup{}
	for machineId := range requests {
		wg.Add(1)
		go func(mId string) {
			defer wg.Done()
			// todo: make use of mId
			url := stockServiceUrl + "tx/checkout/abort/" + txId
			_, _ = http.Post(url, "text/plain", nil)
		}(machineId)
	}
	wg.Wait()
}

func abortCkTxPayment(txId string) {
	// todo: make use of mId
	// mId := common.SnowflakeIDPickMachineIdFast(userId)
	url := paymentServiceUrl + "tx/checkout/abort/" + txId
	_, _ = http.Post(url, "text/plain", nil)
}
