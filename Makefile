# Database connection (override on the command line if needed).
PG_HOST_PORT ?= 55432
DB_URL ?= postgres://postgres:postgres@localhost:$(PG_HOST_PORT)/swipeassist?sslmode=disable
ROOT_DB_URL ?= $(DB_URL)
MIGRATIONS_DIR ?= db/migrations
COMPOSE ?= docker compose -f docker-compose.db.yml

.PHONY: db/up db/down db/migrate db/gen migrate.up migrate.down1 migrate.version

# Start the local Postgres container.
db/up:
	PG_HOST_PORT=$(PG_HOST_PORT) $(COMPOSE) up -d

# Stop and remove the Postgres container.
db/down:
	PG_HOST_PORT=$(PG_HOST_PORT) $(COMPOSE) down

# Apply the SQL migrations to the target database.
db/migrate:
	psql $(DB_URL) -f db/migrations/0001_init.sql

db/login:
	psql "$(ROOT_DB_URL)"

# Apply migrations using golang-migrate.
migrate.up:
	migrate -path $(MIGRATIONS_DIR) -database "$(ROOT_DB_URL)" up

# Roll back exactly one migration.
migrate.down1:
	migrate -path $(MIGRATIONS_DIR) -database "$(ROOT_DB_URL)" down 1

# Show current migration version.
migrate.version:
	migrate -path $(MIGRATIONS_DIR) -database "$(ROOT_DB_URL)" version

# Generate db access layer via sqlc (outputs to internal/decisiondb).
db/gen:
	sqlc generate
