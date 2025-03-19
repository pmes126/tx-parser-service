package parser

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/pmes126/tx-parser-service/internal/store"
)

func TestEthTxParser_GetCurrentBlock(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tests := []struct {
		name    string
		want    int64
		notWant int64
		wantErr bool
	}{
		{
			name:    "Test GetCurrentBlock",
			notWant: 0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etp := NewEthTxParser(store.NewMemTxStore[EthTransaction](), &http.Client{}, logger, 0)
			got, err := etp.GetCurrentBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("EthTxParser.GetCurrentBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == tt.notWant {
				t.Errorf("EthTxParser.GetCurrentBlock() = %v, notWant %v", got, tt.want)
			}
		})
	}
}

func TestEthTxParser_QueryTransactionsByBlock(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	type args struct {
		block int64
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Test QueryTransactionsByBlock0",
			args:    args{block: 0}, // test genesis block
			wantErr: false,
			want:    0,
		},
		{
			name:    "Test QueryTransactionsByBlock1",
			args:    args{block: 10000000},
			wantErr: false,
			want:    103,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etp := NewEthTxParser(store.NewMemTxStore[EthTransaction](), &http.Client{}, logger, 0)
			tx, err := etp.QueryTransactionsFromBlock(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("EthTxParser.QueryTransactionsByBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tx) != tt.want {
				t.Errorf("EthTxParser.QueryTransactionsByBlock() = %v, want %v", len(tx), tt.want)
			}
		})
	}
}

func TestEthTxParser_UpdateAndGetTransactions(t *testing.T) {
	type args struct {
		transactions []EthTransaction
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	Addresses := []string{"0x123", "0x456", "0x789"}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test UpdateTransactionsInStore",
			args: args{transactions: []EthTransaction{
				{
					From:  "0x123",
					To:    "0x456",
					Hash:  "0x1",
					Value: "0x123",
				},
				{
					From:  "0x123",
					To:    "0x456",
					Hash:  "0x2",
					Value: "0x123",
				},
				{
					From:  "0x456",
					To:    "0x789",
					Hash:  "0x3",
					Value: "0x123",
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etp := NewEthTxParser(store.NewMemTxStore[EthTransaction](), &http.Client{}, logger, 0)
			for _, address := range Addresses {
				etp.Subscribe(address)
			}
			if err := etp.UpdateTransactionsInStore(tt.args.transactions); (err != nil) != tt.wantErr {
				t.Errorf("EthTxParser.UpdateTransactionsInStore() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, address := range Addresses {
				tx, err := etp.GetTransactions(address)
				if (err != nil) != tt.wantErr {
					t.Errorf("EthTxParser.UpdateTransactionsInStore() error = %v, wantErr %v", err, tt.wantErr)
				}
				for _, tr := range tx {
					if tr.From != address && tr.To != address {
						t.Errorf("EthTxParser.UpdateTransactionsInStore() Tx not found for addr %s", address)
					}
				}
			}
		})
	}
}

