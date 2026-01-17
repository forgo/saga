# Saga API

A high-performance Go API for the Saga platform, built with clean architecture principles and modern best practices.

## Quick Start

### Prerequisites

- Go 1.23+
- SurrealDB 2.0+
- Docker (optional, for local development)

### Local Development

1. **Start the database**

   ```bash
   docker-compose up -d
   ```

2. **Configure environment**

   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Run migrations**

   ```bash
   # Migrations are applied automatically on startup
   ```

4. **Start the server**

   ```bash
   go run cmd/server/main.go
   ```

   The API will be available at `http://localhost:8080`

5. **Verify it's running**

   ```bash
   curl http://localhost:8080/health
   ```

## Project Structure

```
api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection layer
│   ├── handler/         # HTTP request handlers
│   ├── middleware/      # HTTP middleware (auth, rate limiting)
│   ├── model/           # Domain entities and data structures
│   ├── repository/      # Data access layer (SurrealDB)
│   ├── service/         # Business logic layer
│   ├── jobs/            # Background job processing
│   └── testing/         # Test utilities and fixtures
├── pkg/jwt/             # JWT token utilities (public package)
├── migrations/          # Database migration files
├── openapi/             # OpenAPI 3.0 specification
└── docs/                # Technical documentation
```

## Architecture

The API follows a clean layered architecture:

```
HTTP Request → Handler → Service → Repository → Database
```

- **Handlers**: HTTP protocol handling, request validation, response formatting
- **Services**: Business logic, validation rules, orchestration
- **Repositories**: Data access, SurrealDB queries, entity mapping

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture documentation.

## Configuration

Configuration is loaded from environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DATABASE_URL` | SurrealDB connection URL | `ws://localhost:8000/rpc` |
| `DATABASE_USER` | Database username | `root` |
| `DATABASE_PASS` | Database password | `root` |
| `DATABASE_NS` | Database namespace | `saga` |
| `DATABASE_DB` | Database name | `development` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `JWT_EXPIRATION` | Token expiration | `24h` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | (optional) |
| `APPLE_CLIENT_ID` | Apple OAuth client ID | (optional) |

## API Documentation

The API is documented using OpenAPI 3.0. View the specification:

- [openapi/openapi.yaml](openapi/openapi.yaml) - Main specification
- [openapi/paths/](openapi/paths/) - Endpoint definitions
- [openapi/components/](openapi/components/) - Reusable schemas

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run acceptance tests only
go test ./tests/...

# Run specific test
go test ./tests/... -run TestGuild
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

### Building

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Build Docker image
docker build -t saga-api .
```

## Key Features

- **Authentication**: JWT tokens, OAuth (Google, Apple), Passkeys (WebAuthn)
- **Guilds**: Community groups with roles (admin, moderator, member)
- **Events**: Scheduled activities with RSVP management
- **Adventures**: Special events with admission control
- **Voting**: Multiple vote types (FPTP, ranked choice, approval)
- **Trust Network**: User trust ratings and endorsements
- **Discovery**: Location-based and interest-based matching
- **Real-time**: Server-Sent Events for live updates

## Error Handling

The API uses [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457.html) for error responses:

```json
{
  "type": "https://api.saga.app/errors/not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "Guild with ID 'guild:abc123' was not found",
  "instance": "/v1/guilds/guild:abc123"
}
```

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [DATABASE.md](docs/DATABASE.md) - Database schema
- [SECURITY.md](docs/SECURITY.md) - Security practices
- [DEVELOPMENT.md](docs/DEVELOPMENT.md) - Development guide
- [PERFORMANCE.md](docs/PERFORMANCE.md) - Performance considerations

## Contributing

1. Follow Go conventions and project patterns
2. Write tests for new functionality
3. Update OpenAPI spec for API changes
4. Run `go test ./...` before submitting

## License

Proprietary - All rights reserved
