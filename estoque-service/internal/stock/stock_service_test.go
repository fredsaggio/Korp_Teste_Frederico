package stock

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type debitStoreStub struct {
	debit func(context.Context, DebitInput) (*DebitResult, error)
}

func (s debitStoreStub) Debit(ctx context.Context, input DebitInput) (*DebitResult, error) {
	return s.debit(ctx, input)
}

func TestDebitServiceNormalizesAndSortsInput(t *testing.T) {
	originalItems := []DebitItem{
		{ProductID: 2, Quantity: 1},
		{ProductID: 1, Quantity: 3},
	}
	wantInput := DebitInput{
		Reference: "invoice:123",
		Items: []DebitItem{
			{ProductID: 1, Quantity: 3},
			{ProductID: 2, Quantity: 1},
		},
	}
	wantResult := &DebitResult{Reference: "invoice:123"}

	store := debitStoreStub{
		debit: func(_ context.Context, input DebitInput) (*DebitResult, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Debit input = %#v, want %#v", input, wantInput)
			}
			return wantResult, nil
		},
	}
	service := NewDebitService(store)

	got, err := service.Debit(context.Background(), DebitInput{
		Reference: "  invoice:123  ",
		Items:     originalItems,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, wantResult) {
		t.Fatalf("Debit() = %#v, want %#v", got, wantResult)
	}
	if originalItems[0].ProductID != 2 {
		t.Fatal("Debit() modified the caller's item slice")
	}
}

func TestDebitServiceRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input DebitInput
	}{
		{name: "missing reference", input: DebitInput{Items: []DebitItem{{ProductID: 1, Quantity: 1}}}},
		{name: "blank reference", input: DebitInput{Reference: "  ", Items: []DebitItem{{ProductID: 1, Quantity: 1}}}},
		{name: "missing items", input: DebitInput{Reference: "invoice:123"}},
		{name: "invalid product ID", input: DebitInput{Reference: "invoice:123", Items: []DebitItem{{ProductID: 0, Quantity: 1}}}},
		{name: "invalid quantity", input: DebitInput{Reference: "invoice:123", Items: []DebitItem{{ProductID: 1, Quantity: 0}}}},
		{
			name:  "duplicated product",
			input: DebitInput{Reference: "invoice:123", Items: []DebitItem{{ProductID: 1, Quantity: 1}, {ProductID: 1, Quantity: 2}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeCalled := false
			store := debitStoreStub{
				debit: func(context.Context, DebitInput) (*DebitResult, error) {
					storeCalled = true
					return nil, nil
				},
			}

			_, err := NewDebitService(store).Debit(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Debit() error = %v, want ErrInvalidInput", err)
			}
			if storeCalled {
				t.Fatal("Debit() called store with invalid input")
			}
		})
	}
}

func TestDebitServiceWrapsStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	store := debitStoreStub{
		debit: func(context.Context, DebitInput) (*DebitResult, error) {
			return nil, storeErr
		},
	}

	_, err := NewDebitService(store).Debit(context.Background(), DebitInput{
		Reference: "invoice:123",
		Items:     []DebitItem{{ProductID: 1, Quantity: 1}},
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("Debit() error = %v, want wrapped store error", err)
	}
}
