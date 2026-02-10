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

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/genomdb .

# Copy config files (both regular and docker variants)
COPY configs/ ./configs/

# Create data directory
RUN mkdir -p /app/data

# Expose HTTP and Raft ports
EXPOSE 8001 8002 8003 9001 9002 9003

# Default command (will be overridden by docker-compose)
CMD ["./genomdb", "start", "configs/config-node1.yml"]
