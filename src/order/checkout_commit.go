package order

import (
	"net/http"
)

func commitCkTxRemote(txId string, info orderInfo) {
	// todo: group & send by id
	stockMachineIds := make(map[string]struct{}, 4)
	// for itemId := range info.cart {
	// 	mId := common.SnowflakeIDPickMachineIdFast(itemId)
	// 	stockMachineIds[mId] = struct{}{}
	// }
	stockMachineIds["1"] = struct{}{}

	for machineId := range stockMachineIds {
		go func(mId string) {
			// todo: make use of mId
			url := gatewayUrl + "stock/tx/checkout/commit/" + txId
			_, _ = http.Post(url, "text/plain", nil)
		}(machineId)
	}

	go func() {
		// todo: make use of mId
		url := gatewayUrl + "payment/tx/checkout/commit/" + txId
		_, _ = http.Post(url, "text/plain", nil)
	}()
}
