# Unified Dockerfile for E-commerce Microservices
# Usage: docker build --build-arg SERVICE_NAME=user -t user-service .

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

# SERVICE_NAME build argument
ARG SERVICE_NAME
RUN if [ -z "$SERVICE_NAME" ]; then echo "Error: SERVICE_NAME build-arg is required."; exit 1; fi

# Build the application
# We name the binary 'server' to simplify the CMD in the final stage
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server cmd/${SERVICE_NAME}/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Copy configuration file
# We use the SERVICE_NAME arg again to find the correct config
ARG SERVICE_NAME
COPY --from=builder /app/configs/${SERVICE_NAME}/config.toml ./configs/${SERVICE_NAME}/config.toml

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Run the application
CMD ["./server"]