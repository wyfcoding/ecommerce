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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/groupbuy-service cmd/groupbuy/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/groupbuy-service .
COPY --from=builder /app/configs/groupbuy/config.toml ./configs/groupbuy/config.toml

# Expose ports
EXPOSE 8098 9108

# Run the application
CMD ["./groupbuy-service"]
