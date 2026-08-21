package stock

import (
	"context"
	"errors"
)

var (
	ErrInvalidInput        = errors.New("invalid stock debit input")
	ErrIdempotencyConflict = errors.New("stock debit reference already used with different items")
	ErrProductNotFound     = errors.New("product not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
)

type DebitItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type DebitInput struct {
	Reference string      `json:"reference"`
	Items     []DebitItem `json:"items"`
}

type DebitResult struct {
	Reference        string `json:"reference"`
	AlreadyProcessed bool   `json:"already_processed"`
}

type DebitStore interface {
	Debit(ctx context.Context, input DebitInput) (*DebitResult, error)
}

type DebitService interface {
	Debit(ctx context.Context, input DebitInput) (*DebitResult, error)
}
