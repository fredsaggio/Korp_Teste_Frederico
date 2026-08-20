-- +goose Up
-- +goose StatementBegin

CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    balance BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_products_code_not_blank CHECK (BTRIM(code) <> ''),
    CONSTRAINT chk_products_description_not_blank CHECK (BTRIM(description) <> ''),
    CONSTRAINT chk_products_balance_non_negative CHECK (balance >= 0)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE products;

-- +goose StatementEnd
