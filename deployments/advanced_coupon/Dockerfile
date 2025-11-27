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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/advanced_coupon-service cmd/advanced_coupon/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/advanced_coupon-service .
COPY --from=builder /app/configs/advanced_coupon/config.toml ./configs/advanced_coupon/config.toml

# Expose ports
EXPOSE 8124 9144

# Run the application
CMD ["./advanced_coupon-service"]
