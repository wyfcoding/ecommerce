# Use the official Golang image as the base image
FROM golang:1.20-alpine

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./go.mod
COPY go.sum ./go.sum

# Download all dependencies. Dependencies will be cached if go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
# CGO_ENABLED=0 is important for static linking, making the binary self-contained
# -o product-service specifies the output binary name
# ./cmd/product specifies the main package to build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./cmd/product

# Expose the port that the application listens on
EXPOSE 8000

# Run the application
CMD ["./product-service"]
