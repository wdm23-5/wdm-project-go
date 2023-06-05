package order

import (
	"fmt"
	"net/http"
	"sync"
	"wdm/common"
)

func abortCkTxStock(txId string, cart map[string]int) {
	stockMachineIds := make(map[string]struct{}, 4)
	for itemId, amount := range cart {
		if amount <= 0 {
			continue
		}
		mId := common.SnowflakeIDPickMachineIdFast(itemId)
		stockMachineIds[mId] = struct{}{}
	}

	wg := sync.WaitGroup{}
	for machineId := range stockMachineIds {
		wg.Add(1)
		go func(mId string) {
			defer wg.Done()
			// hack shard key
			url := fmt.Sprintf("%vtx/checkout/abort/%v/t1m%vs0", stockServiceUrl, txId, mId)
			_, _ = http.Post(url, "text/plain", nil)
		}(machineId)
	}
	wg.Wait()
}

func abortCkTxPayment(txId, userId string) {
	url := fmt.Sprintf("%vtx/checkout/abort/%v/%v", paymentServiceUrl, txId, userId)
	_, _ = http.Post(url, "text/plain", nil)
}
