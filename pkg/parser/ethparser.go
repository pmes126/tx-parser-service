package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pmes126/tx-parser-service/internal/store"
)

const (
	RpcUrl                  = "https://ethereum-rpc.publicnode.com"
	GetCurrentBlock         = "eth_blockNumber"
	GetCurrentBlockByNumber = "eth_getBlockByNumber"
	CurrentBlockParam       = "latest"
)

// EthTxParser is a parser for Ethereum transactions.
type EthTxParser struct {
	txStore              store.TxStore[EthTransaction]
	addresses            map[string]bool
	lastBlock            int64
	blockPollingInterval time.Duration
	client               http.Client
	mx                   sync.RWMutex
	logger               slog.Logger
}

// CurrentBlockRequest represents a request to get the current block number.
type CurrentBlockRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

// BlockByNumberRequest represents a request to get a block by number.
type BlockByNumberRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

// EthBlockNumberResponse represents the response from an Ethereum block number request.
type EthBlockNumberResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// EthBlockByNumberResponse represents the response from an Ethereum block request.
type EthBlockByNumberResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Transactions []EthTransaction `json:"transactions"` // Hashes or objects?
	} `json:"result"`
}

// EthTransaction represents an Ethereum transaction.
type EthTransaction struct {
	Address     string `json:"address"`
	Hash        string `json:"hash"`
	Nonce       string `json:"nonce"`
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Input       string `json:"input"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
}

// NewEthTxParser creates a new EthTxParser
func NewEthTxParser(store store.TxStore[EthTransaction], log *slog.Logger, pollingInterval int) *EthTxParser {
	return &EthTxParser{
		addresses:            make(map[string]bool),
		client:               http.Client{},
		txStore:              store,
		logger:               *log,
		blockPollingInterval: time.Duration(pollingInterval) * time.Second,
	}
}

// GetCurrentBlock returns the current block number in the blockchain.
func (ep *EthTxParser) GetCurrentBlock() (int64, error) {
	req := CurrentBlockRequest{
		Jsonrpc: "2.0",
		Method:  GetCurrentBlock,
		Params:  []interface{}{},
		Id:      1,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	resp, err := ep.client.Post(RpcUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var ethResp EthBlockNumberResponse
	err = json.Unmarshal(body, &ethResp)
	if err != nil {
		return 0, err
	}

	blockNumber, err := ParseHex(ethResp.Result)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

// Start starts the EthTxParser, polling the blockchain for new blocks and updating transactions.
func (ep *EthTxParser) Start(ctx context.Context) {
	ticker := time.NewTicker(ep.blockPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latestBlock, err := ep.GetCurrentBlock()
			if err != nil {
				ep.logger.Error("Error getting latest block", slog.String("error", err.Error()))
				continue
			}
			if latestBlock > ep.lastBlock {
				for i := latestBlock; i > ep.lastBlock; i-- {
					if err = ep.UpdateTransactionsFromBlock(i); err != nil {
						ep.logger.Error("Error Updating Transactions from block", slog.Int64("block id", i), slog.String("error", err.Error()))
					}
				}
				ep.lastBlock = latestBlock
			}
		case <-ctx.Done():
			return
		}
	}
}

// QueryTransactionsFromBlock queries the blockchain for transactions in a given block.
func (ep *EthTxParser) QueryTransactionsFromBlock(blockNum int64) ([]EthTransaction, error) {
	req := BlockByNumberRequest{
		Jsonrpc: "2.0",
		Method:  GetCurrentBlockByNumber,
		Params:  []interface{}{fmt.Sprintf("0x%x", blockNum), true}, // `true` includes transactions
		Id:      1,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := ep.client.Post(RpcUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ethResp EthBlockByNumberResponse
	err = json.Unmarshal(body, &ethResp)
	if err != nil {
		return nil, err
	}

	return ethResp.Result.Transactions, nil
}

// UpdateTransactionsFromBlock updates the transaction store with transactions from the given block.
func (ep *EthTxParser) UpdateTransactionsFromBlock(blockNum int64) error {
	ep.mx.RLock()
	defer ep.mx.RUnlock()
	transactions, err := ep.QueryTransactionsFromBlock(blockNum)
	if err != nil {
		return err
	}
	for _, tx := range transactions {
		//fmt.Printf("Transaction: %+v\n", tx)
		var address string
		if ep.addresses[tx.From] {
			address = tx.From
		} else if ep.addresses[tx.To] {
			address = tx.To
		} else {
			continue
		}
		ep.txStore.AddTransaction(address, tx)
	}
	return nil
}

// Subscribe adds an address to the list of addresses to track.
func (ep *EthTxParser) Subscribe(address string) bool {
	ep.mx.Lock()
	defer ep.mx.Unlock()
	ep.addresses[address] = true
	return true
}

// GetTransactions returns a list of transactions for an address from the Transaction store.
func (ep *EthTxParser) GetTransactions(address string) ([]EthTransaction, error) {
	txs, err := ep.txStore.GetTransactions(address)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// ParseHex parses a hex string into an int64.
func ParseHex(hex string) (int64, error) {
	h := strings.TrimPrefix(hex, "0x")
	val, err := strconv.ParseInt(h, 16, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
