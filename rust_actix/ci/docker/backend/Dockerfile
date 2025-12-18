# Stage 1: Сборка
FROM rust:alpine AS builder

WORKDIR /build

# Устанавливаем зависимости для сборки
RUN apk add --no-cache musl-dev

# Копируем манифест и создаем dummy проект для кеширования зависимостей
COPY app/build/cargo/Cargo.toml ./
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Копируем исходный код
COPY app/src ./src

# Собираем приложение
RUN touch src/main.rs && cargo build --release

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Устанавливаем CA сертификаты и wget для dockerize
RUN apk --no-cache add ca-certificates wget

# Скачиваем dockerize для ожидания RabbitMQ
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
    tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.6.1.tar.gz && \
    rm dockerize-linux-amd64-v0.6.1.tar.gz && \
    chmod +x /usr/local/bin/dockerize

# Копируем бинарник из builder stage
COPY --from=builder /build/target/release/rust-actix ./app

# Копируем entrypoint
COPY ci/docker/backend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Создаем директорию для socket
RUN mkdir -p /tmp/rust_actix

# Создаем non-root пользователя
RUN adduser -D -u 1000 rustapp && \
    chown -R rustapp:rustapp /app && \
    chown -R rustapp:rustapp /tmp/rust_actix

USER rustapp

ENTRYPOINT ["/entrypoint.sh"]
