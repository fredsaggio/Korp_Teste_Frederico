package server

import (
	"net/http"
	"time"
)

func New() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

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
