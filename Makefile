.PHONY: help setup dev stop logs test test-api test-ios lint fmt check db-seed db-reset clean keys

# Default target
.DEFAULT_GOAL := help

# ============================================================================
# Setup
# ============================================================================

setup: ## First-time setup for new developers
	@./scripts/setup.sh

# ============================================================================
# Development
# ============================================================================

dev: ## Start all services (API + SurrealDB)
	@$(MAKE) -C api compose-up
	@echo ""
	@echo "Services started:"
	@echo "  API:       http://localhost:8080"
	@echo "  SurrealDB: http://localhost:8000"
	@echo ""
	@echo "Run 'make logs' to view logs"

stop: ## Stop all services
	@$(MAKE) -C api compose-down

logs: ## View service logs
	@$(MAKE) -C api compose-logs

dev-api: ## Run API with hot reload (requires running db)
	@$(MAKE) -C api dev

dev-ios: ## Open iOS project in Xcode
	@open ios/Saga/Saga.xcodeproj

# ============================================================================
# Testing
# ============================================================================

test: test-api test-ios ## Run all tests

test-api: ## Run API tests
	@$(MAKE) -C api test

test-ios: ## Run iOS tests
	@cd ios/Saga && swift test

# ============================================================================
# Linting & Formatting
# ============================================================================

lint: ## Lint all code
	@$(MAKE) -C api lint

fmt: ## Format all code
	@$(MAKE) -C api fmt

check: ## Run all checks (lint, fmt, test)
	@$(MAKE) -C api check

# ============================================================================
# Database
# ============================================================================

db-seed: ## Seed database with sample data
	@$(MAKE) -C api db-seed

db-reset: ## Reset database (drop and recreate with seed data)
	@echo "Stopping services..."
	@$(MAKE) -C api compose-down 2>/dev/null || true
	@echo "Removing database volume..."
	@docker volume rm saga_surrealdb_data 2>/dev/null || true
	@echo "Starting services..."
	@$(MAKE) -C api compose-up
	@echo "Waiting for SurrealDB..."
	@sleep 5
	@echo "Running migrations..."
	@$(MAKE) -C api migrate
	@echo "Seeding database..."
	@$(MAKE) -C api db-seed
	@echo ""
	@echo "Database reset complete!"

# ============================================================================
# Misc
# ============================================================================

clean: ## Clean all build artifacts
	@$(MAKE) -C api clean
	@rm -rf ios/Saga/.build

keys: ## Generate JWT keys
	@$(MAKE) -C api keys-generate

# ============================================================================
# Help
# ============================================================================

help: ## Show this help
	@echo "Saga Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "First time? Run: make setup"
