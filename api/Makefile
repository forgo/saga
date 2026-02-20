.PHONY: all build run test lint clean dev db-start db-stop migrate db-seed db-reset-seed generate-ios generate-web generate-server admin-token dev-full

# Variables
BINARY_NAME=saga-api
GO=go
GOFLAGS=-ldflags="-s -w"
GOLANGCI_LINT_VERSION=v2.10.1

# Default target
all: build

# Build the binary
build:
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) ./cmd/server

# Run the server
run: build
	./bin/$(BINARY_NAME)

# Run in development mode with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# Run tests
test:
	$(GO) test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint the code (requires golangci-lint)
lint:
	@command -v golangci-lint > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Start SurrealDB (development)
db-start:
	surreal start --user root --pass root --bind 0.0.0.0:8000 surrealkv:./data/surreal

# Start SurrealDB in memory (for testing)
db-start-memory:
	surreal start --user root --pass root --bind 0.0.0.0:8000 memory

# Run database migrations (using HTTP API for CLI version compatibility)
migrate:
	@echo "Running migrations..."
	@for f in migrations/*.surql; do \
		case "$$f" in \
			*seed.surql) continue ;; \
		esac; \
		echo "Applying $$f..."; \
		curl -sf -X POST http://localhost:8000/sql \
			-H "Accept: application/json" \
			-H "surreal-ns: saga" \
			-H "surreal-db: saga" \
			-u "root:root" \
			--data-binary @$$f > /dev/null; \
	done
	@echo "Migrations complete"

# Seed database with sample data (development only)
db-seed:
	@echo "Seeding database with sample data..."
	@curl -sf -X POST http://localhost:8000/sql \
		-H "Accept: application/json" \
		-H "surreal-ns: saga" \
		-H "surreal-db: saga" \
		-u "root:root" \
		--data-binary @migrations/seed.surql > /dev/null
	@echo "Database seeded successfully"
	@echo ""
	@echo "Demo credentials:"
	@echo "  Email:    demo@forgo.software"
	@echo "  Password: password123"

# Reset and reseed database
db-reset-seed:
	@$(MAKE) compose-down
	@docker volume rm saga_surrealdb_data 2>/dev/null || true
	@$(MAKE) compose-up
	@sleep 5
	@$(MAKE) migrate
	@$(MAKE) db-seed

# Generate iOS client types from OpenAPI spec
generate-ios:
	@echo "Generating iOS client..."
	swift-openapi-generator generate openapi/openapi.yaml \
		--output ../ios/Saga/Generated

# Generate TypeScript types from OpenAPI spec
generate-web:
	@echo "Generating TypeScript types..."
	npx openapi-typescript openapi/openapi.yaml \
		--output ../web/src/api/types.ts

# Generate Go server types from OpenAPI spec
generate-server:
	@echo "Generating Go server types..."
	@which oapi-codegen > /dev/null || (echo "Installing oapi-codegen..." && go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest)
	oapi-codegen -generate types,server \
		-package api openapi/openapi.yaml > internal/api/openapi.gen.go

# Generate all clients
generate: generate-server

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Docker build
docker-build:
	docker build -t saga-api:latest .

# Docker run (single container)
docker-run:
	docker run --rm -p 8080:8080 saga-api:latest

# Docker Compose commands
compose-up:
	docker compose up -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

compose-restart: compose-down compose-up

# Generate RSA keys for JWT (PKCS#1 format for Go compatibility)
keys-generate:
	@mkdir -p keys
	openssl genrsa -traditional -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys generated in ./keys/"

# Validate OpenAPI spec
openapi-validate:
	@npx @stoplight/spectral-cli lint openapi/openapi.yaml

# Preview OpenAPI docs
openapi-preview:
	@npx @redocly/cli preview-docs openapi/openapi.yaml

# Run benchmarks
bench:
	$(GO) test -bench=. -benchmem ./...

# Format code
fmt:
	gofmt -s -w .

# Vet code
vet:
	$(GO) vet ./...

# Run all checks
check: fmt vet lint test

# Install development tools
tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	go install github.com/air-verse/air@latest
	@echo "Tools installed"

# Generate an admin JWT token for testing admin endpoints
admin-token:
	@if [ ! -f keys/private.pem ]; then \
		echo "Keys not found. Generating..."; \
		$(MAKE) keys-generate; \
	fi
	@$(GO) run ./cmd/admin-token

# Generate admin token and output just the token (for scripting)
admin-token-raw:
	@if [ ! -f keys/private.pem ]; then $(MAKE) keys-generate; fi
	@$(GO) run ./cmd/admin-token -json | jq -r .access_token

# Full development setup: start DB, run migrations, seed, start API with hot reload
dev-full:
	@echo "Starting full development environment..."
	@if ! docker ps | grep -q saga-surrealdb; then \
		echo "Starting SurrealDB via Docker..."; \
		docker run -d --name saga-surrealdb --rm \
			-p 8000:8000 \
			surrealdb/surrealdb:v3.0.0-beta.3 \
			start --user root --pass root --bind 0.0.0.0:8000 memory; \
		sleep 3; \
	fi
	@$(MAKE) migrate
	@$(MAKE) db-seed
	@echo ""
	@echo "Starting API server with hot reload..."
	@echo "Admin token:"
	@$(MAKE) admin-token-raw 2>/dev/null || $(MAKE) admin-token
	@echo ""
	@$(MAKE) dev

# Stop dev environment
dev-stop:
	@echo "Stopping development environment..."
	@-docker stop saga-surrealdb 2>/dev/null || true
	@-pkill -f "air" 2>/dev/null || true
	@echo "Development environment stopped"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  build           - Build the binary"
	@echo "  run             - Build and run the server"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  dev-full        - Full dev setup: DB + migrations + seed + API"
	@echo "  dev-stop        - Stop dev environment (DB + API)"
	@echo "  admin-token     - Generate admin JWT token for testing"
	@echo "  admin-token-raw - Output just the admin token (for scripts)"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  bench           - Run benchmarks"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo "  lint            - Lint the code"
	@echo "  check           - Run all checks (fmt, vet, lint, test)"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  tools           - Install development tools"
	@echo ""
	@echo "Database:"
	@echo "  db-start        - Start SurrealDB (file-based)"
	@echo "  db-start-memory - Start SurrealDB (in-memory)"
	@echo "  migrate         - Run database migrations"
	@echo "  db-seed         - Seed database with sample data"
	@echo "  db-reset-seed   - Reset database and reseed"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  compose-up      - Start all services with docker-compose"
	@echo "  compose-down    - Stop all services"
	@echo "  compose-logs    - View docker-compose logs"
	@echo "  compose-restart - Restart all services"
	@echo ""
	@echo "OpenAPI:"
	@echo "  openapi-validate - Validate OpenAPI spec"
	@echo "  openapi-preview  - Preview OpenAPI docs"
	@echo "  generate         - Generate code from OpenAPI spec"
	@echo ""
	@echo "Keys:"
	@echo "  keys-generate   - Generate RSA keys for JWT"
