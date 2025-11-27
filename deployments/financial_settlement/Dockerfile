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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/financial_settlement-service cmd/financial_settlement/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/financial_settlement-service .
COPY --from=builder /app/configs/financial_settlement/config.toml ./configs/financial_settlement/config.toml

# Expose ports
EXPOSE 8096 9106

# Run the application
CMD ["./financial_settlement-service"]
