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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/data_ingestion-service cmd/data_ingestion/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/data_ingestion-service .
COPY --from=builder /app/configs/data_ingestion/config.toml ./configs/data_ingestion/config.toml

# Expose ports
EXPOSE 8122 9142

# Run the application
CMD ["./data_ingestion-service"]
