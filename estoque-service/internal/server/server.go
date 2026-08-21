package server

import (
	"net/http"
	"time"

	"korp/estoque-service/internal/products"
	"korp/estoque-service/internal/stock"
)

type Handlers struct {
	ProductHandler *products.ProductHandler
	DebitHandler   *stock.DebitHandler
}

type Server struct {
	handlers Handlers
}

func New(handlers Handlers) *http.Server {
	server := &Server{handlers: handlers}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /api/v1/products", server.handlers.ProductHandler.Create)
	mux.HandleFunc("GET /api/v1/products", server.handlers.ProductHandler.List)
	mux.HandleFunc("POST /api/v1/stock/debits", server.handlers.DebitHandler.Debit)

	return &http.Server{
		Addr:              ":5001",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
