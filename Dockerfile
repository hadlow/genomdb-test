# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o genomdb .

# Development stage (hot reload)
FROM golang:1.22-alpine AS dev

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk --no-cache add ca-certificates wget git

# Cache modules first
COPY go.mod go.sum ./
RUN go mod download

# Install Air for live reload
RUN go install github.com/air-verse/air@latest

# Copy source for initial startup (will be bind-mounted in dev)
COPY . .

# Default dev command (can be overridden in compose)
CMD ["air"]

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/genomdb .

# Copy config files (both regular and docker variants)
COPY configs/ ./configs/

# Copy static frontend assets
COPY frontend/ ./frontend/

# Create data directory
RUN mkdir -p /app/data

# Expose HTTP, Raft, and frontend ports
EXPOSE 8001 8002 8003 9001 9002 9003 8080

# Default command (will be overridden by docker-compose)
CMD ["./genomdb", "start", "configs/config-node1.yml"]
