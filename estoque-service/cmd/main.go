package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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

	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("error to access database: %w", err)
	}

	defer dbpool.Close()

	if err := dbpool.Ping(ctx); err != nil {
		return fmt.Errorf("Error to access database: %w", err)
	}

	httpServer := server.New(dbpool)

	log.Printf("Estoque service running on port 5001!")

	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running HTTP server: %w", err)
	}

	return nil
}
