# Stage 1: Сборка
FROM golang:1.25-alpine AS builder
WORKDIR /build
# Устанавливаем зависимости для сборки
RUN apk add --no-cache git
# Копируем go.mod и go.sum для кеширования зависимостей
COPY app/src/go.mod app/src/go.sum ./
RUN go mod download
# Копируем исходный код
COPY app/src/ ./
# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
# Устанавливаем CA сертификаты для HTTPS и wget для dockerize
RUN apk --no-cache add ca-certificates wget
# Скачиваем dockerize для ожидания RabbitMQ
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.6.1.tar.gz && \
  rm dockerize-linux-amd64-v0.6.1.tar.gz && \
  chmod +x /usr/local/bin/dockerize
# Копируем бинарник из builder stage
COPY --from=builder /build/app .
# Копируем entrypoint
COPY ci/docker/backend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
# Создаем директорию для socket
RUN mkdir -p /tmp/go
# Создаем non-root пользователя
RUN adduser -D -u 1000 goapp && \
  chown -R goapp:goapp /app && \
  chown -R goapp:goapp /tmp/go
USER goapp
ENTRYPOINT ["/entrypoint.sh"]
