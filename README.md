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

## Реализации

### Django + uWSGI + Nginx

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- uWSGI 2.0.30
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- Процессы: 5
- Потоки на процесс: 2
- Всего worker'ов: 10
- Harakiri: 30s
- Buffer size: 32KB
- Max requests: 5000

**Запуск тестов:**
```bash
(cd django_uwsgi_nginx/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1250
- **Success rate:** 99.8%
- **Latency (mean):** 17.17ms
- **Latency (p95):** 70.376ms

### Django + Gunicorn + Nginx

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- Gunicorn 21.2.0
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- Процессы: 5
- Потоки на процесс: 2
- Всего worker'ов: 10
- Harakiri: 30s
- Buffer size: 32KB
- Max requests: 5000

**Запуск тестов:**
```bash
(cd django_gunicorn_nginx/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1062
- **Success rate:** 100%
- **Latency (mean):** 132.589ms
- **Latency (p95):** 634.141ms

### Django + Hypercorn + Nginx

**Стек технологий:**
- Python 3.8
- Django 4.1 + adrf 0.1.12
- Hypercorn 0.17.3
- Aio-pika 9.5.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Запуск тестов:**
```bash
(cd django_hypercorn_nginx/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1681
- **Success rate:** 100%
- **Latency (mean):** 14.696ms
- **Latency (p95):** 32.11ms

### Django + Hypercorn + Nginx + http/2

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- uWSGI 2.0.30
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- 

**Запуск тестов:**
```bash
(cd django_hypercorn_nginx_http2/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Django + Uvicorn + Nginx

**Стек технологий:**
- Python 3.8
- Django 3.2.17 + Django REST Framework 3.13.1
- 
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- 

**Запуск тестов:**
```bash
(cd django_uvicorn_nginx/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### FastAPI + Uvicorn + Nginx

**Стек технологий:**
- Python 3.8
-
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- 

**Запуск тестов:**
```bash
(cd fastapi_uvicorn_nginx/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Запуск тестов:**
```bash
(cd golang/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~8067
- **Success rate:** 100%
- **Latency (mean):** 100.654ms
- **Latency (p95):** 371.543ms

### Golang + http/2

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Запуск тестов:**
```bash
(cd golang_http2/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + fasthttp

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Запуск тестов:**
```bash
(cd golang_fasthttp/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + Gin

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Запуск тестов:**
```bash
(cd golang_gin/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + Echo

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Запуск тестов:**
```bash
(cd golang_echo/load_test && ./benchmark.sh)
```

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Success rate:** *%
- **Latency (mean):** *ms
- **Latency (p95):** *ms
