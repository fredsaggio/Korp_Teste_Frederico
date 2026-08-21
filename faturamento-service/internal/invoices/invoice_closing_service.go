package invoices

import (
	"context"
	"fmt"

	"korp/faturamento-service/internal/stock"
)

type invoiceClosingService struct {
	store       InvoiceStore
	stockClient stock.Client
}

func NewInvoiceClosingService(store InvoiceStore, stockClient stock.Client) InvoiceClosingService {
	return &invoiceClosingService{store: store, stockClient: stockClient}
}

func (s *invoiceClosingService) Close(ctx context.Context, number int64) (*Invoice, error) {
	if number <= 0 {
		return nil, fmt.Errorf("%w: invoice number must be positive", ErrInvalidInput)
	}

	invoice, err := s.store.GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("get invoice before closing: %w", err)
	}
	if invoice.Status != StatusOpen {
		return nil, ErrInvoiceNotOpen
	}

	items := make([]stock.Item, len(invoice.Items))
	for i, item := range invoice.Items {
		items[i] = stock.Item{ProductID: item.ProductID, Quantity: item.Quantity}
	}
	if err := s.stockClient.Debit(ctx, stock.DebitInput{
		Reference: fmt.Sprintf("invoice:%d", invoice.Number),
		Items:     items,
	}); err != nil {
		return nil, fmt.Errorf("debit invoice stock: %w", err)
	}

	closedAt, err := s.store.Close(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("persist closed invoice: %w", err)
	}

	closedInvoice := *invoice
	closedInvoice.Status = StatusClosed
	closedInvoice.UpdatedAt = closedAt
	closedInvoice.ClosedAt = &closedAt
	return &closedInvoice, nil
}
