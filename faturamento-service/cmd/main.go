package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"korp/faturamento-service/internal/db"
	"korp/faturamento-service/internal/server"
	"korp/faturamento-service/internal/stock"
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
	stockServiceURL := os.Getenv("ESTOQUE_SERVICE_URL")
	if stockServiceURL == "" {
		return errors.New("environment variable ESTOQUE_SERVICE_URL is required")
	}
	stockClient, err := stock.NewHTTPClient(stockServiceURL, 3*time.Second)
	if err != nil {
		return fmt.Errorf("configure stock service client: %w", err)
	}

	pool, err := db.Connect(ctx, connectionString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	handlers := buildHandlers(pool, stockClient)
	httpServer := server.New(handlers)

	log.Print("Faturamento service running on port 5002")
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running HTTP server: %w", err)
	}

	return nil
}
