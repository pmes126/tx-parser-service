package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pmes126/tx-parser-service/api/handler"
	"github.com/pmes126/tx-parser-service/internal/store"
	"github.com/pmes126/tx-parser-service/pkg/parser"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	run(context.Background(), logger)
}

func run(ctx context.Context, logger *slog.Logger) error {
	_ = ctx
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(shutdown)

	ethTxParser := parser.NewEthTxParser(store.NewMemTxStore[parser.EthTransaction](), logger)
	block, _ := ethTxParser.GetCurrentBlock()
	fmt.Println("Current block:", block)
	ethTxParser.UpdateTransactionsFromBlock(block)

	logger.Info("Starting tx-parser-service")
	//txParser.Start(ctx)
	//defer txParser.Stop()

	// Construct an HTTP server to service requests.
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", 8080),
		Handler:      handler.Routes(&handler.Handler{txStore: store.NewMemTxStore[parser.EthTransaction](), logger: logger, txParser: ethTxParser, httpTimeout: 5 * time.Second}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
		ErrorLog:     &log.Logger{},
	}
	serverErr := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("error starting server: %w", err)
	case <-shutdown:
		// Shutdown the server gracefully.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			if err := server.Close(); err != nil {
				return fmt.Errorf("Could not close server gracefully: %w", err)
			}
			return fmt.Errorf("Could not shutdown server gracefully: %w", err)
		}
	}
	return nil
}
