#!/bin/sh

set -eu

exec goose -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" up
