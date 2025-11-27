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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/inventory_forecast-service cmd/inventory_forecast/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/inventory_forecast-service .
COPY --from=builder /app/configs/inventory_forecast/config.toml ./configs/inventory_forecast/config.toml

# Expose ports
EXPOSE 8099 9109

# Run the application
CMD ["./inventory_forecast-service"]
