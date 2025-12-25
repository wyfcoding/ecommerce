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
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/pointsmall cmd/pointsmall/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/pointsmall .
COPY --from=builder /app/configs/pointsmall/config.toml ./configs/pointsmall/config.toml

# Expose ports
EXPOSE 8106 9126

# Run the application
CMD ["./pointsmall"]
