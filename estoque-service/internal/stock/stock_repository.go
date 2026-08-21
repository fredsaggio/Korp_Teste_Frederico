package stock

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"korp/estoque-service/internal/db"

	"github.com/jackc/pgx/v5"
)

type debitStore struct {
	db db.DB
}

func NewDebitStore(database db.DB) DebitStore {
	return &debitStore{db: database}
}

func (s *debitStore) Debit(ctx context.Context, input DebitInput) (*DebitResult, error) {
	result := &DebitResult{Reference: input.Reference}

	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		debitID, created, err := createDebit(ctx, tx, input.Reference)
		if err != nil {
			return err
		}
		if !created {
			matches, err := debitMatches(ctx, tx, input)
			if err != nil {
				return err
			}
			if !matches {
				return ErrIdempotencyConflict
			}
			result.AlreadyProcessed = true
			return nil
		}

		for _, item := range input.Items {
			if err := applyDebitItem(ctx, tx, debitID, item); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("debit stock transaction: %w", err)
	}

	return result, nil
}

func createDebit(ctx context.Context, tx pgx.Tx, reference string) (int64, bool, error) {
	const query = `
		INSERT INTO stock_debits (reference)
		VALUES ($1)
		ON CONFLICT (reference) DO NOTHING
		RETURNING id
	`

	var debitID int64
	if err := tx.QueryRow(ctx, query, reference).Scan(&debitID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("insert stock debit: %w", err)
	}

	return debitID, true, nil
}

func debitMatches(ctx context.Context, tx pgx.Tx, input DebitInput) (bool, error) {
	const query = `
		SELECT item.product_id, item.quantity
		FROM stock_debit_items AS item
		JOIN stock_debits AS debit ON debit.id = item.stock_debit_id
		WHERE debit.reference = $1
		ORDER BY item.product_id
	`

	rows, err := tx.Query(ctx, query, input.Reference)
	if err != nil {
		return false, fmt.Errorf("query existing stock debit: %w", err)
	}
	defer rows.Close()

	existingItems := make([]DebitItem, 0, len(input.Items))
	for rows.Next() {
		var item DebitItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return false, fmt.Errorf("scan existing stock debit item: %w", err)
		}
		existingItems = append(existingItems, item)
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate existing stock debit items: %w", err)
	}

	return slices.Equal(existingItems, input.Items), nil
}

func applyDebitItem(ctx context.Context, tx pgx.Tx, debitID int64, item DebitItem) error {
	const updateProduct = `
		UPDATE products
		SET balance = balance - @quantity,
			updated_at = NOW()
		WHERE id = @product_id
			AND balance >= @quantity
		RETURNING balance
	`
	args := pgx.StrictNamedArgs{
		"product_id": item.ProductID,
		"quantity":   item.Quantity,
	}

	var remainingBalance int64
	if err := tx.QueryRow(ctx, updateProduct, args).Scan(&remainingBalance); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("update product balance: %w", err)
		}

		exists, err := productExists(ctx, tx, item.ProductID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: product %d", ErrProductNotFound, item.ProductID)
		}
		return fmt.Errorf("%w: product %d", ErrInsufficientStock, item.ProductID)
	}

	const insertItem = `
		INSERT INTO stock_debit_items (stock_debit_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, insertItem, debitID, item.ProductID, item.Quantity); err != nil {
		return fmt.Errorf("insert stock debit item: %w", err)
	}

	return nil
}

func productExists(ctx context.Context, tx pgx.Tx, productID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	if err := tx.QueryRow(ctx, query, productID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check product existence: %w", err)
	}
	return exists, nil
}
