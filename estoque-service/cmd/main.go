package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"korp/estoque-service/internal/db"
	"korp/estoque-service/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	connStr := os.Getenv("ESTOQUE_DATABASE_URL")
	if connStr == "" {
		return errors.New("environment variable ESTOQUE_DATABASE_URL is required")
	}

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	httpServer := server.New()

	log.Print("Estoque service running on port 5001")

	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running HTTP server: %w", err)
	}

	return nil
}
