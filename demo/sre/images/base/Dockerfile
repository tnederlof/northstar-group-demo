# Fider Demo Base Image
# Built from the modified Fider source with demo mode support
#
# Build from repo root:
#   docker build -f demo/sre/images/base/Dockerfile -t ghcr.io/warpdotdev/northstar-group-demo-fider:base .

# ==============================================================================
# Stage 1: Build the Go backend
# ==============================================================================
FROM golang:1.24-alpine AS go-builder

RUN apk add --no-cache make git

WORKDIR /app

# Copy go mod files first for better caching
COPY fider/go.mod fider/go.sum ./
RUN go mod download

# Copy the rest of the source
COPY fider/ ./

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o fider .

# ==============================================================================
# Stage 2: Build the frontend
# ==============================================================================
FROM node:22-alpine AS node-builder

WORKDIR /app

# Copy package files first for better caching
COPY fider/package.json fider/package-lock.json ./
RUN npm ci

# Copy source and build
COPY fider/ ./
RUN npm run build

# ==============================================================================
# Stage 3: Final runtime image
# ==============================================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy built artifacts
COPY --from=go-builder /app/fider ./fider
COPY --from=node-builder /app/dist ./dist
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
