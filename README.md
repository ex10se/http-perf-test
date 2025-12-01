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
    (cd django_gunicorn_nginx/load_test && ./benchmark.sh)
    ```
- Запуск с ограничением ресурсов
    ```bash
    systemd-run --user --scope -p MemoryLimit=6G -p CPUQuota=80% bash -lc "cd $PWD/django_gunicorn_nginx/load_test  && ./benchmark.sh"
    ```
- Холодный запуск (в тестах использовался этот)
    ```bash
    export S=<service>
    find . -path './*/ci/docker/docker-compose.yml' -exec zsh -c 'docker-compose -f {} down -v' \; && docker-compose -f rabbitmq/ci/docker/docker-compose.yml up -d --build && docker-compose -f $S/ci/docker/docker-compose.yml up -d --build && sleep 2 && systemd-run --user --scope -p MemoryLimit=6G -p CPUQuota=80% bash -lc "cd $PWD/$S/load_test  && ./benchmark.sh"
    
    # export S=django_uwsgi_nginx
    # export S=django_gunicorn_nginx
    # export S=django_hypercorn_nginx
    # export S=django_hypercorn_nginx_http2
    # export S=django_uvicorn_nginx
    # export S=fastapi_uvicorn_nginx
    # export S=golang_nginx
    # export S=golang_nginx_http2
    # export S=golang_fasthttp_nginx
    # export S=golang_gin_nginx
    # export S=golang_echo_nginx
    ```

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

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1117
- **Latency (mean):** 7ms
- **Latency (p95):** 15ms

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

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1312
- **Latency (mean):** 1200s
- **Latency (p95):** 3100s

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

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1875
- **Latency (mean):** 24ms
- **Latency (p95):** 58ms

### Django + Hypercorn + Nginx + http/2

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

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1968
- **Latency (mean):** 24ms
- **Latency (p95):** 53ms

### Django + Uvicorn + Nginx

**Стек технологий:**
- Python 3.8
- Django 4.1 + adrf 0.1.12
- Uvicorn 0.33.0
- Aio-pika 9.5.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- Workers: 10
- Backlog: 2048
- Keep-alive: 5s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~2250
- **Latency (mean):** 18ms
- **Latency (p95):** 40ms

### FastAPI + Uvicorn + Nginx

**Стек технологий:**
- Python 3.8
-
- Pika 1.3.2
- Nginx 1.25-alpine

**Конфигурация сервера:**
- 

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + Nginx

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~7593
- **Latency (mean):** 52ms
- **Latency (p95):** 794ms

### Golang + Nginx + http/2

**Стек технологий:**
- Go 1.25.1
- github.com/rabbitmq/amqp091-go v1.10.0

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + fasthttp + Nginx

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + Gin + Nginx

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Latency (mean):** *ms
- **Latency (p95):** *ms

### Golang + Echo + Nginx

**Стек технологий:**
- Go 1.25.1
- 

**Конфигурация сервера:**
- Процессы: 5
- ReadTimeout: 30s
- WriteTimeout: 30s
- IdleTimeout: 120s

**Результаты тестирования:**
<TODO>
- **Максимальный стабильный RPS:** ~*
- **Latency (mean):** *ms
- **Latency (p95):** *ms
