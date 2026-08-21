package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"korp/faturamento-service/internal/db"
	"korp/faturamento-service/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	connectionString := os.Getenv("FATURAMENTO_DATABASE_URL")
	if connectionString == "" {
		return errors.New("environment variable FATURAMENTO_DATABASE_URL is required")
	}

	pool, err := db.Connect(ctx, connectionString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	httpServer := server.New()

	log.Print("Faturamento service running on port 5002")
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running HTTP server: %w", err)
	}

	return nil
}
