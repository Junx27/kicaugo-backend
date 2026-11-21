# KicauGo Backend Makefile
# Usage: make <target>

.PHONY: help up down build logs test test-race migrate shell clean lint

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

# Linting (requires golangci-lint)
lint:
	golangci-lint run ./...