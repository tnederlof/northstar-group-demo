# Fider Application

## Overview

Go+React feedback portal. CQRS backend, SSR frontend, PostgreSQL, no ORM.

## Directory Structure

**Backend (Go):**
- `app/handlers/` - HTTP request handlers
- `app/models/` - Data models (entity, cmd, query, action, dto)
- `app/services/` - Business logic and service integrations
- `app/pkg/bus/` - Service registry and dispatch system
- `app/cmd/routes.go` - All HTTP routes defined here
- `migrations/` - Database migrations (numbered SQL files)

**Frontend (React/TypeScript):**
- `public/pages/` - Page components (lazy-loaded)
- `public/components/` - Reusable UI components
- `public/services/` - Client-side services and API calls
- `public/hooks/` - Custom React hooks
- `public/assets/styles/` - SCSS styles with utility classes

**Configuration:**
- `.env` - Local environment config (copy from `.example.env`)
- `Makefile` - Build and development commands

## Build & Development

- `make build` - Production build
- `make watch` - Hot reload for server and UI (use for active development)
- `make migrate` - Run database migrations
- `make test` - All tests (Go + Jest)
- `make lint` - Lint server and UI code
- `make test-server`, `test-ui`, `test-e2e-ui` - Specific test suites

**Essential:** Always run `make lint` and `make test` after completing changes.

## Architecture Patterns

**Backend:**
- CQRS: Separate commands (write) and queries (read)
- Bus system: Dependency injection and service dispatch
- Middleware chain: Auth, tenant resolution, CORS
- Direct SQL queries, no ORM

**Frontend:**
- SSR Support: React 18 hydration
- i18n: LinguiJS for translations
- Code splitting: Lazy-loaded pages
- Type safety: Full TypeScript coverage

## Model Naming Conventions

Strict prefixes in `app/models/`:
- `entity.*` - Database tables
- `action.*` - User input for POST/PUT/PATCH
- `cmd.*` - Commands to execute
- `query.*` - Queries to fetch data
- `dto.*` - Data transfer between packages

## CSS Conventions

**BEM style** for page-specific styles:
```scss
// Page ID: p-<page_name>
#p-home {
  // Component: c-<component>
  .c-post-list {
    // Element: c-<component>__<element>
    &__item { padding: 1rem; }
    // Modifier: c-<component>--<state>
    &--loading { opacity: 0.5; }
  }
}
```

**Utility classes** (no prefix) - Prefer these over custom styles:
- Defined in `public/assets/styles/utility/`
- Similar to Tailwind but project-specific
- Check existing utilities before adding new styles

**Important:** Do not use Tailwind classes - check available utility SCSS files.

## Adding API Endpoints

1. **Define route** in `app/cmd/routes.go`
2. **Create handler** in `app/handlers/`
3. **Define query/command** in `app/models/query/` or `app/models/cmd/`
4. **Implement service** in `app/services/postgres/`
5. **Register handler** via `bus.AddHandler()` in service `Init()`

## Bus System Usage

```go
// Dispatch a query
q := &query.GetUserByID{UserID: 123}
if err := bus.Dispatch(ctx, q); err != nil {
    return err
}
user := q.Result

// Dispatch a command
c := &cmd.SendEmail{To: "user@example.com"}
if err := bus.Dispatch(ctx, c); err != nil {
    return err
}
```

## Database Migrations

"Up migrations" only. Format: `YYYYMMDDHHMMSS_description.sql`
Place in `migrations/` directory, run with `make migrate`.

## Troubleshooting

**Database errors:**
- Ensure Docker running: `docker compose ps`
- Check `.env` DATABASE_URL
- Run `make migrate`

**Build failures:**
- Clear cache: `make clean && make build`
- Check Go 1.22+ and Node 21/22

**Port conflicts:**
- Defaults: 3000 (app), 5432 (postgres), 8025 (mailhog)
- Change in `.env` if needed

## Demo Mode Extensions

- `EMAIL=none`: `app/pkg/email/noop/noop.go`
- Demo login: `app/handlers/demo/login.go`
- Env config: `app/pkg/env/env.go` (`DEMO_MODE`, `DEMO_LOGIN_KEY`)
