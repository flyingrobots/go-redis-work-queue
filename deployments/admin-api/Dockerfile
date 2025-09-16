# Admin API Dockerfile
FROM golang:1.25-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the admin API
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin-api ./cmd/admin-api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/admin-api .

# Copy config files
COPY --from=builder /app/configs/admin-api.yaml ./configs/

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the admin API
CMD ["./admin-api"]