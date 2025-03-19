package handler

import (
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//switch r.Method {
	//case http.MethodGet:
	//	h.handleGet(w, r)
	//case http.MethodPost:
	//	h.handlePost(w, r)
	//default:
	//	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//}
}

func NewHandler(logger *slog.Logger, txParser parser.Parser, httpTimeout time.Duration) *Handler {
	return &Handler{
		logger:      logger,
		txParser:    txParser,
		httpTimeout: httpTimeout,
	}
}

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

func (h *Handler) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter missing", http.StatusBadRequest)
		return
	}
	txs, err := h.txParser.GetTransactions(address)
	if err != nil {
		if errors.Is(err, store.ErrNoTransactions) {
			http.Error(w, "No transactions found for address", http.StatusNotFound)
			return
		} else if errors.Is(err, store.ErrAddressNotFound) {
			http.Error(w, "Address not Tracked", http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	json.NewEncoder(w).Encode(txs)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleSubscribeAddress(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter missing", http.StatusBadRequest)
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
