# Fider Demo Base Image
# Built from the modified Fider source with demo mode support
#
# Build from repo root:
#   docker build -f demo/sre/images/base/Dockerfile -t ghcr.io/tnederlof/northstar-group-demo:base .
#
# Based on fider/Dockerfile with adjustments for demo use

# ==============================================================================
# Stage 1: Build the Go backend
# ==============================================================================
FROM golang:1.25-bookworm AS server-builder

# Install all build dependencies (matching original Fider Dockerfile)
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /server

# Copy go mod files first for better caching
COPY fider/go.mod fider/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY fider/ ./

# Build the server using make target (matches original)
ARG COMMITHASH=demo
ARG VERSION=demo
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    COMMITHASH=${COMMITHASH} VERSION=${VERSION} make build-server

# ==============================================================================
# Stage 2: Build the frontend
# ==============================================================================
FROM node:24-bookworm AS ui-builder

WORKDIR /ui

# Copy package files first for better caching
COPY fider/package.json fider/package-lock.json ./

# Use cache mount for npm (matching original's maxsockets approach)
RUN --mount=type=cache,target=/root/.npm \
    npm ci --maxsockets 1

# Copy source and build
COPY fider/ ./
RUN make build-ssr
RUN make build-ui

# ==============================================================================
# Stage 3: Final runtime image
# ==============================================================================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy from server-builder (matching original Fider Dockerfile)
COPY --from=server-builder /server/migrations /app/migrations
COPY --from=server-builder /server/views /app/views
COPY --from=server-builder /server/locale /app/locale
COPY --from=server-builder /server/LICENSE /app
COPY --from=server-builder /server/fider /app
COPY --from=server-builder /server/static /app/static

# Copy from ui-builder (matching original Fider Dockerfile)
COPY --from=ui-builder /ui/favicon.png /app
COPY --from=ui-builder /ui/dist /app/dist
COPY --from=ui-builder /ui/robots.txt /app
COPY --from=ui-builder /ui/ssr.js /app

# Create etc directory for legal docs (optional)
RUN mkdir -p /app/etc

EXPOSE 3000

HEALTHCHECK --timeout=5s CMD ./fider ping

# Note: We use ENTRYPOINT instead of CMD because migrations are run
# separately via kubectl run in the demo harness (demo.sh)
ENTRYPOINT ["./fider"]
