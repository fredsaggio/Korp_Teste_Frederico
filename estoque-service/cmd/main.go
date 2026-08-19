package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	connStr := os.Getenv("ESTOQUE_DATABASE_URL")
	dbpool, err := pgxpool.New(ctx, connStr)

	if err != nil {
		log.Fatal("Error to create connection: ", err)
	}

	defer dbpool.Close()

	if err := dbpool.Ping(ctx); err != nil {
		log.Fatal("Error to access database: ", err)
	}

	log.Printf("Connected to PostgreSQL")
}
