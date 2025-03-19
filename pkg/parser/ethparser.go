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
	RpcUrl                      = "https://ethereum-rpc.publicnode.com"
	GetCurrentBlock             = "eth_blockNumber"
	GetCurrentBlockByNumber     = "eth_getBlockByNumber"
	CurrentBlockParam           = "latest"
	CurrentBlockPollingInterval = 5 * time.Second
)

type EthTxParser struct {
	TxStore   store.TxStore[EthTransaction]
	Addresses map[string]bool
	LastBlock int64
	Client    http.Client
	mx        sync.RWMutex
	logger    slog.Logger
}

type CurrentBlockRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type BlockByNumberRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type EthBlockNumberResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

type EthBlockByNumberResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Transactions []EthTransaction `json:"transactions"` // Hashes or objects?
	} `json:"result"`
}

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

func NewEthTxParser(store store.TxStore[EthTransaction], log *slog.Logger) *EthTxParser {
	return &EthTxParser{
		Addresses: make(map[string]bool),
		Client:    http.Client{},
		TxStore:   store,
	}
}

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
	resp, err := ep.Client.Post(RpcUrl, "application/json", bytes.NewBuffer(reqBody))
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

func (ep *EthTxParser) Start(ctx context.Context) {
	//go ep.pollCurrentBlock()
	ticker := time.NewTicker(CurrentBlockPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latestBlock, err := ep.GetCurrentBlock()
			if err != nil {
				fmt.Printf("Error getting latest block: %s", err)
				continue
			}
			if latestBlock > ep.LastBlock {
				for i := latestBlock; i > ep.LastBlock; i-- {
					ep.UpdateTransactionsFromBlock(i)
				}
				ep.LastBlock = latestBlock
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ep *EthTxParser) UpdateTransactionsFromBlock(blockNum int64) error {
	req := BlockByNumberRequest{
		Jsonrpc: "2.0",
		Method:  GetCurrentBlockByNumber,
		Params:  []interface{}{fmt.Sprintf("0x%x", blockNum), true}, // `true` includes transactions
		Id:      1,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := ep.Client.Post(RpcUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Println("BODY:", string(body))
	var ethResp EthBlockByNumberResponse
	err = json.Unmarshal(body, &ethResp)
	if err != nil {
		return err
	}

	transactions := ethResp.Result.Transactions
	ep.mx.RLock()
	defer ep.mx.RUnlock()
	for _, tx := range transactions {
		//fmt.Printf("Transaction: %+v\n", tx)
		var address string
		if ep.Addresses[tx.From] {
			address = tx.From
		} else if ep.Addresses[tx.To] {
			address = tx.To
		} else {
			continue
		}
		ep.TxStore.AddTransaction(address, tx)
	}

	return nil
}

func (ep *EthTxParser) Subscribe(address string) bool {
	ep.mx.Lock()
	defer ep.mx.Unlock()
	ep.Addresses[address] = true
	return true
}

func ParseHex(hex string) (int64, error) {
	h := strings.TrimPrefix(hex, "0x")
	val, err := strconv.ParseInt(h, 16, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
