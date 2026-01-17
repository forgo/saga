# Saga

A social coordination platform for meaningful connections.

## Overview

Saga combines the best elements of community platforms:

- **Guilds** - Social groups with shared activities and contacts
- **Adventures** - Multi-day, multi-location coordination for groups
- **Events** - Host-controlled gatherings with concrete times and places
- **Discovery** - Find compatible people through questionnaire matching
- **Trust Network** - Event-anchored trust ratings with endorsements

## Architecture

```
saga/
├── api/              # Go backend (SurrealDB, REST API, SSE)
│   ├── cmd/          # Application entrypoints
│   ├── internal/     # Private application code
│   ├── migrations/   # Database migrations
│   ├── openapi/      # API specification
│   └── docs/         # Technical documentation
├── ios/              # SwiftUI iOS app
└── .github/          # CI/CD workflows
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.23 with Chi router |
| Database | SurrealDB 2.1.0 |
| Auth | JWT + Passkeys (WebAuthn) + OAuth |
| iOS | SwiftUI, iOS 17+ |
| CI/CD | GitHub Actions |

## Quick Start

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Make
- Xcode 16+ (for iOS)

### Setup

```bash
# Clone and setup
git clone https://github.com/forgo/saga.git
cd saga
make setup

# Start development environment
make dev

# Run tests
make test
```

### Commands

| Command | Description |
|---------|-------------|
| `make setup` | First-time developer setup |
| `make dev` | Start API and SurrealDB |
| `make stop` | Stop all services |
| `make test` | Run all tests (306 tests) |
| `make lint` | Lint code |
| `make test-coverage` | Run tests with coverage report |

## Documentation

### API Documentation

| Document | Description |
|----------|-------------|
| [OpenAPI Spec](api/openapi/) | REST API specification (OpenAPI 3.1) |
| [ARCHITECTURE.md](api/docs/ARCHITECTURE.md) | System design and layers |
| [DATABASE.md](api/docs/DATABASE.md) | Database patterns and queries |
| [FEATURES.md](api/docs/FEATURES.md) | Feature implementation guide |
| [SCHEMA.md](api/docs/SCHEMA.md) | Database schema reference |
| [SECURITY.md](api/docs/SECURITY.md) | Security model and access control |
| [PERFORMANCE.md](api/docs/PERFORMANCE.md) | Performance tuning |
| [DEVELOPMENT.md](api/docs/DEVELOPMENT.md) | Development status and TODOs |

### Project Documentation

| Document | Description |
|----------|-------------|
| [SAGA.md](SAGA.md) | Product vision and feature design |
| [CHANGELOG.md](CHANGELOG.md) | Release history |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines |
| [VERSIONING.md](VERSIONING.md) | API versioning strategy |

### iOS Documentation

| Document | Description |
|----------|-------------|
| [ios/README.md](ios/README.md) | iOS app setup and architecture |

## API Endpoints

The API uses URL path versioning (`/v1/...`). Key endpoints:

```
Authentication:
  POST /v1/auth/register     - Email/password registration
  POST /v1/auth/login        - Login with credentials
  POST /v1/auth/passkey/*    - Passkey authentication
  GET  /v1/auth/oauth/*      - OAuth flows

Guilds:
  GET  /v1/guilds            - List user's guilds
  POST /v1/guilds            - Create guild
  POST /v1/guilds/{id}/join  - Join guild

Trust:
  POST /v1/trust-ratings     - Create trust rating
  GET  /v1/trust-ratings/*   - Query trust ratings

Real-time:
  GET  /v1/guilds/{id}/events - SSE event stream
```

Full specification: [api/openapi/openapi.yaml](api/openapi/openapi.yaml)

## Testing

The project includes 306 tests across 18 domains:

```bash
# Run all tests
make test

# Run specific domain tests
go test -v ./tests/... -run TestAuth
go test -v ./tests/... -run TestGuild
go test -v ./tests/... -run TestTrust

# With coverage
make test-coverage
```

Test domains: Auth, Guilds, Trust, Voting, Adventures, Resonance, Events, Discovery, Activities, People, Timers, Roles, Moderation, Visibility, Location Privacy, Matching Pools, Compatibility, Forums.

## Release Process

Releases are automated via GitHub Actions:

1. Update [CHANGELOG.md](CHANGELOG.md)
2. Create and push a tag: `git tag v1.2.3 && git push --tags`
3. CI runs tests, builds binaries, publishes Docker image

See [VERSIONING.md](VERSIONING.md) for versioning policy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Code architecture
- Testing requirements
- Pull request process
- Commit message format

## License

MIT License - see [LICENSE](LICENSE) for details.
