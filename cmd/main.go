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

	"github.com/spf13/viper"

	"github.com/pmes126/tx-parser-service/api/handler"
	"github.com/pmes126/tx-parser-service/internal/store"
	"github.com/pmes126/tx-parser-service/pkg/parser"
)

type Config struct {
	HTTPPort     int `mapstructure:"http_port"`
	ReadTimeout  int `mapstructure:"readTimeout"`
	WriteTimeout int `mapstructure:"writeTimeout"`
	IdleTimeout  int `mapstructure:"idleTimeout"`
	HTTPTimeout  int `mapstructure:"httpTimeout"`
	PollInterval int `mapstructure:"pollInterval"`
	WorkerCount  int `mapstructure:"workerCount"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := initConfig(); err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Config: %+v\n", cfg)
	if err := run(context.Background(), logger, &cfg); err != nil {
		log.Fatal(err)
	}
}

func run(c context.Context, logger *slog.Logger, cfg *Config) error {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(shutdown)

	ethTxParser := parser.NewEthTxParser(store.NewMemTxStore[parser.EthTransaction](), logger, cfg.PollInterval)
	block, _ := ethTxParser.GetCurrentBlock()
	fmt.Println("Current block:", block)
	ethTxParser.UpdateTransactionsFromBlock(block)

	go func() {
		logger.Info("Starting tx-parser-service")
		ethTxParser.Start(ctx)
	}()

	// Construct an HTTP server to service requests.
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      handler.Routes(handler.NewHandler(logger, ethTxParser, 5*time.Second)),
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		ErrorLog:     &log.Logger{},
	}
	serverErr := make(chan error, 1)

	go func() {
		logger.Info("Starting tx-parser-service")
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-shutdown:
		// Cancel existing context.
		cancel()
		// Shutdown the server gracefully.
		sctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := server.Shutdown(sctx); err != nil {
			if err := server.Close(); err != nil {
				return fmt.Errorf("Could not close server gracefully: %w", err)
			}
			return fmt.Errorf("Could not shutdown server gracefully: %w", err)
		}
		return nil
	}
}

func initConfig() error {
	// Load configuration from environment variables.
	viper.SetConfigName("config")
	viper.AddConfigPath("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	return viper.ReadInConfig()
}
