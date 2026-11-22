# KicauGo Backend Makefile
# Usage: make <target>

.PHONY: help up down build logs test test-race migrate shell clean lint migration index seed scaffold

# Default target
help:
	@echo "Available targets:"
	@echo "  up           - Start all services (detached)"
	@echo "  down         - Stop and remove all services"
	@echo "  build        - Build Docker images"
	@echo "  logs         - Follow logs from all services"
	@echo "  test         - Run Go tests in container"
	@echo "  test-race    - Run Go tests with race detector"
	@echo "  migrate      - Run database migrations manually"
	@echo "  shell        - Open shell in api container"
	@echo "  clean        - Remove containers, volumes, images"
	@echo "  lint         - Run golangci-lint (if installed)"
	@echo "  migration    - Create new migration file (NAME=<name>)"
	@echo "  index        - Create new index file (NAME=<name>)"
	@echo "  seed         - Create new seeder file (NAME=<name>)"
	@echo "  scaffold     - Generate CRUD feature (NAME=<domain> FIELDS=\"Name:string,Price:int\")"
	@echo ""
	@echo "API Documentation (Huma built-in):"
	@echo "  http://localhost:8080/docs     - Scalar UI"
	@echo "  http://localhost:8080/openapi.json - OpenAPI spec"

# Docker Compose
up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

# Testing
test:
	docker compose run --rm api go test ./...

test-race:
	docker compose run --rm api go test -race ./...

# Database
migrate:
	docker compose run --rm api /kicaugo migrate

# Debug
shell:
	docker compose run --rm api sh

# Cleanup
clean:
	docker compose down -v --rmi all --remove-orphans

# Scaffold Generator
# Usage: make scaffold NAME=<domain> FIELDS="Name:string,Price:int"
scaffold:
	@if [ -z "$(NAME)" ]; then echo "Usage: make scaffold NAME=<domain> FIELDS=\"Name:string,Price:int\""; exit 1; fi
	go run ./cmd/scaffold -name=$(NAME) -fields="$(FIELDS)"

# Migration, Index & Seeder Generator
# Usage: make migration NAME=create_users_table
#        make index NAME=create_users_index
#        make seed NAME=insert_demo_data
migration:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migration NAME=<name>"; exit 1; fi
	@LAST=$$(ls internal/database/migrations/*.sql 2>/dev/null | sort | tail -1 | sed 's/.*\///' | cut -d'_' -f1); \
	NEXT=$$(printf "%04d" $$((10#$${LAST:-0} + 1))); \
	FILE="internal/database/migrations/$${NEXT}_$(NAME).sql"; \
	echo "-- Migration: $(NAME)" > "$$FILE"; \
	echo "CREATE TABLE IF NOT EXISTS example (" >> "$$FILE"; \
	echo "    id BIGSERIAL PRIMARY KEY," >> "$$FILE"; \
	echo "    created_at TIMESTAMPTZ NOT NULL DEFAULT now()" >> "$$FILE"; \
	echo ");" >> "$$FILE"; \
	echo "Created: $$FILE"

index:
	@if [ -z "$(NAME)" ]; then echo "Usage: make index NAME=<name>"; exit 1; fi
	@LAST=$$(ls internal/database/indexes/*.sql 2>/dev/null | sort | tail -1 | sed 's/.*\///' | cut -d'_' -f1); \
	NEXT=$$(printf "%04d" $$((10#$${LAST:-0} + 1))); \
	FILE="internal/database/indexes/$${NEXT}_$(NAME).sql"; \
	echo "-- Index: $(NAME)" > "$$FILE"; \
	echo "CREATE INDEX IF NOT EXISTS idx_example ON example(id);" >> "$$FILE"; \
	if [ ! -f internal/database/indexes/embed.go ]; then \
		printf 'package indexes\n\nimport "embed"\n\n//go:embed *.sql\nvar IndexFiles embed.FS\n' > internal/database/indexes/embed.go; \
	fi; \
	echo "Created: $$FILE"

seed:
	@if [ -z "$(NAME)" ]; then echo "Usage: make seed NAME=<name>"; exit 1; fi
	@LAST=$$(ls internal/database/seeders/*.sql 2>/dev/null | sort | tail -1 | sed 's/.*\///' | cut -d'_' -f1); \
	NEXT=$$(printf "%04d" $$((10#$${LAST:-0} + 1))); \
	FILE="internal/database/seeders/$${NEXT}_$(NAME).sql"; \
	echo "-- Seed: $(NAME)" > "$$FILE"; \
	echo "INSERT INTO example (id) VALUES (1) ON CONFLICT DO NOTHING;" >> "$$FILE"; \
	echo "Created: $$FILE"
