package invoices

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type invoiceStoreStub struct {
	create      func(context.Context, CreateInput) (*Invoice, error)
	list        func(context.Context) ([]Invoice, error)
	getByNumber func(context.Context, int64) (*Invoice, error)
}

func (s invoiceStoreStub) Create(ctx context.Context, input CreateInput) (*Invoice, error) {
	return s.create(ctx, input)
}

func (s invoiceStoreStub) List(ctx context.Context) ([]Invoice, error) {
	return s.list(ctx)
}

func (s invoiceStoreStub) GetByNumber(ctx context.Context, number int64) (*Invoice, error) {
	return s.getByNumber(ctx, number)
}

func TestInvoiceServiceCreateSortsItems(t *testing.T) {
	originalItems := []InvoiceItem{
		{ProductID: 2, Quantity: 1},
		{ProductID: 1, Quantity: 3},
	}
	wantInput := CreateInput{Items: []InvoiceItem{
		{ProductID: 1, Quantity: 3},
		{ProductID: 2, Quantity: 1},
	}}
	wantInvoice := &Invoice{Number: 1, Status: StatusOpen, Items: wantInput.Items}

	store := invoiceStoreStub{
		create: func(_ context.Context, input CreateInput) (*Invoice, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create input = %#v, want %#v", input, wantInput)
			}
			return wantInvoice, nil
		},
	}

	got, err := NewInvoiceService(store).Create(context.Background(), CreateInput{Items: originalItems})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, wantInvoice) {
		t.Fatalf("Create() = %#v, want %#v", got, wantInvoice)
	}
	if originalItems[0].ProductID != 2 {
		t.Fatal("Create() modified the caller's item slice")
	}
}

func TestInvoiceServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
	}{
		{name: "missing items", input: CreateInput{}},
		{name: "invalid product ID", input: CreateInput{Items: []InvoiceItem{{ProductID: 0, Quantity: 1}}}},
		{name: "invalid quantity", input: CreateInput{Items: []InvoiceItem{{ProductID: 1, Quantity: 0}}}},
		{
			name: "duplicated product",
			input: CreateInput{Items: []InvoiceItem{
				{ProductID: 1, Quantity: 1},
				{ProductID: 1, Quantity: 2},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeCalled := false
			store := invoiceStoreStub{
				create: func(context.Context, CreateInput) (*Invoice, error) {
					storeCalled = true
					return nil, nil
				},
			}

			_, err := NewInvoiceService(store).Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
			}
			if storeCalled {
				t.Fatal("Create() called store with invalid input")
			}
		})
	}
}

func TestInvoiceServiceWrapsStoreErrors(t *testing.T) {
	storeErr := errors.New("database unavailable")
	store := invoiceStoreStub{
		create: func(context.Context, CreateInput) (*Invoice, error) { return nil, storeErr },
		list:   func(context.Context) ([]Invoice, error) { return nil, storeErr },
		getByNumber: func(context.Context, int64) (*Invoice, error) {
			return nil, storeErr
		},
	}
	service := NewInvoiceService(store)

	_, createErr := service.Create(context.Background(), CreateInput{
		Items: []InvoiceItem{{ProductID: 1, Quantity: 1}},
	})
	if !errors.Is(createErr, storeErr) {
		t.Fatalf("Create() error = %v, want wrapped store error", createErr)
	}
	_, listErr := service.List(context.Background())
	if !errors.Is(listErr, storeErr) {
		t.Fatalf("List() error = %v, want wrapped store error", listErr)
	}
	_, getErr := service.GetByNumber(context.Background(), 1)
	if !errors.Is(getErr, storeErr) {
		t.Fatalf("GetByNumber() error = %v, want wrapped store error", getErr)
	}
}

func TestInvoiceServiceGetByNumberRejectsInvalidNumber(t *testing.T) {
	storeCalled := false
	store := invoiceStoreStub{
		getByNumber: func(context.Context, int64) (*Invoice, error) {
			storeCalled = true
			return nil, nil
		},
	}

	_, err := NewInvoiceService(store).GetByNumber(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("GetByNumber() error = %v, want ErrInvalidInput", err)
	}
	if storeCalled {
		t.Fatal("GetByNumber() called store with invalid number")
	}
}
