package invoices

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"korp/faturamento-service/internal/stock"
)

type stockClientStub struct {
	debit func(context.Context, stock.DebitInput) error
}

func (s stockClientStub) Debit(ctx context.Context, input stock.DebitInput) error {
	return s.debit(ctx, input)
}

func TestInvoiceClosingServiceClose(t *testing.T) {
	invoice := &Invoice{
		Number: 10,
		Status: StatusOpen,
		Items: []InvoiceItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	}
	wantDebit := stock.DebitInput{
		Reference: "invoice:10",
		Items: []stock.Item{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	}
	closedAt := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)

	store := invoiceStoreStub{
		getByNumber: func(_ context.Context, number int64) (*Invoice, error) {
			if number != invoice.Number {
				t.Fatalf("GetByNumber number = %d, want %d", number, invoice.Number)
			}
			return invoice, nil
		},
		close: func(_ context.Context, number int64) (time.Time, error) {
			if number != invoice.Number {
				t.Fatalf("Close number = %d, want %d", number, invoice.Number)
			}
			return closedAt, nil
		},
	}
	client := stockClientStub{
		debit: func(_ context.Context, input stock.DebitInput) error {
			if !reflect.DeepEqual(input, wantDebit) {
				t.Fatalf("Debit input = %#v, want %#v", input, wantDebit)
			}
			return nil
		},
	}

	got, err := NewInvoiceClosingService(store, client).Close(context.Background(), invoice.Number)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusClosed || got.ClosedAt == nil || !got.ClosedAt.Equal(closedAt) {
		t.Fatalf("Close() = %#v, want closed invoice", got)
	}
	if invoice.Status != StatusOpen {
		t.Fatal("Close() modified the invoice returned by the store")
	}
}

func TestInvoiceClosingServiceDoesNotDebitClosedInvoice(t *testing.T) {
	clientCalled := false
	store := invoiceStoreStub{
		getByNumber: func(context.Context, int64) (*Invoice, error) {
			return &Invoice{Number: 10, Status: StatusClosed}, nil
		},
	}
	client := stockClientStub{
		debit: func(context.Context, stock.DebitInput) error {
			clientCalled = true
			return nil
		},
	}

	_, err := NewInvoiceClosingService(store, client).Close(context.Background(), 10)
	if !errors.Is(err, ErrInvoiceNotOpen) {
		t.Fatalf("Close() error = %v, want ErrInvoiceNotOpen", err)
	}
	if clientCalled {
		t.Fatal("Close() debited stock for a closed invoice")
	}
}

func TestInvoiceClosingServiceKeepsInvoiceOpenWhenDebitFails(t *testing.T) {
	debitErr := &stock.RejectionError{StatusCode: 409, Message: "Saldo insuficiente."}
	storeCloseCalled := false
	store := invoiceStoreStub{
		getByNumber: func(context.Context, int64) (*Invoice, error) {
			return &Invoice{
				Number: 10,
				Status: StatusOpen,
				Items:  []InvoiceItem{{ProductID: 1, Quantity: 2}},
			}, nil
		},
		close: func(context.Context, int64) (time.Time, error) {
			storeCloseCalled = true
			return time.Time{}, nil
		},
	}
	client := stockClientStub{
		debit: func(context.Context, stock.DebitInput) error { return debitErr },
	}

	_, err := NewInvoiceClosingService(store, client).Close(context.Background(), 10)
	if !errors.Is(err, debitErr) {
		t.Fatalf("Close() error = %v, want wrapped debit error", err)
	}
	if storeCloseCalled {
		t.Fatal("Close() persisted closed invoice after debit failure")
	}
}

func TestInvoiceClosingServiceValidatesNumber(t *testing.T) {
	_, err := NewInvoiceClosingService(invoiceStoreStub{}, stockClientStub{}).Close(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Close() error = %v, want ErrInvalidInput", err)
	}
}
