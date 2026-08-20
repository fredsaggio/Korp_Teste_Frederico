package products

import (
	"context"
	"fmt"

	"korp/estoque-service/internal/db"

	"github.com/jackc/pgx/v5"
)

const productCodeUniqueConstraint = "products_code_key"

type productStore struct {
	db db.DB
}

func NewProductStore(database db.DB) ProductStore {
	return &productStore{db: database}
}

func (s *productStore) Create(ctx context.Context, input ProductInput) (*Product, error) {
	const query = `
		INSERT INTO products (code, description, balance)
		VALUES (@code, @description, @balance)
		RETURNING id, code, description, balance, created_at, updated_at
	`
	args := pgx.StrictNamedArgs{
		"code":        input.Code,
		"description": input.Description,
		"balance":     input.Balance,
	}

	var product Product
	err := s.db.QueryRow(ctx, query, args).Scan(
		&product.ID,
		&product.Code,
		&product.Description,
		&product.Balance,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if db.IsUniqueViolation(err, productCodeUniqueConstraint) {
			return nil, ErrCodeAlreadyExists
		}
		return nil, fmt.Errorf("insert product: %w", err)
	}

	return &product, nil
}

func (s *productStore) List(ctx context.Context) ([]Product, error) {
	const query = `
		SELECT id, code, description, balance, created_at, updated_at
		FROM products
		ORDER BY id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Code,
			&product.Description,
			&product.Balance,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}

	return products, nil
}
