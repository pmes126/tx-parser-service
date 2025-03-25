package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pmes126/tx-parser-service/internal/store"
	"github.com/pmes126/tx-parser-service/pkg/parser"
)

func TestHandler_handleGetTransactions(t *testing.T) {
	type fields struct {
		logger      *slog.Logger
		httpTimeout int
		rr          *httptest.ResponseRecorder
	}
	type args struct {
		w            http.ResponseWriter
		r            *http.Request
		transactions []parser.EthTransaction
		address      string
		codeWant     int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test handleGetTransactions",
			fields: fields{
				logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
				httpTimeout: 5,
			},
			args: args{
				address: "0xc0ffee254729296a45a3885639AC7E10F9d54979",
				w:       httptest.NewRecorder(),
				r:       httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/transactions?address=%s", "0xc0ffee254729296a45a3885639AC7E10F9d54979"), nil),
				transactions: []parser.EthTransaction{
					{Hash: "0x1", From: "0xc0ffee254729296a45a3885639AC7E10F9d54979", To: "0x456", Value: "100"},
					{Hash: "0x2", From: "0xc0ffee254729296a45a3885639AC7E10F9d54979", To: "0x456", Value: "101"},
					{Hash: "0x3", From: "0x456", To: "0xc0ffee254729296a45a3885639AC7E10F9d54979", Value: "102"},
					{Hash: "0x4", From: "0x386", To: "0xc0ffee254729296a45a3885639AC7E10F9d54979", Value: "102"},
				},
				codeWant: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rr = httptest.NewRecorder()
			txStore := store.NewMemTxStore[parser.EthTransaction]()
			h := &Handler{
				logger:      tt.fields.logger,
				txParser:    parser.NewEthTxParser(txStore, &http.Client{}, tt.fields.logger, 0),
				httpTimeout: time.Duration(tt.fields.httpTimeout) * time.Second,
			}
			handler := http.HandlerFunc(h.handleGetTransactions)
			h.txParser.Subscribe(tt.args.address)
			h.txParser.(*parser.EthTxParser).UpdateTransactionsInStore(tt.args.transactions)
			handler.ServeHTTP(tt.fields.rr, tt.args.r)
			if tt.fields.rr.Code != tt.args.codeWant {
				t.Errorf("Handler.handleGetTransactions() = %v, want %v", tt.fields.rr.Code, http.StatusOK)
			}
			if tt.fields.rr.Body.Len() == 0 {
				t.Errorf("Handler.handleGetTransactions() = %v, want %v", tt.fields.rr.Body.Len(), 0)
			}
			var txs []parser.EthTransaction
			err := json.NewDecoder(tt.fields.rr.Body).Decode(&txs)
			if err != nil {
				t.Errorf("Handler.handleGetTransactions() error = %v", err)
			}
			if len(txs) != len(tt.args.transactions) {
				t.Errorf("Handler.handleGetTransactions() = %v, want %v", len(txs), len(tt.args.transactions))
			}
		})
	}
}

func TestHandler_handleGetTransactionsUntrackedAddress(t *testing.T) {
	type fields struct {
		logger      *slog.Logger
		httpTimeout int
		rr          *httptest.ResponseRecorder
	}
	type args struct {
		w             http.ResponseWriter
		r             *http.Request
		transactions  []parser.EthTransaction
		address       string
		untrackedAddr string
		codeWant      int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test handleGetTransactions Untracked Address",
			fields: fields{
				logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
				httpTimeout: 5,
			},
			args: args{
				address: "0xc0ffee254729296a45a3885639AC7E10F9d54979",
				w:       httptest.NewRecorder(),
				r:       httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/transactions?address=%s", "0x999999cf1046e68e36E1aA2E0E07105eDDD1f08E"), nil),
				transactions: []parser.EthTransaction{
					{Hash: "0x1", From: "0xc0ffee254729296a45a3885639AC7E10F9d54979", To: "0x456", Value: "100"},
					{Hash: "0x2", From: "0xc0ffee254729296a45a3885639AC7E10F9d54979", To: "0x456", Value: "101"},
					{Hash: "0x3", From: "0x456", To: "0xc0ffee254729296a45a3885639AC7E10F9d54979", Value: "102"},
					{Hash: "0x4", From: "0x386", To: "0xc0ffee254729296a45a3885639AC7E10F9d54979", Value: "102"},
				},
				codeWant:      http.StatusNotFound,
				untrackedAddr: "0x999999cf1046e68e36E1aA2E0E07105eDDD1f08E",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rr = httptest.NewRecorder()
			txStore := store.NewMemTxStore[parser.EthTransaction]()
			h := &Handler{
				logger:      tt.fields.logger,
				txParser:    parser.NewEthTxParser(txStore, &http.Client{}, tt.fields.logger, 0),
				httpTimeout: time.Duration(tt.fields.httpTimeout) * time.Second,
			}
			handler := http.HandlerFunc(h.handleGetTransactions)
			h.txParser.Subscribe(tt.args.address)
			h.txParser.(*parser.EthTxParser).UpdateTransactionsInStore(tt.args.transactions)
			handler.ServeHTTP(tt.fields.rr, tt.args.r)
			if tt.fields.rr.Code != tt.args.codeWant {
				t.Errorf("Handler.handleGetTransactionsUntrackedAddress() = %v, want %v", tt.fields.rr.Code, http.StatusOK)
			}
		})
	}
}

func TestHandler_handleGetTransactionsBadInput(t *testing.T) {
	type fields struct {
		logger      *slog.Logger
		httpTimeout int
		rr          *httptest.ResponseRecorder
	}
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		codeWant int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test handleGetTransactions Bad Input",
			fields: fields{
				logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
				httpTimeout: 5,
			},
			args: args{
				w:        httptest.NewRecorder(),
				r:        httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/transactions?address=%s", "0xc0ffee254729296a"), nil),
				codeWant: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rr = httptest.NewRecorder()
			h := &Handler{
				logger:      tt.fields.logger,
				txParser:    nil,
				httpTimeout: time.Duration(tt.fields.httpTimeout) * time.Second,
			}
			handler := http.HandlerFunc(h.handleGetTransactions)
			handler.ServeHTTP(tt.fields.rr, tt.args.r)
			if tt.fields.rr.Code != tt.args.codeWant {
				t.Errorf("Handler.handleGetTransactionsBadInput() = %v, want %v", tt.fields.rr.Code, http.StatusOK)
			}
		})
	}
}

func TestHandler_handleSubscribe(t *testing.T) {
	type Address struct {
		Address string `json:"address"`
	}
	type fields struct {
		logger      *slog.Logger
		httpTimeout int
		rr          *httptest.ResponseRecorder
	}
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		address  Address
		body     string
		codeWant int
	}
	add := Address{"0xc0ffee254729296a45a3885639AC7E10F9d54979"}
	reqBody, _ := json.Marshal(add)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test handleSubscribe",
			fields: fields{
				logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
				httpTimeout: 5,
			},
			args: args{
				w:        httptest.NewRecorder(),
				r:        httptest.NewRequest(http.MethodGet, "/v1/subscribe", bytes.NewBuffer(reqBody)),
				address:  add,
				codeWant: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rr = httptest.NewRecorder()
			h := &Handler{
				logger:      tt.fields.logger,
				txParser:    parser.NewEthTxParser(store.NewMemTxStore[parser.EthTransaction](), &http.Client{}, tt.fields.logger, 0),
				httpTimeout: time.Duration(tt.fields.httpTimeout) * time.Second,
			}
			handler := http.HandlerFunc(h.handleSubscribeAddress)
			handler.ServeHTTP(tt.fields.rr, tt.args.r)
			if tt.fields.rr.Code != tt.args.codeWant {
				t.Errorf("Handler.handleSubscribe() = %v, want %v", tt.fields.rr.Code, http.StatusOK)
			}
		})
	}
}

func TestHandler_handleSubscribeBadInput(t *testing.T) {
	type fields struct {
		logger      *slog.Logger
		httpTimeout int
		rr          *httptest.ResponseRecorder
	}
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		address  string
		codeWant int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test handleSubscribe Bad Input",
			fields: fields{
				logger:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
				httpTimeout: 5,
			},
			args: args{
				w:        httptest.NewRecorder(),
				r:        httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/subscribe?address=%s", "0xc0ffee254729296a4"), nil),
				address:  "0xc0ffee254729296a4",
				codeWant: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rr = httptest.NewRecorder()
			h := &Handler{
				logger:      tt.fields.logger,
				txParser:    parser.NewEthTxParser(store.NewMemTxStore[parser.EthTransaction](), &http.Client{}, tt.fields.logger, 0),
				httpTimeout: time.Duration(tt.fields.httpTimeout) * time.Second,
			}
			handler := http.HandlerFunc(h.handleSubscribeAddress)
			handler.ServeHTTP(tt.fields.rr, tt.args.r)
			if tt.fields.rr.Code != tt.args.codeWant {
				t.Errorf("Handler.handleSubscribeBadInput() = %v, want %v", tt.fields.rr.Code, http.StatusOK)
			}
		})
	}
}
