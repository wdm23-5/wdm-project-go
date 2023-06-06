package order

func restartAllTx() {
	// In the current impl, we can only do provide fast look up from orderId to txId with keyCkTxId but
	// not the reverse. Therefore, we have to we scan the whole sharded database. It is slow, indeed.
	// We can save the orderId and userId in KeyTxLocked in the future since it is not used yet.

	// pattern on coordinator: abort PRP and ABT, commit ACK and CMT

}
