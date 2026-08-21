package invoices

import (
	"context"
	"fmt"
	"sort"
)

type invoiceService struct {
	store InvoiceStore
}

func NewInvoiceService(store InvoiceStore) InvoiceService {
	return &invoiceService{store: store}
}

func (s *invoiceService) Create(ctx context.Context, input CreateInput) (*Invoice, error) {
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one item is required", ErrInvalidInput)
	}

	items := append([]InvoiceItem(nil), input.Items...)
	seenProducts := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item.ProductID <= 0 {
			return nil, fmt.Errorf("%w: product ID must be positive", ErrInvalidInput)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: quantity must be positive", ErrInvalidInput)
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, fmt.Errorf("%w: product %d is duplicated", ErrInvalidInput, item.ProductID)
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})
	input.Items = items

	invoice, err := s.store.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	return invoice, nil
}

func (s *invoiceService) List(ctx context.Context) ([]Invoice, error) {
	invoices, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}

	return invoices, nil
}

func (s *invoiceService) GetByNumber(ctx context.Context, number int64) (*Invoice, error) {
	if number <= 0 {
		return nil, fmt.Errorf("%w: invoice number must be positive", ErrInvalidInput)
	}

	invoice, err := s.store.GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}

	return invoice, nil
}
