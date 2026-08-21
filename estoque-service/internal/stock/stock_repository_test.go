package stock

import (
	"context"
	"errors"
	"os"
	"testing"

	"korp/estoque-service/internal/db"
	"korp/estoque-service/internal/products"
)

func TestDebitStoreIntegration(t *testing.T) {
	connectionString := os.Getenv("ESTOQUE_TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("ESTOQUE_TEST_DATABASE_URL is not set")
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

	if _, err := tx.Exec(ctx, "DELETE FROM stock_debit_items"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM stock_debits"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM products"); err != nil {
		t.Fatal(err)
	}

	productStore := products.NewProductStore(tx)
	first := createTestProduct(t, ctx, productStore, "STOCK-001", 10)
	second := createTestProduct(t, ctx, productStore, "STOCK-002", 5)

	store := NewDebitStore(tx)
	input := DebitInput{
		Reference: "invoice:123",
		Items: []DebitItem{
			{ProductID: first.ID, Quantity: 3},
			{ProductID: second.ID, Quantity: 2},
		},
	}

	result, err := store.Debit(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Reference != input.Reference || result.AlreadyProcessed {
		t.Fatalf("Debit() = %#v, want first processing", result)
	}
	assertProductBalance(t, ctx, tx, first.ID, 7)
	assertProductBalance(t, ctx, tx, second.ID, 3)
	assertDebitItemCount(t, ctx, tx, input.Reference, 2)

	result, err = store.Debit(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.AlreadyProcessed {
		t.Fatalf("Debit() retry = %#v, want already processed", result)
	}
	assertProductBalance(t, ctx, tx, first.ID, 7)
	assertProductBalance(t, ctx, tx, second.ID, 3)

	_, err = store.Debit(ctx, DebitInput{
		Reference: input.Reference,
		Items:     []DebitItem{{ProductID: first.ID, Quantity: 1}},
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("Debit() conflicting retry error = %v, want ErrIdempotencyConflict", err)
	}

	_, err = store.Debit(ctx, DebitInput{
		Reference: "invoice:insufficient",
		Items: []DebitItem{
			{ProductID: first.ID, Quantity: 1},
			{ProductID: second.ID, Quantity: 4},
		},
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("Debit() insufficient error = %v, want ErrInsufficientStock", err)
	}
	assertProductBalance(t, ctx, tx, first.ID, 7)
	assertProductBalance(t, ctx, tx, second.ID, 3)
	assertDebitDoesNotExist(t, ctx, tx, "invoice:insufficient")

	_, err = store.Debit(ctx, DebitInput{
		Reference: "invoice:missing",
		Items:     []DebitItem{{ProductID: 999999, Quantity: 1}},
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("Debit() missing product error = %v, want ErrProductNotFound", err)
	}
	assertDebitDoesNotExist(t, ctx, tx, "invoice:missing")
}

func createTestProduct(t *testing.T, ctx context.Context, store products.ProductStore, code string, balance int64) *products.Product {
	t.Helper()

	product, err := store.Create(ctx, products.ProductInput{
		Code:        code,
		Description: code,
		Balance:     balance,
	})
	if err != nil {
		t.Fatal(err)
	}
	return product
}

func assertProductBalance(t *testing.T, ctx context.Context, database db.DB, productID, want int64) {
	t.Helper()

	var got int64
	if err := database.QueryRow(ctx, "SELECT balance FROM products WHERE id = $1", productID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("product %d balance = %d, want %d", productID, got, want)
	}
}

func assertDebitItemCount(t *testing.T, ctx context.Context, database db.DB, reference string, want int) {
	t.Helper()

	const query = `
		SELECT COUNT(*)
		FROM stock_debit_items AS item
		JOIN stock_debits AS debit ON debit.id = item.stock_debit_id
		WHERE debit.reference = $1
	`
	var got int
	if err := database.QueryRow(ctx, query, reference).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("debit %q item count = %d, want %d", reference, got, want)
	}
}

func assertDebitDoesNotExist(t *testing.T, ctx context.Context, database db.DB, reference string) {
	t.Helper()

	var exists bool
	if err := database.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM stock_debits WHERE reference = $1)", reference).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("debit %q exists after failed transaction", reference)
	}
}
