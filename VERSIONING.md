# Versioning Strategy

This document describes the versioning strategy for the Saga platform, including the API, clients, and releases.

## Semantic Versioning

Saga follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

| Component | When to Increment | Example |
|-----------|------------------|---------|
| **MAJOR** | Breaking API changes | `1.0.0` → `2.0.0` |
| **MINOR** | New features (backward compatible) | `1.0.0` → `1.1.0` |
| **PATCH** | Bug fixes (backward compatible) | `1.0.0` → `1.0.1` |

### Pre-release Versions

During initial development (0.x.x), minor versions may contain breaking changes:

- `0.1.0` → `0.2.0`: May include breaking changes
- `1.0.0`+: Stable, follows strict semver

## API Versioning

### URL Path Versioning

The REST API uses URL path versioning:

```
https://api.saga.app/v1/guilds
https://api.saga.app/v2/guilds  (future)
```

### Current Version

- **API Version**: `v1`
- **OpenAPI Spec Version**: `1.0.0`

### Version Lifecycle

| Stage | Duration | Description |
|-------|----------|-------------|
| **Current** | Active | Fully supported, receives features and fixes |
| **Deprecated** | 6 months | Receives security fixes only, migration warnings |
| **Sunset** | - | No longer available |

### Deprecation Policy

When a new major API version is released:

1. **Announcement**: 3 months before deprecation begins
2. **Deprecation Headers**: API returns `Deprecation` and `Sunset` headers
3. **Migration Guide**: Published with upgrade instructions
4. **Grace Period**: 6 months with both versions available
5. **Sunset**: Old version removed

Example deprecation headers:

```http
Deprecation: true
Sunset: Sat, 01 Jan 2026 00:00:00 GMT
Link: <https://docs.saga.app/migration/v1-to-v2>; rel="deprecation"
```

## Breaking Changes

### What Constitutes a Breaking Change

**Breaking** (requires MAJOR version bump):
- Removing an endpoint
- Removing a required field from responses
- Adding a required field to requests
- Changing field types
- Changing authentication requirements
- Changing error response formats

**Non-Breaking** (MINOR version bump):
- Adding new endpoints
- Adding optional fields to requests
- Adding new fields to responses
- Adding new enum values (if clients handle unknown values)
- Performance improvements
- Bug fixes that don't change API contract

### Avoiding Breaking Changes

Prefer additive changes:

```go
// Instead of renaming:
// "name" → "display_name"  ❌ Breaking

// Add new field, deprecate old:
{
  "name": "...",           // Deprecated
  "display_name": "..."    // New
}
```

## Client Compatibility

### iOS Client

The iOS client version is independent of the API version:

| iOS App | Minimum API | Notes |
|---------|-------------|-------|
| 1.x | v1 | Initial release |

The iOS client:
- Sends `Accept: application/json` with API version preference
- Handles unknown response fields gracefully
- Displays upgrade prompts when API version is deprecated

### API Client Headers

Clients should send:

```http
Accept: application/json
User-Agent: Saga-iOS/1.0.0
```

## Release Process

### Creating a Release

1. Update `CHANGELOG.md` with changes
2. Commit: `git commit -m "chore: prepare release vX.Y.Z"`
3. Tag: `git tag vX.Y.Z`
4. Push: `git push origin main --tags`

### Automated Release

On tag push, CI automatically:
1. Runs full test suite
2. Builds cross-platform binaries
3. Builds and pushes Docker image
4. Creates GitHub Release with notes

### Version Injection

The build injects version into the binary:

```bash
go build -ldflags="-X main.Version=v1.2.3" ./cmd/server
```

Access in code:

```go
var Version = "development"

func main() {
    log.Printf("Saga API %s", Version)
}
```

## Database Migrations

### Migration Versioning

Migrations use sequential numbering:

```
migrations/
├── 001_initial_schema.surql
├── 002_add_trust_ratings.surql
└── 003_add_voting_system.surql
```

### Migration Policy

- Migrations are **append-only** (never modify existing migrations)
- Each migration includes both `up` and rollback comments
- Migrations run automatically on startup
- Breaking schema changes require application-level migration support

## OpenAPI Specification

### Spec Versioning

The OpenAPI spec version (`info.version`) tracks API changes:

```yaml
openapi: 3.1.0
info:
  title: Saga API
  version: 1.0.0  # API version
```

### Updating the Spec

1. Update `openapi/paths/*.yaml` for endpoint changes
2. Update `openapi/components/schemas/` for model changes
3. Increment `info.version` in `openapi/openapi.yaml`
4. Generate client code: `make generate`

## Compatibility Matrix

| API Version | SurrealDB | Go | iOS SDK |
|-------------|-----------|-----|---------|
| v1 | 2.1.0+ | 1.23+ | 17+ |

## Questions?

For versioning questions:
- Open a GitHub Discussion
- See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines
- Check [CHANGELOG.md](CHANGELOG.md) for version history
