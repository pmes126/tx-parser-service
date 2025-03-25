package store

import (
	"sync"
)

// MemTxStore is an in-memory implementation of TxStore
type MemTxStore[T any] struct {
	// Transactions is a map of address to transactions
	Transactions map[string][]T
	// Mutex for synchronizing access to Transactions
	mx sync.Mutex
}

// NewMemTxStore creates a new MemTxStore
func NewMemTxStore[T any]() *MemTxStore[T] {
	return &MemTxStore[T]{
		Transactions: make(map[string][]T),
	}
}

// AddTransaction adds a transaction to the store
func (mts *MemTxStore[T]) AddTransaction(address string, tx T) error {
	mts.mx.Lock()
	defer mts.mx.Unlock()
	if _, ok := mts.Transactions[address]; !ok {
		mts.Transactions[address] = []T{tx}
	} else {
		mts.Transactions[address] = append(mts.Transactions[address], tx)
	}
	return nil
}

// GetTransactions returns a list of transactions for an address
func (mts *MemTxStore[T]) GetTransactions(address string) ([]T, error) {
	mts.mx.Lock()
	defer mts.mx.Unlock()
	if txs, ok := mts.Transactions[address]; ok {
		res := make([]T, len(txs))
		copy(res, txs) // Deep copy.
		return res, nil
	}
	return nil, ErrNoTransactions
}
