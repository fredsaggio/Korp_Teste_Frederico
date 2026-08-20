package server

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *http.Server {
	server := &Server{
		db: db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)

	return &http.Server{
		Addr:              ":5001",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
