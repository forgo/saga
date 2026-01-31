# Contributing to Saga

Thank you for your interest in contributing to Saga! This document provides guidelines and information for contributors.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Architecture](#code-architecture)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Commit Messages](#commit-messages)

## Development Setup

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Make
- Xcode 16+ (for iOS development)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/forgo/saga.git
cd saga

# Run setup (creates .env, generates keys, starts services)
make setup

# Start development environment
make dev

# Run tests
make test
```

### Environment Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp api/.env.example api/.env
```

Key configuration:
- `DB_*` - SurrealDB connection settings
- `JWT_*` - JWT signing key paths
- `OAUTH_*` - OAuth provider credentials (optional)

## Code Architecture

Saga follows a **layered architecture** pattern:

```
api/internal/
├── handler/     # HTTP request handlers (parse, validate, respond)
├── service/     # Business logic (orchestration, rules)
├── repository/  # Data access (database queries)
├── model/       # Domain models and errors
├── middleware/  # HTTP middleware chain
├── config/      # Configuration loading
└── database/    # Database abstraction
```

### Layer Responsibilities

| Layer | Responsibility | Depends On |
|-------|---------------|------------|
| Handler | HTTP parsing, validation, response formatting | Service |
| Service | Business rules, orchestration | Repository |
| Repository | Database queries, result parsing | Database |
| Model | Data structures, validation methods | None |

### Key Patterns

**Dependency Injection**: Services receive dependencies via config structs:

```go
type InterestServiceConfig struct {
    InterestRepo InterestRepository
}

func NewInterestService(cfg InterestServiceConfig) *InterestService
```

**Interface Segregation**: Services define the repository interfaces they need:

```go
// In service/interest.go
type InterestRepository interface {
    Create(ctx context.Context, interest *model.Interest) error
    GetByID(ctx context.Context, id string) (*model.Interest, error)
}
```

**Error Handling**: Use domain-specific errors from `model/errors.go`:

```go
var ErrInterestNotFound = errors.New("interest not found")
```

## Making Changes

### Before Starting

1. Check existing issues and PRs
2. For new features, open an issue to discuss first
3. Review [DEVELOPMENT.md](api/docs/DEVELOPMENT.md) for current status

### Directory Guide

| Change Type | Where to Look |
|-------------|--------------|
| API endpoint | `handler/`, then `service/`, then `repository/` |
| Business logic | `service/` |
| Database queries | `repository/` |
| Data models | `model/` |
| Auth/middleware | `middleware/` |
| OpenAPI spec | `openapi/paths/`, `openapi/components/` |
| Database schema | `migrations/` |

### Adding a New Feature

1. **Model**: Define types in `model/`
2. **Repository**: Add data access in `repository/`
3. **Service**: Add business logic in `service/`
4. **Handler**: Add HTTP endpoints in `handler/`
5. **Routes**: Wire up in `cmd/server/main.go`
6. **Tests**: Add tests in `tests/`
7. **OpenAPI**: Update spec in `openapi/`

## Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test -v ./tests/... -run TestAuth

# With race detection
go test -race ./...
```

### Test Structure

Tests use BDD-style acceptance criteria:

```go
/*
FEATURE: User Authentication
DOMAIN: Auth

ACCEPTANCE CRITERIA:
===================

AC-AUTH-001: Register with Email/Password
  GIVEN valid email and password (8+ chars)
  WHEN user submits registration
  THEN user is created with hashed password
*/

func TestAuth_RegisterWithEmailPassword(t *testing.T) {
    // AC-AUTH-001
    ...
}
```

### Test Infrastructure

- `internal/testing/testdb/` - Test database setup
- `internal/testing/fixtures/` - Test data factories
- `internal/testing/helpers/` - Common test utilities

## Pull Request Process

### Before Submitting

1. **Tests pass**: `make test`
2. **Linting passes**: `make lint`
3. **Coverage maintained**: Check `make test-coverage`
4. **Documentation updated**: Update relevant docs

### PR Checklist

- [ ] Tests added/updated for changes
- [ ] Documentation updated (if applicable)
- [ ] OpenAPI spec updated (if API changes)
- [ ] CHANGELOG.md updated (for notable changes)
- [ ] No unrelated changes included

### Review Process

1. Create PR against `main` branch
2. CI runs automatically (tests, lint, build)
3. Address review feedback
4. Squash and merge when approved

## Code Style

### Go

Follow standard Go conventions:

```go
// Good: Clear, idiomatic Go
func (s *InterestService) GetByID(ctx context.Context, id string) (*model.Interest, error) {
    interest, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get interest: %w", err)
    }
    return interest, nil
}
```

Key guidelines:
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Handle all errors explicitly
- Use context for cancellation
- Prefer explicit over clever

### Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Package | lowercase, short | `handler`, `service` |
| Interface | -er suffix or descriptive | `InterestRepository` |
| Constructor | `New` prefix | `NewInterestService` |
| Error variables | `Err` prefix | `ErrNotFound` |

### Comments

```go
// InterestService handles interest-related business logic.
// It manages user interests and compatibility matching.
type InterestService struct {
    repo InterestRepository
}

// GetByID retrieves an interest by its unique identifier.
// Returns ErrInterestNotFound if the interest doesn't exist.
func (s *InterestService) GetByID(ctx context.Context, id string) (*model.Interest, error)
```

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code change, no feature/fix |
| `test` | Adding tests |
| `chore` | Maintenance tasks |
| `ci` | CI/CD changes |

### Examples

```
feat(auth): add passkey authentication support

fix(timer): correct threshold calculation for daylight saving

docs(api): update OpenAPI spec for new endpoints

test(trust): add acceptance tests for trust ratings
```

## Getting Help

- **Documentation**: See `api/docs/` for detailed guides
- **Issues**: Open a GitHub issue for bugs or features
- **Discussions**: Use GitHub Discussions for questions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
