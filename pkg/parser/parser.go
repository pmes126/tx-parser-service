package parser

import (
	"errors"
)

const (
	EthAddressLength = 42
)

var (
	ErrAddressNotTracked = errors.New("Address not tracked")
)

type Parser interface {
	// GetCurrentBlock last parsed block
	GetCurrentBlock() (int64, error)
	// Subscribe address to observer
	Subscribe(address string) bool
	// GetTransactions list of inbound or outbound transactions for an address
	GetTransactions(address string) ([]EthTransaction, error)
}
