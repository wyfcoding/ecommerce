FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o abtest-service ./cmd/abtest/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/abtest-service .
COPY configs/abtest/config.toml ./configs/abtest/config.toml
EXPOSE 8013 9013
CMD ["./abtest-service"]
创新: 保持了镜像的轻量化，同时包含必要的运行时配置。
