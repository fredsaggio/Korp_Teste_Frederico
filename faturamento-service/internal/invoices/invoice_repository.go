package invoices

import (
	"context"
	"fmt"

	"korp/faturamento-service/internal/db"

	"github.com/jackc/pgx/v5"
)

type invoiceStore struct {
	db db.DB
}

func NewInvoiceStore(database db.DB) InvoiceStore {
	return &invoiceStore{db: database}
}

func (s *invoiceStore) Create(ctx context.Context, input CreateInput) (*Invoice, error) {
	var invoice Invoice

	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		const insertInvoice = `
			INSERT INTO invoices DEFAULT VALUES
			RETURNING id, status, created_at, updated_at, closed_at
		`
		if err := tx.QueryRow(ctx, insertInvoice).Scan(
			&invoice.Number,
			&invoice.Status,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.ClosedAt,
		); err != nil {
			return fmt.Errorf("insert invoice: %w", err)
		}

		const insertItem = `
			INSERT INTO invoice_items (invoice_id, product_id, quantity)
			VALUES ($1, $2, $3)
		`
		for _, item := range input.Items {
			if _, err := tx.Exec(ctx, insertItem, invoice.Number, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("insert invoice item: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice transaction: %w", err)
	}

	invoice.Items = append([]InvoiceItem(nil), input.Items...)
	return &invoice, nil
}

func (s *invoiceStore) List(ctx context.Context) ([]Invoice, error) {
	const query = `
		SELECT
			invoice.id,
			invoice.status,
			invoice.created_at,
			invoice.updated_at,
			invoice.closed_at,
			item.product_id,
			item.quantity
		FROM invoices AS invoice
		JOIN invoice_items AS item ON item.invoice_id = invoice.id
		ORDER BY invoice.id, item.product_id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}
	defer rows.Close()

	invoices, err := scanInvoices(rows)
	if err != nil {
		return nil, fmt.Errorf("scan invoices: %w", err)
	}
	return invoices, nil
}

func (s *invoiceStore) GetByNumber(ctx context.Context, number int64) (*Invoice, error) {
	const query = `
		SELECT
			invoice.id,
			invoice.status,
			invoice.created_at,
			invoice.updated_at,
			invoice.closed_at,
			item.product_id,
			item.quantity
		FROM invoices AS invoice
		JOIN invoice_items AS item ON item.invoice_id = invoice.id
		WHERE invoice.id = $1
		ORDER BY item.product_id
	`

	rows, err := s.db.Query(ctx, query, number)
	if err != nil {
		return nil, fmt.Errorf("query invoice: %w", err)
	}
	defer rows.Close()

	invoices, err := scanInvoices(rows)
	if err != nil {
		return nil, fmt.Errorf("scan invoice: %w", err)
	}
	if len(invoices) == 0 {
		return nil, ErrInvoiceNotFound
	}
	return &invoices[0], nil
}

func scanInvoices(rows pgx.Rows) ([]Invoice, error) {
	invoices := make([]Invoice, 0)
	for rows.Next() {
		var invoice Invoice
		var item InvoiceItem
		if err := rows.Scan(
			&invoice.Number,
			&invoice.Status,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.ClosedAt,
			&item.ProductID,
			&item.Quantity,
		); err != nil {
			return nil, err
		}

		last := len(invoices) - 1
		if last < 0 || invoices[last].Number != invoice.Number {
			invoice.Items = make([]InvoiceItem, 0, 1)
			invoices = append(invoices, invoice)
			last++
		}
		invoices[last].Items = append(invoices[last].Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}
