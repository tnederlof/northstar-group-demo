# Fider Demo Base Image
# Built from the modified Fider source with demo mode support
#
# Build from repo root:
#   docker build -f demo/sre/images/base/Dockerfile -t ghcr.io/tnederlof/northstar-group-demo:base .

# ==============================================================================
# Stage 1: Build the Go backend
# ==============================================================================
FROM golang:1.25-bookworm AS go-builder

RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libc6-dev \
    make \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better caching
COPY fider/go.mod fider/go.sum ./
RUN go mod download

# Copy the rest of the source
COPY fider/ ./

# Build the server (CGO required for v8go)
RUN go build -ldflags="-s -w" -o fider .

# ==============================================================================
# Stage 2: Build the frontend
# ==============================================================================
FROM node:24-bookworm AS node-builder

RUN apt-get update && apt-get install -y make && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy package files first for better caching
COPY fider/package.json fider/package-lock.json ./
RUN npm ci

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
