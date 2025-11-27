# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/customer_service-service cmd/customer_service/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/customer_service-service .
COPY --from=builder /app/configs/customer_service/config.toml ./configs/customer_service/config.toml

# Expose ports
EXPOSE 8080 9090

# Run the application
CMD ["./customer_service-service"]
