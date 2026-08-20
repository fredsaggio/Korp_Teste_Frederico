-include .env
export

ESTOQUE_MIGRATIONS_DIR := estoque-service/internal/db/migrations
ESTOQUE_DB_HOST ?= localhost
ESTOQUE_DB_PORT ?= 5433
ESTOQUE_LOCAL_DATABASE_URL := postgres://$(POSTGRES_ESTOQUE_USER):$(POSTGRES_ESTOQUE_PASSWORD)@$(ESTOQUE_DB_HOST):$(ESTOQUE_DB_PORT)/$(POSTGRES_ESTOQUE_DB)?sslmode=disable

.PHONY: migration/estoque/new
migration/estoque/new:
	@goose -dir $(ESTOQUE_MIGRATIONS_DIR) create "$(name)" sql

.PHONY: migration/estoque/up
migration/estoque/up:
	@goose -dir $(ESTOQUE_MIGRATIONS_DIR) postgres "$(ESTOQUE_LOCAL_DATABASE_URL)" up

.PHONY: migration/estoque/down
migration/estoque/down:
	@goose -dir $(ESTOQUE_MIGRATIONS_DIR) postgres "$(ESTOQUE_LOCAL_DATABASE_URL)" down

.PHONY: migration/estoque/status
migration/estoque/status:
	@goose -dir $(ESTOQUE_MIGRATIONS_DIR) postgres "$(ESTOQUE_LOCAL_DATABASE_URL)" status

.PHONY: test/estoque/integration
test/estoque/integration:
	@cd estoque-service && ESTOQUE_TEST_DATABASE_URL="$(ESTOQUE_LOCAL_DATABASE_URL)" go test -count=1 ./internal/products
