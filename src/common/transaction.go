package common

import (
	"math/rand"
	"time"
)

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

var rand3aIdx uint8
var rand3a []time.Duration

// fast though not so random number in [3, 10]
// exclusive for waiting on tx state change
func TxRand3A() time.Duration {
	rand3aIdx++
	return rand3a[rand3aIdx]
}

func init() {
	rand3aIdx = 0
	rand3a = make([]time.Duration, 1<<8)
	for i := range rand3a {
		rand3a[i] = time.Duration(3 + rand.Int63n(8))
	}
}
