package order

import (
	"fmt"
	"net/http"
	"wdm/common"
)

func commitCkTxRemote(txId string, info orderInfo) {
	stockMachineIds := make(map[string]struct{}, 4)
	for itemId, amount := range info.cart {
		if amount <= 0 {
			continue
		}
		mId := common.SnowflakeIDPickMachineIdFast(itemId)
		stockMachineIds[mId] = struct{}{}
	}

	for machineId := range stockMachineIds {
		go func(mId string) {
			// hack shard key
			url := fmt.Sprintf("%vtx/checkout/commit/%v/t1m%vs0", stockServiceUrl, txId, mId)
			_, _ = http.Post(url, "text/plain", nil)
		}(machineId)
	}

	go func() {
		url := fmt.Sprintf("%vtx/checkout/commit/%v/%v", paymentServiceUrl, txId, info.userId)
		_, _ = http.Post(url, "text/plain", nil)
	}()
}
