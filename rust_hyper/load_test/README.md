# Load Testing for Rust Hyper + Nginx

## Запуск тестирования

```bash
cd rust_hyper/load_test
./benchmark.sh
```

## Требования

- Vegeta load testing tool
- Docker и docker-compose
- Запущенный RabbitMQ и сервис rust_hyper

## Описание

Скрипт использует бинарный поиск для определения максимального стабильного RPS
при котором система поддерживает success rate >= 99.5%.
