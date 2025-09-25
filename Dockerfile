# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Set build environment variables
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

# Copy dependency management files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# SERVICE_NAME will be passed in at build time (e.g., --build-arg SERVICE_NAME=user)
ARG SERVICE_NAME
# Build the specific service binary
RUN go build -a -installsuffix cgo -o /app/bin/${SERVICE_NAME} ./cmd/${SERVICE_NAME}

# Stage 2: Create the final lightweight production image
FROM alpine:latest

WORKDIR /app

# Argument for service name, to be provided during build
ARG SERVICE_NAME

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bin/${SERVICE_NAME} .

# Copy the configuration file for the specific service
COPY configs/${SERVICE_NAME}.toml ./configs/

# Expose the HTTP and gRPC ports defined in the user.toml config.
# This is for documentation purposes; actual port mapping is done elsewhere.
EXPOSE 8080
EXPOSE 9090

# Define the command to run the service, pointing to the correct config file path
CMD ["./${SERVICE_NAME}", "-conf", "./configs/${SERVICE_NAME}.toml"]
