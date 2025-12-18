# HTTP Performance Testing Project

Проект для сравнительного нагрузочного тестирования нескольких HTTP-фреймворков и стеков.

## Цель проекта

Провести сравнительный анализ производительности различных стеков технологий для обработки HTTP-запросов с последующей отправкой сообщений в RabbitMQ.

## Тестируемое приложение

Каждое приложение реализует один эндпойнт:
- **Метод:** POST
- **Путь:** `/status/status/`
- **Формат данных:** JSON-массив событий
- **Логика:** Валидация входных данных, маршрутизация в RabbitMQ очереди на основе поля `is_system`

## Пример запроса

```json
[
  {
    "state": "delivered",
    "updatedAt": "2024-01-01T00:00:00Z",
    "txId": "test123",
    "trackData": {
      "is_system": false
    }
  }
]
```

## Метод тестирования

- Инструмент: Vegeta (**требует локальной установки**)
- Длительность: 30 секунд на тест
- Алгоритм: Бинарный поиск максимального RPS
- Критерий успеха: Success rate ≥ 99.5%

## Запуск тестов
- Простой запуск
    ```bash
    (cd django_gunicorn/load_test && ./benchmark.sh)
    ```
- Запуск с ограничением ресурсов
    ```bash
    systemd-run --user --scope -p MemoryLimit=6G -p CPUQuota=80% bash -lc "cd $PWD/django_gunicorn/load_test  && ./benchmark.sh"
    ```
- Холодный запуск (в тестах использовался этот)
    ```bash
    export S=<service>
    find . -path './*/ci/docker/docker-compose.yml' -exec zsh -c 'docker-compose -f {} down -v' \; && docker-compose -f rabbitmq/ci/docker/docker-compose.yml up -d --build && docker-compose -f $S/ci/docker/docker-compose.yml up -d --build && sleep 2 && systemd-run --user --scope -p MemoryLimit=6G -p CPUQuota=80% bash -lc "cd $PWD/$S/load_test  && ./benchmark.sh"
    
    # export S=django_uwsgi
    # export S=django_gunicorn
    # export S=django_hypercorn
    # export S=django_hypercorn_http2
    # export S=django_uvicorn
    # export S=fastapi_uvicorn
    # export S=go
    # export S=go_http2
    # export S=go_fasthttp
    # export S=go_gin
    # export S=go_echo
    # export S=rust_actix
    # export S=rust_axum
    # export S=rust_hyper
    ```

## Реализации

Везде используется Nginx 1.25-alpine

### Django + uWSGI

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- uWSGI 2.0.30
- Pika 1.3.2

**Конфигурация сервера:**
- Процессы: 5
- Потоки на процесс: 2
- Всего worker'ов: 10
- Harakiri: 30s
- Buffer size: 32KB
- Max requests: 5000

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1183
- **Latency (mean):** 6ms
- **Latency (p95):** 29ms

### Django + Gunicorn

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- Gunicorn 21.2.0
- Pika 1.3.2

**Конфигурация сервера:**
- Процессы: 5
- Потоки на процесс: 2
- Всего worker'ов: 10
- Harakiri: 30s
- Buffer size: 32KB
- Max requests: 5000

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1221
- **Latency (mean):** 444s
- **Latency (p95):** 1362s

### Django + Hypercorn

**Стек технологий:**
- Python 3.8
- Django 4.1 + adrf 0.1.12
- Hypercorn 0.17.3
- Aio-pika 9.5.2

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1677
- **Latency (mean):** 69ms
- **Latency (p95):** 485ms

### Django + Hypercorn + http/2

**Стек технологий:**
- Python 3.8
- Django 4.1 + adrf 0.1.12
- Hypercorn 0.17.3
- Aio-pika 9.5.2

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1961
- **Latency (mean):** 62ms
- **Latency (p95):** 203ms

### Django + Uvicorn

**Стек технологий:**
- Python 3.8
- Django 4.1 + adrf 0.1.12
- Uvicorn 0.33.0
- Aio-pika 9.5.2

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~2250
- **Latency (mean):** 16ms
- **Latency (p95):** 39ms

### FastAPI + Uvicorn

**Стек технологий:**
- Python 3.8
- FastAPI 0.123.0
- Uvicorn 0.33.0
- Aio-pika 9.5.2

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~11377
- **Latency (mean):** 61ms
- **Latency (p95):** 149ms

### Go

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 10
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~7593
- **Latency (mean):** 11ms
- **Latency (p95):** 46ms

### Go + http/2

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 10
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~7578
- **Latency (mean):** 61ms
- **Latency (p95):** 230ms

### Go + fasthttp

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0
- github.com/valyala/fasthttp v1.68.0

**Конфигурация сервера:**
- Процессы: 10
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~7584
- **Latency (mean):** 27ms
- **Latency (p95):** 105ms

### Go + Gin

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0
- github.com/gin-gonic/gin v1.11.0

**Конфигурация сервера:**
- Процессы: 10
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~6327
- **Latency (mean):** 41ms
- **Latency (p95):** 235ms

### Go + Echo

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0
- github.com/labstack/echo/v4 v4.14.0

**Конфигурация сервера:**
- Процессы: 10
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~7581
- **Latency (mean):** 32ms
- **Latency (p95):** 856ms

### Rust + Actix-Web

**Стек технологий:**
- Rust (latest stable)
- actix-web 4
- lapin 2

**Конфигурация сервера:**
- Workers: 10

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~4000
- **Latency (mean):** 4ms
- **Latency (p95):** 11ms

### Rust + Axum

**Стек технологий:**
- Rust (latest stable)
- axum 0.7
- lapin 2

**Конфигурация сервера:**
- Workers: 10

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~5373
- **Latency (mean):** 4ms
- **Latency (p95):** 12ms

### Rust + Hyper

**Стек технологий:**
- Rust (latest stable)
- hyper 1
- lapin 2

**Конфигурация сервера:**
- Workers: 10

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~6281
- **Latency (mean):** 3ms
- **Latency (p95):** 11ms
