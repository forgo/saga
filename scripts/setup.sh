#!/usr/bin/env bash
set -euo pipefail

echo "========================================"
echo "Saga Development Environment Setup"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_command() {
    if command -v "$1" &> /dev/null; then
        echo -e "${GREEN}[ok]${NC} $1 found"
        return 0
    else
        echo -e "${RED}[missing]${NC} $1 not found"
        return 1
    fi
}

# Check prerequisites
echo "Checking prerequisites..."
echo ""

MISSING=0

check_command "go" || MISSING=1
check_command "docker" || MISSING=1
check_command "make" || MISSING=1

# Check Go version
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "  Go version: $GO_VERSION"
fi

# Check Docker is running
if command -v docker &> /dev/null; then
    if ! docker info &> /dev/null; then
        echo -e "${RED}[error]${NC} Docker is installed but not running"
        MISSING=1
    fi
fi

# Optional: iOS development
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo ""
    echo "iOS development (macOS only):"
    check_command "swift" || echo "  Install Xcode from App Store"
fi

if [[ $MISSING -eq 1 ]]; then
    echo ""
    echo -e "${RED}Please install missing dependencies and re-run setup${NC}"
    exit 1
fi

echo ""
echo "Setting up project..."
echo ""

# Navigate to repo root (script is in /scripts)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Copy environment example
if [[ ! -f "api/.env" ]]; then
    if [[ -f "api/.env.example" ]]; then
        cp api/.env.example api/.env
        echo -e "${GREEN}[ok]${NC} Created api/.env from example"
    fi
else
    echo -e "${YELLOW}[skip]${NC} api/.env already exists"
fi

# Generate JWT keys if missing
if [[ ! -f "api/keys/private.pem" ]]; then
    echo "Generating JWT keys..."
    make -C api keys-generate
    echo -e "${GREEN}[ok]${NC} JWT keys generated"
else
    echo -e "${YELLOW}[skip]${NC} JWT keys already exist"
fi

# Download Go dependencies
echo "Downloading Go dependencies..."
(cd api && go mod download)
echo -e "${GREEN}[ok]${NC} Go dependencies downloaded"

# Install Go tools
echo "Installing Go development tools..."
make -C api tools 2>/dev/null || true
echo -e "${GREEN}[ok]${NC} Go tools installed"

# Install git hooks
echo "Installing git hooks..."
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
echo -e "${GREEN}[ok]${NC} Git hooks installed"

# Start services
echo ""
echo "Starting services..."
docker compose -f api/docker-compose.yml up -d
echo -e "${GREEN}[ok]${NC} Services started"

# Wait for SurrealDB to be healthy
echo "Waiting for SurrealDB..."
for i in {1..30}; do
    if curl -sf http://localhost:8000/health > /dev/null 2>&1; then
        echo -e "${GREEN}[ok]${NC} SurrealDB is ready"
        break
    fi
    if [[ $i -eq 30 ]]; then
        echo -e "${YELLOW}[warn]${NC} SurrealDB health check timed out, continuing anyway..."
    fi
    sleep 1
done

# Run migrations
echo "Running migrations..."
make -C api migrate
echo -e "${GREEN}[ok]${NC} Migrations applied"

# Seed database
echo "Seeding database..."
make -C api db-seed
echo -e "${GREEN}[ok]${NC} Sample data created"

echo ""
echo "========================================"
echo -e "${GREEN}Setup complete!${NC}"
echo "========================================"
echo ""
echo "Quick start:"
echo "  make dev      - Start all services"
echo "  make dev-api  - Run API with hot reload"
echo "  make dev-ios  - Open iOS project in Xcode"
echo "  make test     - Run all tests"
echo "  make help     - Show all commands"
echo ""
echo "API:       http://localhost:8080"
echo "SurrealDB: http://localhost:8000"
echo ""
echo "Demo credentials:"
echo "  Email:    demo@saga.app"
echo "  Password: password123"
