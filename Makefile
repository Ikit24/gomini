.PHONY: up down status
DB_PATH = ./gomini.db
MIGRATIONS_DIR = internal/database/sql
GOOSE_CMD = goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH)
up:
	$(GOOSE_CMD) up

down:
	$(GOOSE_CMD) down

status:
	$(GOOSE_CMD) status
