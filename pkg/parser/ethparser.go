package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RpcUrl                      = "https://ethereum-rpc.publicnode.com"
	GetCurrentBlock             = "eth_blockNumber"
	GetCurrentBlockByNumber     = "eth_getBlockByNumber"
	CurrentBlockParam           = "latest"
	CurrentBlockPollingInterval = 5 * time.Second
)

type EthTxParser struct {
	//Store Storage
	Addresses map[string]bool
	LastBlock uint64
	Client    http.Client
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

func NewEthTxParser() *EthTxParser {
	return &EthTxParser{
		Addresses: make(map[string]bool),
		Client:    http.Client{},
	}
}

func (ep *EthTxParser) GetCurrentBlock() (uint64, error) {
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

	blockNumber, err := strconv.ParseUint(ethResp.Result, 0, 64)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func (ep *EthTxParser) Start() {
	//go ep.pollCurrentBlock()
	ticker := time.NewTicker(CurrentBlockPollingInterval)
	defer ticker.Stop()

	for _ = range ticker.C {
		latestBlock, err := ep.GetCurrentBlock()
		if err != nil {
			fmt.Println("Error getting latest block:", err)
			continue
		}
		if latestBlock > ep.LastBlock {
			for i := latestBlock; i > ep.LastBlock; i-- {
				ep.UpdateTransactionsFromBlock(i)
			}
			ep.LastBlock = latestBlock
		}
	}
}
func (ep *EthTxParser) Stop() {
	//ep.Store.Close()
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
	var ethResp EthBlockByNumberResponse
	err = json.Unmarshal(body, &ethResp)
	if err != nil {
		return err
	}

	transactions := ethResp.Result.Transactions
	for _, tx := range transactions {
		fmt.Println("Transaction:", tx)
		//if ep.Addresses[from] || ep.Addresses[to] {
		//	t := EthTransaction{
		//ep.Store.AddTransaction(Transaction{
		//	From:  from,
		//	To:    to,
		//	Value: value,
		//})
		//}
	}

	return nil
}

func ParseHex(hex string) (int64, error) {
	h := strings.TrimPrefix(hex, "0x")
	val, err := strconv.ParseInt(h, 16, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
