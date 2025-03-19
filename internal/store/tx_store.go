package store

import "errors"

// TxStore is an interface for storing transactions of any type
type TxStore[T any] interface {
	// AddTransaction adds a transaction to the store
	AddTransaction(address string, tx T) error
	// GetTransactions returns a list of transactions for an address
	GetTransactions(address string) ([]T, error)
	// LockStore locks the store for writing/reading
	LockStore()
	// UnlockStore unlocks the store
	UnlockStore()
}

var (
	ErrAddressNotFound = errors.New("Address not tracked")
	ErrNoTransactions  = errors.New("no transactions found for address")
)
