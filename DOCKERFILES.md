# Dockerfile Overview

This repository contains multiple types of Dockerfiles for different purposes:

## Development Dockerfiles (`.dev`)

**Purpose**: Used by Warp Agent Mode for development environments on the Warp platform.

**Files**:
- `fider/Dockerfile.dev` - Development environment for Fider application code
- `demo/Dockerfile.dev` - Development environment for demo infrastructure work

**Key Characteristics**:
- Include both Go and Node.js toolchains
- Include development tools (kubectl, kind, helm for demo work)
- Designed for interactive development and debugging
- Keep containers running with `/bin/bash` as default command
- Match local development versions (Go 1.25, Node.js 24)

**When to Use**:
- Working with Warp Agent Mode on the Warp platform
- Need full development toolchain for building and testing
- Interactive debugging sessions
- Infrastructure development (kind clusters, kubectl, etc.)

## Production Dockerfiles

**Purpose**: Used for running the Fider application in demo scenarios (both SRE and Engineering tracks).

**Files**:
- `fider/Dockerfile` - Production-optimized multi-stage build for Fider

**Key Characteristics**:
- Multi-stage build (smaller final image)
- Only includes runtime dependencies
- Optimized for fast startup and low resource usage
- Used in both Kubernetes (SRE) and Docker Compose (Engineering) deployments

**When to Use**:
- Running demo scenarios (`make run SCENARIO=...`)
- Production deployments
- CI/CD pipelines
- Scenario testing and verification

## Key Differences

| Feature | Development (`.dev`) | Production |
|---------|---------------------|------------|
| **Size** | Larger (~1GB+) | Smaller (~200MB) |
| **Build Time** | Fast (single stage) | Slower (multi-stage) |
| **Tools** | Full dev toolchain | Runtime only |
| **Use Case** | Development/debugging | Running scenarios |
| **Environment** | Warp platform | Kubernetes/Docker Compose |
| **Default CMD** | `/bin/bash` | Application entrypoint |

## Example Usage

### Development Dockerfile (Warp Agent Mode)
```bash
# Used automatically by Warp when running agent sessions
# No manual docker build needed
```

### Production Dockerfile (Demo Scenarios)
```bash
# Used automatically by demo scripts
make run SCENARIO=backend/ui-regression
# This builds fider/Dockerfile and runs containers
```

## Maintenance Notes

- Development Dockerfiles should match local development tool versions
- Production Dockerfiles should be optimized for size and startup time
- Both should use the same language/runtime versions to ensure consistency
- Update both when upgrading Go or Node.js versions
