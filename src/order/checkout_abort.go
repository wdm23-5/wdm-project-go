package order

import (
	"net/http"
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

	for machineId := range requests {
		go func(mId string) {
			// todo: make use of mId
			url := gatewayUrl + "stock/tx/checkout/abort/" + txId
			_, _ = http.Post(url, "text/plain", nil)
		}(machineId)
	}
}

func abortCkTxPayment(txId string) {
	// todo: make use of mId
	// mId := common.SnowflakeIDPickMachineIdFast(userId)
	url := gatewayUrl + "payment/tx/checkout/abort/" + txId
	_, _ = http.Post(url, "text/plain", nil)
}
