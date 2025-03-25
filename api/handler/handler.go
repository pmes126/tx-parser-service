package handler

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pmes126/tx-parser-service/internal/store"
	"github.com/pmes126/tx-parser-service/pkg/parser"
)

// Handler service
type Handler struct {
	logger      *slog.Logger
	txParser    parser.Parser
	httpTimeout time.Duration
}

// NewHandler creates a new handler
func NewHandler(logger *slog.Logger, txParser parser.Parser, httpTimeout time.Duration) *Handler {
	return &Handler{
		logger:      logger,
		txParser:    txParser,
		httpTimeout: httpTimeout,
	}
}

// Routes returns the router for the handler
func Routes(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(h.httpTimeout))
	r.Route("/v1", func(r chi.Router) {
		r.Get("/transactions", h.handleGetTransactions)
		r.Post("/subscribe", h.handleSubscribeAddress)
	})
	return r
}

// handleGetTransactions godoc
// @Summary Get transactions for an address
// @Description Get transactions for an address
// @Produce json
// @Param address query string true "Address to get transactions for"
// @Success 200 {array} EthTransaction
// @Failure 400 {string} string "Address parameter missing"
// @Failure 400 {string} string "Invalid address"
// @Failure 404 {string} string "Address not tracked"
// @Failure 404 {string} string "Transactions not found"
// @Failure 500 {string} string
// @Router /v1/transactions [get]
func (h *Handler) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter missing", http.StatusBadRequest)
		return
	}
	if !isValidEthAddress(address) {
		http.Error(w, "Invalid address", http.StatusBadRequest)
		return
	}
	txs, err := h.txParser.GetTransactions(address)
	if err != nil {
		if errors.Is(err, store.ErrNoTransactions) {
			http.Error(w, "No transactions found for address", http.StatusNotFound)
			return
		} else if errors.Is(err, parser.ErrAddressNotTracked) {
			http.Error(w, "Address not Tracked", http.StatusNotFound)
			return
		} else {
			h.logger.Error("Failed to get transactions for address %s: %v", address, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	json.NewEncoder(w).Encode(txs)
	w.WriteHeader(http.StatusOK)
}

// handleSubscribeAddress godoc
// @Summary Subscribe to an address
// @Description Subscribe to an address to receive notifications of transactions
// @Tags subscribe
// @Param address query string true "Address to subscribe to"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Address parameter missing"
// @Failure 400 {string} string "Invalid address"
// @Failure 500 {string} string "Failed to subscribe to address"
// @Router /v1/subscribe [post]
func (h *Handler) handleSubscribeAddress(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter missing", http.StatusBadRequest)
		return
	}
	if !isValidEthAddress(address) {
		http.Error(w, "Invalid address", http.StatusBadRequest)
		return
	}
	if h.txParser.Subscribe(address) {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Failed to subscribe to address", http.StatusInternalServerError)
		return
	}
}

func isValidEthAddress(address string) bool {
	if len(address) != parser.EthAddressLength || address[:2] != "0x" {
		return false
	}
	if _, err := hex.DecodeString(address[2:]); err != nil {
		return false
	}
	return true
}
