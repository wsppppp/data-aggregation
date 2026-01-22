# Build stage
FROM golang:1.25.1-alpine AS builder
WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/app

# Run stage
FROM alpine:3.19
WORKDIR /app
ENV GIN_MODE=release

# Копируем бинарник и Swagger/OpenAPI YAML (нужен для UI)
COPY --from=builder /app/app /app/app
COPY docs /app/docs

EXPOSE 8080
CMD ["/app/app"]