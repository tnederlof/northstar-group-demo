# Shared Contract

This directory contains the **single source of truth** for configuration shared between the SRE and Engineering demo tracks.

## Purpose

All demo scenarios must use these files to ensure consistency. This prevents configuration drift between tracks and makes updates simpler.

## Files

### `versions.env`
Pinned versions for infrastructure dependencies:
- `POSTGRES_VERSION` - PostgreSQL image version
- `TRAEFIK_VERSION` - Traefik proxy version  
- `MINIO_VERSION` - MinIO version (CI tests only)

### `fider.env.example`
Canonical environment variable contract defining all required and optional variables for Fider demo scenarios. This file is:
- **NOT copied** per scenario
- **NOT edited** directly
- Used as **input** to `render-env.sh` to generate runtime env files

## Invariants

These rules must be maintained across all scenarios:

1. **Both tracks consume these files** - SRE (Kubernetes) and Engineering (Compose) tracks must source from here
2. **No track-specific copies** - Never duplicate these files in track directories
3. **Consistent DATABASE_URL** - All scenarios use: `postgres://fider:fider@postgres:5432/fider?sslmode=disable`
4. **SQL blob storage only** - Demo scenarios use `BLOB_STORAGE=sql`, not S3/MinIO

## Usage

Runtime environment files are generated dynamically:

```bash
# Generate env file for a scenario
./shared/scripts/render-env.sh engineering backend/ui-regression
# Output: demo/.state/engineering/ui-regression/runtime.env
```

### Never Do This
- ❌ Copy `fider.env.example` to a scenario directory
- ❌ Hand-edit generated files in `demo/.state/`
- ❌ Create scenario-specific versions of these files

### Always Do This
- ✅ Use `render-env.sh` to generate runtime env files
- ✅ Update this contract when adding new required variables
- ✅ Keep both tracks in sync by sourcing from here

## Modifying the Contract

When adding new environment variables:

1. Add to `fider.env.example` with appropriate documentation
2. Update `render-env.sh` if special handling is needed
3. Test both SRE and Engineering tracks
4. Update this README if adding new invariants
