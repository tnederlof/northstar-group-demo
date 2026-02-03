# Architecture

This document provides a technical overview of the Fider application and the demo environment architecture.

## Fider Application Overview

Fider is a feedback management platform built with modern technologies:

- **Backend**: Go 1.24 with CQRS (Command Query Responsibility Segregation) pattern
- **Frontend**: React 18 + TypeScript
- **Database**: PostgreSQL 17, using raw SQL (no ORM)
- **Features**: Server-Side Rendering (SSR), LinguiJS for internationalization (i18n)

### Key Code Touchpoints

Understanding where to find specific functionality:

- **Routes**: `app/cmd/routes.go` - HTTP route definitions and middleware setup
- **Handlers**: `app/handlers/` - HTTP request handlers organized by domain
- **Services/DB**: `app/services/sqlstore/postgres/` - Database queries and data access
- **Migrations**: `migrations/` - SQL migration files for schema evolution
- **Frontend Pages**: `public/pages/` - React page components
- **Frontend Components**: `public/components/` - Reusable React components

## Demo Mode Extensions

The demo environment includes custom extensions to Fider for demonstration purposes:

### EMAIL=none Noop Provider

A no-operation email provider that satisfies the email interface without sending actual messages:
- Logs outgoing emails at debug level
- Returns success for all send operations
- Requires no SMTP/API credentials
- Enabled by setting `EMAIL=none` environment variable

### Demo Login Endpoint

Special authentication endpoint for demo personas:
- **Endpoint**: `/__demo/login/:persona`
- **Purpose**: Quick login as different user types without email verification
- **Available Personas**: admin, user, moderator, etc.
- **Only Active**: When running in demo mode

## Runtime Architectures

The demo supports two distinct runtime architectures for different learning tracks:

### SRE Track

**Purpose**: Demonstrate cloud-native infrastructure and GitOps practices

**Stack**:
- **Kubernetes**: kind (Kubernetes in Docker) cluster
- **Gateway API**: Envoy Gateway for ingress and traffic management
- **Service Mesh**: Demonstrates advanced networking patterns
- **Infrastructure as Code**: Kubernetes manifests for declarative configuration

**Use Cases**:
- Service deployment and scaling
- Traffic routing and load balancing
- Observability and monitoring
- GitOps workflows

### Engineering Track

**Purpose**: Demonstrate local development and debugging workflows

**Stack**:
- **Container Runtime**: Docker Compose for service orchestration
- **Edge Proxy**: Traefik for routing and TLS termination
- **Development Tools**: Hot reload, debugging support
- **Simplicity**: Minimal infrastructure overhead

**Use Cases**:
- Local development environment
- Feature development and testing
- Debugging and troubleshooting
- Integration testing

## Architecture Diagrams

### High-Level System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Gateway/Ingress                      │
│              (Envoy Gateway / Traefik)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Fider Application                       │
│                                                               │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │   Frontend   │◄────────┤   Backend    │                  │
│  │ React + TS   │         │   Go + CQRS  │                  │
│  └──────────────┘         └───────┬──────┘                  │
│                                    │                          │
└────────────────────────────────────┼──────────────────────────┘
                                     │
                                     ▼
                          ┌──────────────────┐
                          │   PostgreSQL 17   │
                          └──────────────────┘
```

### Request Flow

```
User → Gateway → Middleware Chain → Handler → Service → Database
                      │
                      ├─ Session
                      ├─ Tenant
                      ├─ Authentication
                      ├─ CSRF
                      └─ Authorization
```

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Backend Language | Go | 1.24 | Application logic |
| Frontend Framework | React | 18 | UI components |
| Frontend Language | TypeScript | Latest | Type safety |
| Database | PostgreSQL | 17 | Data persistence |
| Container Runtime | Docker | Latest | Containerization |
| Orchestration (SRE) | Kubernetes (kind) | Latest | Container orchestration |
| Orchestration (Eng) | Docker Compose | Latest | Local development |
| Gateway (SRE) | Envoy Gateway | Latest | Ingress/routing |
| Gateway (Eng) | Traefik | Latest | Local routing |
| I18n | LinguiJS | Latest | Internationalization |

## Additional Resources

- **Demo Guide**: See `docs/DEMO_GUIDE.md` for running the demo
- **Track Runbooks**: See `docs/runbooks/` for track-specific guides
- **Plan Document**: See `PLAN.md` for detailed implementation plan
