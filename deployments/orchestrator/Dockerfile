# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o orchestrator ecommerce/cmd/orchestrator/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/orchestrator .
COPY --from=builder /app/ecommerce/configs/orchestrator/config.toml ./configs/

EXPOSE 50062

CMD ["./orchestrator"]
