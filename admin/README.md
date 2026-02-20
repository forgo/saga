# Saga Admin

Admin web application for Saga, built with
[Deno Fresh](https://fresh.deno.dev/).

## Overview

The admin app provides a full admin suite for user/content management and
analytics/monitoring. It communicates with the Go API backend and does not
access SurrealDB directly.

**Architecture:** Admin → Go API → SurrealDB

## Prerequisites

- [Deno](https://docs.deno.com/runtime/getting_started/installation) v2.x+

## Development

```bash
# Start development server (http://localhost:8000)
make dev

# Or using deno directly
deno task dev
```

## Commands

| Command      | Description                      |
| ------------ | -------------------------------- |
| `make dev`   | Start dev server with hot reload |
| `make test`  | Run tests                        |
| `make lint`  | Lint code                        |
| `make fmt`   | Format code                      |
| `make check` | Run lint and tests               |
| `make build` | Build for production             |
| `make clean` | Remove build artifacts           |

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

| Variable       | Description     | Default                    |
| -------------- | --------------- | -------------------------- |
| `API_BASE_URL` | Go API base URL | `http://localhost:8080/v1` |

## Project Structure

```
admin/
├── main.ts             # Server entry point
├── client.ts           # Client-side entry (loaded on every page)
├── vite.config.ts      # Vite build configuration
├── deno.json           # Dependencies and tasks
├── routes/
│   ├── _app.tsx        # App wrapper (HTML structure)
│   └── index.tsx       # Home page
├── islands/            # Interactive components (hydrated on client)
├── components/         # Static UI components (server-rendered only)
├── static/             # Static assets (CSS, images)
└── assets/             # Processed assets (CSS with Tailwind)
```

## Architecture

The admin app follows the Fresh framework conventions:

- **Routes** (`routes/`) - File-based routing, server-rendered by default
- **Islands** (`islands/`) - Interactive components that hydrate on the client
- **Components** (`components/`) - Static components, server-rendered only

All data access goes through the Go API. The admin app does not have direct
database access, ensuring:

- Single source of truth for business logic
- Consistent authentication and authorization
- Reuse of existing API validation
