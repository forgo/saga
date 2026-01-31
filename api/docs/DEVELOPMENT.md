# Development Status

This document tracks the current implementation status of Saga features and development workflows. Updated: January 2025.

## Feature Status Overview

| Feature | Status | Notes |
|---------|--------|-------|
| Authentication | ✅ Complete | JWT, OAuth (Google/Apple), Passkeys (WebAuthn) |
| Guilds | ✅ Complete | Create, join, leave, roles, alliances |
| Events | ✅ Complete | CRUD, RSVPs, roles, attendance verification |
| Adventures | ✅ Complete | Multi-day coordination, admission control |
| Discovery | ✅ Complete | People/event discovery with geo & compatibility |
| Profiles | ✅ Complete | Privacy controls, location privacy buckets |
| Availability | ✅ Complete | Hangout scheduling with 5 hangout types |
| Questionnaire | ✅ Complete | OkCupid-style compatibility matching |
| Interests | ✅ Complete | Teach/learn matching |
| Trust | ✅ Complete | Trust grants, IRL verification |
| Trust Ratings | ✅ Complete | Anchor-based trust with endorsements |
| Reviews | ✅ Complete | Context-based peer reviews |
| Resonance | ✅ Complete | XP-style gamification with daily caps |
| Moderation | ✅ Complete | Reports, blocks, actions, admin tools |
| Matching Pools | ✅ Complete | Donut-style random matching |
| Role Catalogs | ✅ Complete | Reusable role templates |
| Voting | ✅ Complete | FPTP, RCV, Approval, Multi-select |
| Devices | ✅ Complete | Push notification token management |

## Implementation Layers

All features follow the layered architecture:

```
Handler → Service → Repository → Database
```

### Layer Files

| Layer | Count | Location |
|-------|-------|----------|
| Handlers | 25+ | `internal/handler/` |
| Services | 24+ | `internal/service/` |
| Repositories | 24+ | `internal/repository/` |
| Models | 22+ | `internal/model/` |
| Middleware | 6 | `internal/middleware/` |

### Key Services

- **AuthService** - Email/password, OAuth, session management
- **PasskeyService** - WebAuthn registration and authentication
- **OAuthService** - Google/Apple sign-in with PKCE
- **GuildService** - Guild membership and administration
- **EventService** - Event creation, RSVPs, completion verification
- **DiscoveryService** - Profile and event discovery with geo-filtering
- **CompatibilityService** - OkCupid-style scoring
- **ResonanceService** - Gamification scoring with daily caps
- **TrustService** - Trust grants and IRL verification
- **ModerationService** - Reports, blocks, admin actions

## Database Schema Status

### Tables Implemented (35+)

```
Auth:        user, identity, passkey, refresh_token
Profile:     user_profile, answer, user_bias_profile
Guild:       guild, member, guild_alliance, guild_moderation_settings
Events:      event, event_role, event_role_assignment, unified_rsvp
Adventures:  adventure, destination, adventure_activity, adventure_admission
Rideshares:  rideshare, rideshare_segment, rideshare_seat
Discovery:   availability, discovery_daily_count
Trust:       trust_relation, irl_verification, trust_rating, trust_endorsement
Reviews:     review
Resonance:   resonance_ledger, resonance_score, resonance_daily_cap
Moderation:  report, moderation_action, block, user_flag
Pools:       matching_pool, pool_member, match_result
Roles:       role_catalog, rideshare_role, rideshare_role_assignment
Voting:      vote, vote_option, vote_ballot
Interests:   interest, has_interest
Forums:      forum, forum_post
```

### Database Features

- **43 Triggers** - Validation, auto-updates, cascades
- **11 Functions** - Location privacy, access control
- **165+ Indexes** - Query optimization

## API Routes Status

### All Routes Active

**Auth:**
```
POST   /v1/auth/register
POST   /v1/auth/login
POST   /v1/auth/logout
POST   /v1/auth/refresh
POST   /v1/auth/passkey/register/start
POST   /v1/auth/passkey/register/finish
POST   /v1/auth/passkey/login/start
POST   /v1/auth/passkey/login/finish
GET    /v1/auth/oauth/google
GET    /v1/auth/oauth/apple
```

**Profile & Discovery:**
```
GET    /v1/me/profile
PUT    /v1/me/profile
GET    /v1/discover/people
GET    /v1/discover/events
GET    /v1/discover/interest/{interestId}
GET    /v1/users/{userId}/profile
```

**Guilds:**
```
GET    /v1/guilds
POST   /v1/guilds
GET    /v1/guilds/{guildId}
PATCH  /v1/guilds/{guildId}
DELETE /v1/guilds/{guildId}
POST   /v1/guilds/{guildId}/join
POST   /v1/guilds/{guildId}/leave
GET    /v1/guilds/{guildId}/members
```

**Events:**
```
POST   /v1/guilds/{guildId}/events
GET    /v1/guilds/{guildId}/events
GET    /v1/events/{eventId}
PATCH  /v1/events/{eventId}
DELETE /v1/events/{eventId}
POST   /v1/events/{eventId}/rsvp
POST   /v1/events/{eventId}/confirm
POST   /v1/events/{eventId}/feedback
```

**Availability & Hangouts:**
```
GET    /v1/availability
POST   /v1/availability
GET    /v1/availability/nearby
POST   /v1/availability/{availabilityId}/request
POST   /v1/hangout-requests/{requestId}/respond
GET    /v1/me/hangouts
```

**Trust & Reviews:**
```
POST   /v1/trust
GET    /v1/me/trust
POST   /v1/trust/irl/request
POST   /v1/trust/irl/{requestId}/confirm
POST   /v1/reviews
GET    /v1/users/{userId}/reputation
```

**Moderation:**
```
POST   /v1/reports
POST   /v1/blocks
DELETE /v1/blocks/{userId}
GET    /v1/me/moderation-status
```

**And more:** Resonance, Pools, Interests, Questionnaire, Votes, Devices

## Testing Infrastructure

### Test Files

Service-level tests are located alongside their implementations:

```
internal/service/
├── auth.go
├── auth_test.go          # 17 authentication tests
├── oauth.go
├── oauth_test.go         # OAuth flow tests
├── passkey.go
├── passkey_test.go       # WebAuthn tests
├── guild.go
├── guild_test.go         # Guild operation tests
├── event.go
├── event_test.go         # Event lifecycle tests
└── ...
```

### Running Tests

```bash
# All tests
make test

# With coverage report
make test-coverage

# Specific service
go test -v ./internal/service/... -run TestAuth

# With race detection
go test -race ./...

# Build verification
make build
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

6. **Tests** (`internal/service/`)
   ```go
   func TestWidgetService_Create(t *testing.T) { ... }
   ```

7. **OpenAPI** (`openapi/paths/widgets.yaml`)

### Error Handling

The codebase uses centralized error handling:

**Service errors** (`internal/service/errors.go`):
```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already registered")
    // ... ~70 service-specific errors
)
```

**Handler error mapping** (`internal/handler/error_mapper.go`):
```go
// Automatically maps service errors to HTTP status codes
func MapServiceError(err error, defaultMsg string) *model.ProblemDetails {
    switch {
    case errors.Is(err, service.ErrUserNotFound):
        return model.NewNotFoundError("user", err.Error())
    // ... handles 401, 403, 404, 409, 422 mapping
    }
}
```

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

## OpenAPI Documentation

The API is fully documented in OpenAPI 3.1.0 format:

```
openapi/
├── openapi.yaml              # Main spec with all path references
├── components/
│   └── schemas/_index.yaml   # 100+ schema definitions
└── paths/
    ├── auth.yaml             # Authentication endpoints
    ├── guilds.yaml           # Guild management
    ├── events.yaml           # Event lifecycle
    ├── discovery.yaml        # People/event discovery
    ├── availability.yaml     # Hangout scheduling
    ├── profiles.yaml         # User profiles
    ├── interests.yaml        # Interest management
    ├── questionnaire.yaml    # Compatibility questions
    ├── trust.yaml            # Trust grants/IRL
    ├── reviews.yaml          # Peer reviews
    ├── resonance.yaml        # Gamification
    ├── moderation.yaml       # Reports/blocks
    ├── pools.yaml            # Random matching
    ├── devices.yaml          # Push notifications
    └── ...                   # 20+ path files
```

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
