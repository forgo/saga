# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive E2E test suite (306 tests across 18 domains)
- Test infrastructure with SurrealDB in-memory support
- Acceptance criteria documentation in test files

### Changed
- Repository fixes for SurrealDB record casting
- Moderation repository NULL vs NONE handling

### Fixed
- `parseTime()` now handles `models.CustomDateTime` from SurrealDB
- Report/Block creation uses SET-style queries to avoid NULL issues

## [0.6.0] - 2025-01-XX

### Added
- Modern SwiftUI iOS app (Phase 6)
- Native iOS 17+ support
- Passkey authentication on iOS
- Real-time SSE event handling

### Changed
- iOS app completely rewritten in SwiftUI
- Removed UIKit dependencies

## [0.5.0] - 2025-01-XX

### Added
- Production infrastructure (Phase 5)
- Docker multi-stage builds
- GitHub Container Registry publishing
- CI/CD workflows for testing and releases
- Dependabot for dependency updates

### Changed
- Dockerfile uses scratch base for minimal image size
- Cross-platform binary builds (Linux AMD64/ARM64, macOS ARM64)

## [0.4.0] - 2025-01-XX

### Added
- Family operations and real-time updates (Phase 4)
- Server-Sent Events (SSE) for live updates
- Guild merge functionality
- Activity and timer management

### Changed
- Renamed "Circle" to "Guild" throughout codebase
- Renamed "Baby" to "Person" for flexibility

## [0.3.0] - 2025-01-XX

### Added
- Core resources CRUD operations (Phase 3)
- Guild management (create, join, leave)
- People/contacts management
- Activity types with thresholds
- Timer tracking system

## [0.2.0] - 2025-01-XX

### Added
- Authentication system (Phase 2)
- JWT token-based auth with refresh tokens
- Passkey/WebAuthn support
- OAuth integration (Google, Apple)
- Rate limiting middleware
- CORS configuration

### Security
- bcrypt password hashing
- RSA JWT signing
- Token rotation on refresh

## [0.1.0] - 2025-01-XX

### Added
- Go API foundation (Phase 1)
- SurrealDB integration
- Repository pattern implementation
- Layered architecture (handler → service → repository)
- Configuration management
- Health check endpoint

### Infrastructure
- Go 1.23 with Chi router
- Docker Compose for local development
- Makefile build system
- OpenAPI 3.1.0 specification

---

## Versioning Policy

This project uses [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking API changes
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, backward compatible

### API Versioning

The REST API uses URL path versioning:
- Current: `/v1/...`
- Future versions will use `/v2/...`, etc.

See [VERSIONING.md](VERSIONING.md) for detailed versioning strategy.

[Unreleased]: https://github.com/forgo/saga/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/forgo/saga/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/forgo/saga/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/forgo/saga/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/forgo/saga/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/forgo/saga/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/forgo/saga/releases/tag/v0.1.0
