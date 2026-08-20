package products

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type productStoreStub struct {
	create func(context.Context, ProductInput) (*Product, error)
	list   func(context.Context) ([]Product, error)
}

func (s productStoreStub) Create(ctx context.Context, input ProductInput) (*Product, error) {
	return s.create(ctx, input)
}

func (s productStoreStub) List(ctx context.Context) ([]Product, error) {
	return s.list(ctx)
}

func TestProductServiceCreateNormalizesInput(t *testing.T) {
	wantInput := ProductInput{
		Code:        "PROD-001",
		Description: "Product one",
		Balance:     10,
	}
	wantProduct := &Product{
		ID:          1,
		Code:        wantInput.Code,
		Description: wantInput.Description,
		Balance:     wantInput.Balance,
	}

	store := productStoreStub{
		create: func(_ context.Context, input ProductInput) (*Product, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create input = %#v, want %#v", input, wantInput)
			}
			return wantProduct, nil
		},
	}
	service := NewProductService(store)

	got, err := service.Create(context.Background(), ProductInput{
		Code:        "  PROD-001  ",
		Description: "  Product one  ",
		Balance:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, wantProduct) {
		t.Fatalf("Create() = %#v, want %#v", got, wantProduct)
	}
}

func TestProductServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input ProductInput
	}{
		{name: "missing code", input: ProductInput{Description: "Product", Balance: 1}},
		{name: "blank code", input: ProductInput{Code: "  ", Description: "Product", Balance: 1}},
		{name: "missing description", input: ProductInput{Code: "PROD-001", Balance: 1}},
		{name: "blank description", input: ProductInput{Code: "PROD-001", Description: "  ", Balance: 1}},
		{name: "negative balance", input: ProductInput{Code: "PROD-001", Description: "Product", Balance: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeCalled := false
			store := productStoreStub{
				create: func(context.Context, ProductInput) (*Product, error) {
					storeCalled = true
					return nil, nil
				},
			}

			_, err := NewProductService(store).Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
			}
			if storeCalled {
				t.Fatal("Create() called store with invalid input")
			}
		})
	}
}

func TestProductServiceCreateWrapsStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	store := productStoreStub{
		create: func(context.Context, ProductInput) (*Product, error) {
			return nil, storeErr
		},
	}

	_, err := NewProductService(store).Create(context.Background(), ProductInput{
		Code:        "PROD-001",
		Description: "Product",
		Balance:     1,
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("Create() error = %v, want wrapped store error", err)
	}
}

func TestProductServiceList(t *testing.T) {
	want := []Product{{ID: 1, Code: "PROD-001", Description: "Product", Balance: 1}}
	store := productStoreStub{
		list: func(context.Context) ([]Product, error) {
			return want, nil
		},
	}

	got, err := NewProductService(store).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
}
