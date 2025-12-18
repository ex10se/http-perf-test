# Load Testing for Rust Axum + Nginx

## Запуск тестирования

```bash
cd rust_axum/load_test
./benchmark.sh
```

## Требования

- Vegeta load testing tool
- Docker и docker-compose
- Запущенный RabbitMQ и сервис rust_axum

## Описание

Скрипт использует бинарный поиск для определения максимального стабильного RPS
при котором система поддерживает success rate >= 99.5%.
