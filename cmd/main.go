package main

import (
	"context"
	"fmt"
	"github.com/pmes126/tx-parser-service/parser"
	"os"
	"os/signal"
	"syscall"
)

func main() {}

func run() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(shutdown)

	txParser := parser.NewEthTxParser()
	//txParser.Start()
	block, _ := txParser.GetCurrentBlock()
	txParser.UpdateTransactionsFromBlock(block)
	//defer txParser.Stop()

	go func() {
		<-shutdown
		os.Exit(1)
	}()
}
