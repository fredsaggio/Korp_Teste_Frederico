package products

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidInput      = errors.New("invalid product input")
	ErrCodeAlreadyExists = errors.New("product code already exists")
)

type Product struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Balance     int64     `json:"balance"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductInput struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Balance     int64  `json:"balance"`
}

type ProductStore interface {
	Create(ctx context.Context, input ProductInput) (*Product, error)
	List(ctx context.Context) ([]Product, error)
}

type ProductService interface {
	Create(ctx context.Context, input ProductInput) (*Product, error)
	List(ctx context.Context) ([]Product, error)
}
