FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy-service ./cmd/proxy-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy binary from builder
COPY --from=builder /app/proxy-service .
COPY --from=builder /app/config/config.production.yaml ./config/config.yaml

# Create necessary directories
RUN mkdir -p /app/cert

# Expose ports for the application and metrics
EXPOSE 8080 9090

CMD ["./proxy-service"]