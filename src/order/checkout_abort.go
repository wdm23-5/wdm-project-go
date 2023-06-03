package order

import (
	"bytes"
	"encoding/json"
	"net/http"
	"wdm/common"
)

func abortCkTxStock(txId string, cart map[string]int) {
	requests := make(map[string]*common.ItemTxPrpRequest, 4)
	for itemId, amount := range cart {
		if amount <= 0 {
			continue
		}
		// todo: group & send by id
		// mId := common.SnowflakeIDPickMachineIdFast(itemId)
		mId := "1"
		req := requests[mId]
		if req == nil {
			req = &common.ItemTxPrpRequest{TxId: txId, Items: make([]common.IdAmountPair, 0, 8)}
			requests[mId] = req
		}
		req.Items = append(req.Items, common.IdAmountPair{Id: itemId, Amount: amount})
	}

	for machineId, request := range requests {
		go func(mId string, req *common.ItemTxPrpRequest) {
			payload, err := json.Marshal(*req)
			if err != nil {
				return
			}
			// todo: make use of mId
			url := gatewayUrl + "stock/tx/checkout/abort/" + txId
			_, _ = http.Post(url, "application/json", bytes.NewBuffer(payload))
		}(machineId, request)
	}
}

func abortCkTxPayment(txId, userId string, price int) {
	req := common.CreditTxPrpRequest{TxId: txId, Payer: common.IdAmountPair{Id: userId, Amount: price}}
	payload, err := json.Marshal(req)
	if err != nil {
		return
	}
	// todo: make use of mId
	// mId := common.SnowflakeIDPickMachineIdFast(userId)
	url := gatewayUrl + "payment/tx/checkout/abort/" + txId
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(payload))
}
