FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o cms-service ./cmd/cms/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cms-service .
COPY configs/cms/config.toml ./configs/cms/config.toml
EXPOSE 8011 9011
CMD ["./cms-service"]
