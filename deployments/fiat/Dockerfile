# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o fiat ecommerce/cmd/fiat/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/fiat .
COPY --from=builder /app/ecommerce/configs/fiat/config.toml ./configs/

EXPOSE 50061

CMD ["./fiat"]
