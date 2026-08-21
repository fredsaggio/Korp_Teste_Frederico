-- +goose Up
-- +goose StatementBegin

CREATE TABLE stock_debits (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    reference TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_stock_debits_reference_not_blank CHECK (BTRIM(reference) <> '')
);

CREATE TABLE stock_debit_items (
    stock_debit_id BIGINT NOT NULL REFERENCES stock_debits(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity BIGINT NOT NULL,
    PRIMARY KEY (stock_debit_id, product_id),
    CONSTRAINT chk_stock_debit_items_quantity_positive CHECK (quantity > 0)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE stock_debit_items;
DROP TABLE stock_debits;

-- +goose StatementEnd
