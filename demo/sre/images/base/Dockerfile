# Fider Demo Base Image
# Built from the modified Fider source with demo mode support
#
# Build from repo root:
#   docker build -f demo/sre/images/base/Dockerfile -t ghcr.io/tnederlof/northstar-group-demo:base .

# ==============================================================================
# Stage 1: Build the Go backend
# ==============================================================================
FROM golang:1.25-bookworm AS go-builder

# Most packages already in golang image, only add what's missing
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better caching
COPY fider/go.mod fider/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY fider/ ./

# Build the server with cache mounts for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o fider .

# ==============================================================================
# Stage 2: Build the frontend
# ==============================================================================
FROM node:24-bookworm AS node-builder

# make is already in node image
WORKDIR /app

# Copy package files first for better caching
COPY fider/package.json fider/package-lock.json ./

# Use cache mount for npm to avoid re-downloading packages
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit

# Copy source and build
COPY fider/ ./

# Combine build steps to reduce layers and enable parallel processing
RUN make build-ssr && make build-ui

# ==============================================================================
# Stage 3: Final runtime image
# ==============================================================================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy built artifacts
COPY --from=go-builder /app/fider ./fider
COPY --from=node-builder /app/dist ./dist
COPY --from=node-builder /app/ssr.js ./ssr.js
COPY --from=node-builder /app/migrations ./migrations
COPY --from=node-builder /app/views ./views
COPY --from=node-builder /app/locale ./locale

# Copy static files
COPY fider/favicon.png ./favicon.png
COPY fider/robots.txt ./robots.txt

# Create etc directory for legal docs (optional)
RUN mkdir -p /app/etc

EXPOSE 3000

ENTRYPOINT ["./fider"]
