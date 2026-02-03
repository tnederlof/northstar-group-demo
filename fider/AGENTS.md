# Fider Application

## Overview

Go+React feedback portal. CQRS backend, SSR frontend, PostgreSQL, no ORM.

## Key Code Touchpoints

- **Routes**: `app/cmd/routes.go`
- **Handlers**: `app/handlers/` (organized by feature)
- **Services/DB**: `app/services/sqlstore/postgres/`
- **Migrations**: `migrations/*.up.sql`
- **Frontend pages**: `public/pages/`
- **Frontend components**: `public/components/`

## Build Targets

- `make build` - full build
- `make watch` - dev mode
- `make test` - all tests
- `make test-server`, `test-ui`, `test-e2e-ui`
- `make lint`

## Demo Mode Extensions

- `EMAIL=none`: `app/pkg/email/noop/noop.go`
- Demo login: `app/handlers/demo/login.go`
- Env config: `app/pkg/env/env.go` (`DEMO_MODE`, `DEMO_LOGIN_KEY`)
