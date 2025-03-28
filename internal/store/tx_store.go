package store

import "errors"

// TxStore is an interface for storing transactions of any type
type TxStore[T any] interface {
	// AddTransaction adds a transaction to the store
	AddTransaction(address string, tx T) error
	// GetTransactions returns a list of transactions for an address
	GetTransactions(address string) ([]T, error)
}

var (
	ErrAddressNotFound = errors.New("Address not found in store")
	ErrNoTransactions  = errors.New("no transactions found for address")
)
