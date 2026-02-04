# Demo Environment Contract

This document defines the contracts and conventions used across demo scenarios to ensure consistency and predictability.

## URL Contract

All scenarios expose services at consistent URLs following the pattern:

```
http://<slug>.localhost:8080
```

Where `<slug>` is the second component of the scenario path (e.g., `bad-rollout` from `platform/bad-rollout`).

### SRE Track

When running in kind cluster:

| Scenario | URL |
|----------|-----|
| `platform/healthy` | `http://healthy.localhost:8080` |
| `platform/bad-rollout` | `http://bad-rollout.localhost:8080` |
| `platform/resource-exhaustion` | `http://resource-exhaustion.localhost:8080` |

Routing is handled by:
- Kind cluster with hostPort 8080 mapped to nodePort 30080
- Envoy Gateway listening on `*.localhost` hostname
- Per-scenario HTTPRoute resources

**Note**: `.localhost` domains resolve to 127.0.0.1 automatically (no `/etc/hosts` entries needed).

### Engineering Track

When running with Docker Compose:

| Scenario | URL |
|----------|-----|
| `backend/ui-regression` | `http://ui-regression.localhost:8080` |

Routing is handled by:
- Traefik edge proxy on host port 8080
- Per-scenario container labels defining routing rules

## Environment Variable Contract

### Required Variables

All scenarios require these environment variables:

| Variable | Purpose | Example Value |
|----------|---------|---------------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/fider` |
| `JWT_SECRET` | Session token signing | `random-secret-key-for-demo` |
| `EMAIL` | Email provider type | `none` (for demo mode) |

### Optional Variables

| Variable | Purpose | Default | Notes |
|----------|---------|---------|-------|
| `GO_ENV` | Environment mode | `development` | Set to `production` for production builds |
| `LOG_LEVEL` | Logging verbosity | `info` | Values: `debug`, `info`, `warn`, `error` |
| `PORT` | HTTP server port | `3000` | Application listening port |
| `HOST_DOMAIN` | Base domain | `localhost` | Used for URL generation |

### Demo-Specific Variables

These variables are specific to demo mode:

| Variable | Purpose | Default | Notes |
|----------|---------|---------|-------|
| `DEMO_MODE` | Enable demo features | `true` | Enables demo login endpoint |
| `DEMO_ADMIN_EMAIL` | Admin persona email | `alex@demo.local` | Pre-seeded admin user |
| `DEMO_USER_EMAIL` | Regular user email | `sarah@demo.local` | Pre-seeded regular user |

## Seed Data Reference

All scenarios start with consistent seed data to ensure reproducibility.

### Users

| ID | Name | Email | Role | Password |
|----|------|-------|------|----------|
| 1 | Alex Chen | alex@demo.local | Administrator | `demo123` |
| 2 | Sarah Johnson | sarah@demo.local | Collaborator | `demo123` |
| 3 | Marcus Williams | marcus@demo.local | Visitor | `demo123` |
| 4 | Jennifer Lee | jennifer@demo.local | Visitor | `demo123` |

**Note**: In demo mode, you can also use the `/__demo/login/:persona` endpoint to log in without passwords.

### Tenants

| ID | Name | Subdomain | Status |
|----|------|-----------|--------|
| 1 | Demo Company | demo | Active |

### Posts (Feature Requests)

| ID | Title | Status | Votes |
|----|-------|--------|-------|
| 1 | Add dark mode | Open | 42 |
| 2 | Mobile app support | Planned | 35 |
| 3 | SSO integration | Started | 28 |
| 4 | API webhooks | Completed | 15 |

## API Contract

### Authentication

All API endpoints require authentication except:
- `GET /api/health` - Health check
- `POST /api/signin` - Sign in
- `POST /api/tenants` - Create tenant

Authentication methods:
1. **Session Cookie**: Set after successful signin
2. **API Key**: Pass in `Authorization: Bearer <key>` header
3. **Demo Mode**: Use `/__demo/login/:persona` endpoint

### Common Response Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid input data |
| 401 | Unauthorized | Not authenticated |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Server-side error |

### Error Response Format

All error responses follow this format:

```json
{
  "error": "Short error description",
  "details": "Detailed explanation of what went wrong",
  "code": "ERROR_CODE"
}
```

## Database Schema Contract

### Migration Numbering

Migrations follow the format: `YYYYMMDDHHMM_description.sql`

Example: `202601151430_add_feature_flags.sql`

**Rules**:
- Use UTC timestamps
- Migrations must be idempotent
- Always include both `up` and `down` (if reversible)
- Test migrations locally before committing

### Reserved Tables

These tables are managed by the system and should not be modified directly:

- `schema_migrations` - Migration tracking
- `tenants` - Tenant configuration
- `users` - User accounts
- `posts` - Feature requests
- `comments` - Post comments
- `votes` - User votes
- `tags` - Post categorization

## File System Contract

### State Directory

Demo state is stored in `.state/` (gitignored):

```
.state/
├── sre/
│   ├── cluster-config.yaml
│   └── scenarios/
│       └── platform/bad-rollout/
│           └── state.json
└── engineering/
    └── scenarios/
        └── backend/api-regression/
            └── state.json
```

### Worktree Directory

Git worktrees for Engineering scenarios:

```
.worktrees/
├── backend/
│   ├── api-regression/
│   └── migration-conflict/
└── frontend/
    └── error-boundary/
```

**Note**: Both `.state/` and `.worktrees/` are gitignored.

## Port Allocations

To avoid conflicts, scenarios use these port assignments:

### Engineering Track

| Service | Port | Protocol |
|---------|------|----------|
| Fider HTTP | 80 | HTTP |
| Fider HTTPS | 443 | HTTPS |
| PostgreSQL | 5432 | TCP |
| Fider Dev Server | 3000 | HTTP |
| Webpack Dev Server | 3001 | HTTP |
| Traefik Dashboard | 8080 | HTTP |

### SRE Track

| Service | Port | Protocol |
|---------|------|----------|
| Fider (via Gateway) | 80 | HTTP |
| Envoy Admin | 19001 | HTTP |
| Grafana | 3000 | HTTP |
| Prometheus | 9090 | HTTP |

## Breaking the Contract

If you need to deviate from these contracts for a specific scenario:

1. Document the deviation in the scenario's README
2. Explain why it's necessary
3. Provide clear instructions for users
4. Consider if the contract should be updated for all scenarios

## Contract Versioning

This contract is versioned alongside the demo codebase. Changes to the contract should:

- Be backward compatible when possible
- Be clearly documented in CHANGELOG
- Include migration guides for breaking changes
