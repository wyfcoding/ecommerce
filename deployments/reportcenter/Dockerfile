FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o report-service ./cmd/reportcenter/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/report-service .
COPY configs/reportcenter/config.toml ./configs/reportcenter/config.toml
EXPOSE 8012 9012
CMD ["./report-service"]
