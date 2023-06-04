package common

// always cast to string when passed to redis
type TxState string

// On receiver we have
//
//  ø  -> ABT (fast abort)
//  v
// PRP -> ABT
//  v
// ACK -> ABT
//  v
// CMT

const (
	TxPreparing    TxState = "PRP"
	TxAcknowledged TxState = "ACK"
	TxCommitted    TxState = "CMT"
	TxAborted      TxState = "ABT"
)

func KeyTxState(txId string) string {
	return "tx_" + txId + ":state"
}

// key of the data locked by the tx
func KeyTxLocked(txId string) string {
	return "tx_" + txId + ":locked"
}
