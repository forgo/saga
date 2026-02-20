.PHONY: help setup dev stop restart logs dev-api dev-ios dev-admin test test-api test-ios test-admin lint lint-api lint-admin fmt check db-seed db-reset clean keys install-hooks

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

restart: ## Stop, rebuild, and restart everything from scratch
	@echo "Stopping all services..."
	@$(MAKE) -C api compose-down 2>/dev/null || true
	@echo "Removing database volume..."
	@docker volume rm saga_surrealdb_data 2>/dev/null || true
	@echo "Rebuilding and starting services..."
	@cd api && docker compose up -d --build
	@echo "Waiting for SurrealDB to be ready..."
	@sleep 5
	@echo "Running migrations..."
	@$(MAKE) -C api migrate
	@echo "Seeding database..."
	@$(MAKE) -C api db-seed
	@echo ""
	@echo "Restart complete! Services running:"
	@echo "  API:       http://localhost:8080"
	@echo "  SurrealDB: http://localhost:8000"

logs: ## View service logs
	@$(MAKE) -C api compose-logs

dev-api: ## Run API with hot reload (requires running db)
	@$(MAKE) -C api dev

dev-ios: ## Open iOS project in Xcode
	@open ios/Saga/Saga.xcodeproj

dev-admin: ## Start admin dev server
	@$(MAKE) -C admin dev

# ============================================================================
# Testing
# ============================================================================

test: test-api test-ios test-admin ## Run all tests

test-api: ## Run API tests
	@$(MAKE) -C api test

test-ios: ## Run iOS tests
	@cd ios/Saga && swift test

test-admin: ## Run admin tests
	@$(MAKE) -C admin test

# ============================================================================
# Linting & Formatting
# ============================================================================

lint: lint-api lint-admin ## Lint all code

lint-api: ## Lint API code
	@$(MAKE) -C api lint

lint-admin: ## Lint admin code
	@$(MAKE) -C admin lint

fmt: ## Format all code
	@$(MAKE) -C api fmt
	@$(MAKE) -C admin fmt

check: ## Run all checks (lint, fmt, test)
	@$(MAKE) -C api check
	@$(MAKE) -C admin check

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
# Git Hooks
# ============================================================================

install-hooks: ## Install git pre-commit hooks
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed"

# ============================================================================
# Misc
# ============================================================================

clean: ## Clean all build artifacts
	@$(MAKE) -C api clean
	@$(MAKE) -C admin clean
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
