package invoices

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidInput    = errors.New("invalid invoice input")
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrInvoiceNotOpen  = errors.New("invoice is not open")
)

type Status string

const (
	StatusOpen   Status = "OPEN"
	StatusClosed Status = "CLOSED"
)

type InvoiceItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type Invoice struct {
	Number    int64         `json:"number"`
	Status    Status        `json:"status"`
	Items     []InvoiceItem `json:"items"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	ClosedAt  *time.Time    `json:"closed_at"`
}

type CreateInput struct {
	Items []InvoiceItem `json:"items"`
}

type InvoiceStore interface {
	Create(ctx context.Context, input CreateInput) (*Invoice, error)
	List(ctx context.Context) ([]Invoice, error)
	GetByNumber(ctx context.Context, number int64) (*Invoice, error)
	Close(ctx context.Context, number int64) (time.Time, error)
}

type InvoiceClosingService interface {
	Close(ctx context.Context, number int64) (*Invoice, error)
}

type InvoiceService interface {
	Create(ctx context.Context, input CreateInput) (*Invoice, error)
	List(ctx context.Context) ([]Invoice, error)
	GetByNumber(ctx context.Context, number int64) (*Invoice, error)
}
