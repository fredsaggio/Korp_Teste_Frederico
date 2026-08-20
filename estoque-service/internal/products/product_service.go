package products

import (
	"context"
	"fmt"
	"strings"
)

type productService struct {
	store ProductStore
}

func NewProductService(store ProductStore) ProductService {
	return &productService{store: store}
}

func (s *productService) Create(ctx context.Context, input ProductInput) (*Product, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Description = strings.TrimSpace(input.Description)

	if input.Code == "" {
		return nil, fmt.Errorf("%w: code is required", ErrInvalidInput)
	}
	if input.Description == "" {
		return nil, fmt.Errorf("%w: description is required", ErrInvalidInput)
	}
	if input.Balance < 0 {
		return nil, fmt.Errorf("%w: balance cannot be negative", ErrInvalidInput)
	}

	product, err := s.store.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

func (s *productService) List(ctx context.Context) ([]Product, error) {
	products, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	return products, nil
}
