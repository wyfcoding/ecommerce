FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o returns-service ./cmd/returns/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/returns-service .
COPY configs/returns/config.toml ./configs/returns/config.toml
EXPOSE 8014 9014
CMD ["./returns-service"]
