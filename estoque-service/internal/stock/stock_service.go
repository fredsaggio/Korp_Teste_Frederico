package stock

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type debitService struct {
	store DebitStore
}

func NewDebitService(store DebitStore) DebitService {
	return &debitService{store: store}
}

func (s *debitService) Debit(ctx context.Context, input DebitInput) (*DebitResult, error) {
	input.Reference = strings.TrimSpace(input.Reference)
	if input.Reference == "" {
		return nil, fmt.Errorf("%w: reference is required", ErrInvalidInput)
	}
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one item is required", ErrInvalidInput)
	}

	items := append([]DebitItem(nil), input.Items...)
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

	result, err := s.store.Debit(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("debit stock: %w", err)
	}

	return result, nil
}
