package products

import (
	"context"
	"errors"
	"os"
	"testing"

	"korp/estoque-service/internal/db"
)

func TestProductStoreIntegration(t *testing.T) {
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

	if _, err := tx.Exec(ctx, "DELETE FROM products"); err != nil {
		t.Fatal(err)
	}

	store := NewProductStore(tx)
	first, err := store.Create(ctx, ProductInput{
		Code:        "PROD-001",
		Description: "Product one",
		Balance:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || first.Code != "PROD-001" || first.Balance != 10 {
		t.Fatalf("Create() = %#v", first)
	}
	if first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatal("Create() returned zero timestamps")
	}

	second, err := store.Create(ctx, ProductInput{
		Code:        "PROD-002",
		Description: "Product two",
		Balance:     20,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d products, want 2", len(got))
	}
	if got[0].ID != first.ID || got[1].ID != second.ID {
		t.Fatalf("List() order = [%d, %d], want [%d, %d]", got[0].ID, got[1].ID, first.ID, second.ID)
	}

	_, err = store.Create(ctx, ProductInput{
		Code:        "PROD-001",
		Description: "Duplicate",
		Balance:     1,
	})
	if !errors.Is(err, ErrCodeAlreadyExists) {
		t.Fatalf("Create() duplicate error = %v, want ErrCodeAlreadyExists", err)
	}
}
