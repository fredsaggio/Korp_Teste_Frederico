package invoices

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"korp/faturamento-service/internal/db"
)

func TestInvoiceStoreIntegration(t *testing.T) {
	connectionString := os.Getenv("FATURAMENTO_TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("FATURAMENTO_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Errorf("rollback transaction: %v", err)
		}
	}()

	if _, err := tx.Exec(ctx, "DELETE FROM invoices"); err != nil {
		t.Fatal(err)
	}

	store := NewInvoiceStore(tx)
	firstInput := CreateInput{Items: []InvoiceItem{
		{ProductID: 1, Quantity: 3},
		{ProductID: 2, Quantity: 1},
	}}
	first, err := store.Create(ctx, firstInput)
	if err != nil {
		t.Fatal(err)
	}
	if first.Number == 0 || first.Status != StatusOpen {
		t.Fatalf("Create() = %#v, want numbered open invoice", first)
	}
	if !reflect.DeepEqual(first.Items, firstInput.Items) {
		t.Fatalf("Create() items = %#v, want %#v", first.Items, firstInput.Items)
	}
	if first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() || first.ClosedAt != nil {
		t.Fatalf("Create() timestamps = %#v", first)
	}

	second, err := store.Create(ctx, CreateInput{
		Items: []InvoiceItem{{ProductID: 3, Quantity: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.Number <= first.Number {
		t.Fatalf("second invoice number = %d, want greater than %d", second.Number, first.Number)
	}

	got, err := store.GetByNumber(ctx, first.Number)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Items, firstInput.Items) {
		t.Fatalf("GetByNumber() items = %#v, want %#v", got.Items, firstInput.Items)
	}

	listed, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].Number != first.Number || listed[1].Number != second.Number {
		t.Fatalf("List() = %#v, want invoices ordered by number", listed)
	}

	_, err = store.GetByNumber(ctx, second.Number+999999)
	if !errors.Is(err, ErrInvoiceNotFound) {
		t.Fatalf("GetByNumber() error = %v, want ErrInvoiceNotFound", err)
	}

	var countBefore int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM invoices").Scan(&countBefore); err != nil {
		t.Fatal(err)
	}
	_, err = store.Create(ctx, CreateInput{
		Items: []InvoiceItem{{ProductID: 4, Quantity: 0}},
	})
	if err == nil {
		t.Fatal("Create() with invalid item returned nil error")
	}
	var countAfter int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM invoices").Scan(&countAfter); err != nil {
		t.Fatal(err)
	}
	if countAfter != countBefore {
		t.Fatalf("invoice count after failed transaction = %d, want %d", countAfter, countBefore)
	}
}
