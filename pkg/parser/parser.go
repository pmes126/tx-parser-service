package parser

type Parser interface {
// GetCurrentBlock last parsed block
GetCurrentBlock() (uint64, error)
// Subscribe address to observer
Subscribe(address string) bool
// GetTransactions list of inbound or outbound transactions for an address
GetTransactions(address string) ([]Transaction, error)

