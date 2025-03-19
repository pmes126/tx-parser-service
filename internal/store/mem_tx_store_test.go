package store

import (
	"testing"
)

type Transaction struct {
	Hash  string
	From  string
	To    string
	Value string
}

func TestMemTxStore_AddGetTransactions(t *testing.T) {
	type args struct {
		address string
		tx      []Transaction
	}
	tests := []struct {
		name    string
		wantErr bool
		args    args
	}{
		{
			name:    "Test AddGetTransactions",
			wantErr: false,
			args: args{"0x123",
				// TODO: Randomly generate transactions in large numbers
				[]Transaction{
					{Hash: "0x1", From: "0x123", To: "0x456", Value: "100"},
					{Hash: "0x2", From: "0x123", To: "0x456", Value: "101"},
					{Hash: "0x3", From: "0x456", To: "0x123", Value: "102"},
					{Hash: "0x4", From: "0x386", To: "0x123", Value: "102"},
					{Hash: "0x5", From: "0x486", To: "0x123", Value: "101"},
					{Hash: "0x6", From: "0x586", To: "0x123", Value: "104"},
					{Hash: "0x7", From: "0x123", To: "0x686", Value: "105"},
				},
			},
		},
		{
			name:    "Test AddTransaction",
			wantErr: true,
			args: args{"0x124",
				[]Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mts := NewMemTxStore[Transaction]()
			for _, tx := range tt.args.tx {
				err := mts.AddTransaction(tt.args.address, tx)
				if (err != nil) != tt.wantErr {
					t.Errorf("MemTxStore.AddTransaction() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			tx, err := mts.GetTransactions(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemTxStore.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tx) != len(tt.args.tx) {
				t.Errorf("MemTxStore.GetTransactions() = %v, want %v", len(tx), len(tt.args.tx))
			}
		})
	}
}
