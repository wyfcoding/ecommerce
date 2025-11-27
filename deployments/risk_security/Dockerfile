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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/risk_security-service cmd/risk_security/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/risk_security-service .
COPY --from=builder /app/configs/risk_security/config.toml ./configs/risk_security/config.toml

# Expose ports
EXPOSE 8110 9130

# Run the application
CMD ["./risk_security-service"]
