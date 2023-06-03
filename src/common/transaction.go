package common

type TxState string

const (
	TxPreparing    TxState = "PRP"
	TxAcknowledged TxState = "ACK"
	TxCommitted    TxState = "CMT"
	TxAborted      TxState = "ABT"
)

func KeyTxState(txId string) string {
	return "tx_" + txId + ":state"
}