package server

import (
	"net/http"
	"time"
)

func New() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

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
