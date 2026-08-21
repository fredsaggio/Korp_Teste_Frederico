package server

import (
	"net/http"
	"time"

	"korp/faturamento-service/internal/invoices"
)

type Handlers struct {
	InvoiceHandler *invoices.InvoiceHandler
}

type Server struct {
	handlers Handlers
}

func New(handlers Handlers) *http.Server {
	server := &Server{handlers: handlers}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /api/v1/invoices", server.handlers.InvoiceHandler.Create)
	mux.HandleFunc("GET /api/v1/invoices", server.handlers.InvoiceHandler.List)
	mux.HandleFunc("GET /api/v1/invoices/{number}", server.handlers.InvoiceHandler.GetByNumber)

	return &http.Server{
		Addr:              ":5002",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
