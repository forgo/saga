# Development Status

This document tracks the current implementation status of Saga features, known issues, and areas needing work. Updated: January 2025.

## Feature Status Overview

| Feature | Status | Test Coverage | Notes |
|---------|--------|---------------|-------|
| Authentication | ✅ Complete | 13 tests | JWT, OAuth, Passkeys |
| Guilds | ✅ Complete | 13 tests | Create, join, leave, merge |
| Trust Ratings | ✅ Complete | 16 tests | Anchor-based trust system |
| Voting | ✅ Complete | 20 tests | FPTP, RCV, Approval, Multi-select |
| Adventures | ✅ Complete | 15 tests | Multi-day coordination |
| Resonance | ✅ Complete | 9 tests | XP-style rewards |
| Event Verification | ✅ Complete | 6 tests | Attendance confirmation |
| Discovery | ✅ Complete | 6 tests | Profile matching |
| Activities | ✅ Complete | 7 tests | Activity type management |
| People | ✅ Complete | 5 tests | Contact management |
| Timers | ✅ Complete | 5 tests | Activity tracking |
| Role Catalogs | ✅ Complete | 8 tests | Role templates |
| Moderation | ✅ Complete | 13 tests | Blocking, reporting |
| Visibility | ✅ Complete | 6 tests | Access control cascade |
| Location Privacy | ✅ Complete | 10 tests | Coordinate protection |
| Matching Pools | ⚠️ Partial | 10 tests | Model tests only |
| Compatibility | ✅ Complete | 30 tests | OkCupid-style scoring |
| Forums | ✅ Complete | 25 tests | Discussion threads |

**Total: 306 tests passing**

## Implementation Layers

### Fully Wired (Handler → Service → Repository)

These features have complete request-response flow:

- ✅ Authentication (auth, passkey, oauth)
- ✅ Profile management
- ✅ Interest system
- ✅ Availability
- ✅ Discovery
- ✅ Trust ratings
- ✅ Reviews
- ✅ Questionnaire/compatibility
- ✅ Moderation (blocks, reports)

### Partially Wired

These have handlers and services but incomplete wiring in `main.go`:

| Feature | Handler | Service | Repository | Wired |
|---------|---------|---------|------------|-------|
| Guild | ✅ | ⚠️ | ⚠️ | ❌ |
| Event | ✅ | ⚠️ | ✅ | ❌ |
| Adventure | ✅ | ✅ | ✅ | ❌ |
| Pool | ✅ | ⚠️ | ✅ | ❌ |
| Rideshare | ❌ | ❌ | ⚠️ | ❌ |

### Known TODOs in `cmd/server/main.go`

```go
// Line ~75: Guild repository not implemented
// TODO: Implement Guild repository (renamed from Circle)

// Line ~151: Guild service incomplete
// TODO: Implement Guild service (renamed from Circle)

// Line ~217: EventService broken
// TODO: Fix EventRepository interface - missing GetByCircle method

// Line ~226: PoolService broken
// TODO: Fix PoolService - references CircleRepo which no longer exists

// Line ~331: Multiple incomplete handlers
// TODO: Implement Rideshare handler
// TODO: Fix pool handler
// TODO: Implement Adventure handler
```

## Naming Migration Status

The codebase underwent a naming migration:

| Old Name | New Name | Migration Status |
|----------|----------|------------------|
| Circle | Guild | ⚠️ In Progress |
| Baby | Person | ✅ Complete |
| Parent | User | ✅ Complete |
| Family | Guild | ⚠️ In Progress |
| Trip | Adventure | ⚠️ In Progress |

### Files Still Using Old Names

Some repository interfaces still reference old names:
- `CircleRepository` → should be `GuildRepository`
- `GetByCircle` → should be `GetByGuild`

## Database Schema Status

### Tables Implemented

```
user, identity, passkey, token
guild, member
person, activity, timer
trust_rating, endorsement, trust_history
review
vote, vote_option, ballot
adventure, admission
event, event_role, event_role_assignment
pool, pool_member, pool_match
profile, availability, interest
answer, question
block, report
resonance_ledger
forum, forum_post
```

### Triggers & Computed Fields

All triggers are implemented in `migrations/001_initial_schema.surql`:
- Trust aggregate updates
- Member count tracking
- Vote counting
- Resonance ledger immutability

## API Routes Status

### Active Routes (in `main.go`)

```
Auth:
  POST   /v1/auth/register
  POST   /v1/auth/login
  POST   /v1/auth/logout
  POST   /v1/auth/refresh
  POST   /v1/auth/passkey/*
  GET    /v1/auth/oauth/*

Profile:
  GET    /v1/profile
  PUT    /v1/profile
  GET    /v1/profile/progress

Discovery:
  GET    /v1/discover
  POST   /v1/discover/{id}/view

Trust:
  POST   /v1/trust-ratings
  GET    /v1/trust-ratings/received
  GET    /v1/trust-ratings/given

Moderation:
  POST   /v1/blocks
  DELETE /v1/blocks/{id}
  POST   /v1/reports

SSE:
  GET    /v1/events/stream
```

### Routes Defined But Not Wired

These routes are in the router but handlers may not be connected:

```
Guilds:     /v1/guilds/*
Events:     /v1/events/* (partially)
Adventures: /v1/adventures/*
Pools:      /v1/pools/*
Votes:      /v1/votes/*
```

## Testing Infrastructure

### Test Utilities Location

```
api/internal/testing/
├── testdb/testdb.go      # SurrealDB in-memory setup
├── fixtures/fixtures.go  # Test data factories
└── helpers/helpers.go    # JWT, HTTP, assertions
```

### Running Tests

```bash
# All tests
make test

# With coverage report
make test-coverage

# Specific domain
go test -v ./tests/... -run TestAuth

# With race detection
go test -race ./...
```

### Test Database

Tests use SurrealDB in-memory mode:
- Fresh database per test file
- Migrations applied automatically
- No cleanup needed between tests

## Common Development Tasks

### Adding a New Endpoint

1. **Model** (`internal/model/`)
   ```go
   type Widget struct {
       ID   string `json:"id"`
       Name string `json:"name"`
   }
   ```

2. **Repository** (`internal/repository/`)
   ```go
   func (r *WidgetRepository) Create(ctx context.Context, w *Widget) error
   ```

3. **Service** (`internal/service/`)
   ```go
   type WidgetRepository interface { ... }
   func NewWidgetService(cfg WidgetServiceConfig) *WidgetService
   ```

4. **Handler** (`internal/handler/`)
   ```go
   func (h *WidgetHandler) Create(w http.ResponseWriter, r *http.Request)
   ```

5. **Routes** (`cmd/server/main.go`)
   ```go
   r.Post("/v1/widgets", widgetHandler.Create)
   ```

6. **Tests** (`tests/`)
   ```go
   func TestWidget_Create(t *testing.T) { ... }
   ```

7. **OpenAPI** (`openapi/paths/widgets.yaml`)

### Debugging SurrealDB Queries

Enable query logging:
```go
// In repository
log.Printf("Query: %s, Vars: %+v", query, vars)
```

Common issues:
- Record ID format: `user:abc123` not `abc123`
- Record casting: `<record<user>>$user_id`
- NULL vs NONE: Use SET-style queries for optional fields

## Performance Considerations

See `api/docs/PERFORMANCE.md` for detailed tuning guide.

Key points:
- SSE connections need keepalive
- Database connection pooling via SurrealDB driver
- Rate limiting: 100 req/min per user (configurable)

## Security Checklist

See `api/docs/SECURITY.md` for full security model.

Quick reference:
- All endpoints require authentication except `/health`, `/v1/auth/*`
- Location coordinates never exposed via API
- Blocked users hidden bidirectionally
- Report content admin-only visibility

## Getting Help

- **Architecture**: See `api/docs/ARCHITECTURE.md`
- **Database**: See `api/docs/DATABASE.md`
- **Features**: See `api/docs/FEATURES.md`
- **Schema**: See `api/docs/SCHEMA.md`
- **Contributing**: See `CONTRIBUTING.md` in repo root
