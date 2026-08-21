-- +goose Up
-- +goose StatementBegin

CREATE TABLE invoices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CONSTRAINT chk_invoices_status CHECK (status IN ('OPEN', 'CLOSED')),
    CONSTRAINT chk_invoices_closed_at CHECK (
        (status = 'OPEN' AND closed_at IS NULL)
        OR (status = 'CLOSED' AND closed_at IS NOT NULL)
    )
);

CREATE TABLE invoice_items (
    invoice_id BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    PRIMARY KEY (invoice_id, product_id),
    CONSTRAINT chk_invoice_items_product_id_positive CHECK (product_id > 0),
    CONSTRAINT chk_invoice_items_quantity_positive CHECK (quantity > 0)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE invoice_items;
DROP TABLE invoices;

-- +goose StatementEnd
