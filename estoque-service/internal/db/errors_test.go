package db

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	err := fmt.Errorf("insert product: %w", &pgconn.PgError{
		Code:           uniqueViolationCode,
		ConstraintName: "products_code_key",
	})

	if !IsUniqueViolation(err, "products_code_key") {
		t.Fatal("IsUniqueViolation() = false, want true")
	}
	if IsUniqueViolation(err, "another_constraint") {
		t.Fatal("IsUniqueViolation() = true for another constraint")
	}
}
