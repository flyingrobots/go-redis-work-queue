# Admin API Dockerfile
FROM golang:1.25-alpine AS builder

ARG VERSION=dev

# Install dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the admin API
RUN CGO_ENABLED=0 GOOS=linux \
	go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /app/bin/admin-api ./cmd/admin-api

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

# Create dedicated user
RUN addgroup -g 1000 -S adminapi && \
	adduser -u 1000 -S adminapi -G adminapi

WORKDIR /app

# Copy the binary
COPY --from=builder /app/bin/admin-api /app/admin-api

# Ensure working directories exist
RUN mkdir -p /app/logs /app/configs && \
	chown -R adminapi:adminapi /app

# Configs must be mounted at runtime (ConfigMap/Secret)

# Expose port
EXPOSE 8080

# Drop privileges
USER adminapi

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the admin API
CMD ["/app/admin-api"]
