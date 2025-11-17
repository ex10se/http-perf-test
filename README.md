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

### Пример запроса

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

## Реализации

### 1. Django + uWSGI + Nginx

**Стек технологий:**
- Django 3.2.17 + Django REST Framework 3.13.1
- uWSGI 2.0.30
- Nginx 1.25-alpine
- Python 3.8
- RabbitMQ 4.1.3
- Pika 1.3.2
- vegeta для тестирования (требует локальной установки)

**Конфигурация uWSGI:**
- Процессы: 5
- Потоки на процесс: 2
- Всего worker'ов: 10
- Harakiri: 30s
- Buffer size: 32KB
- Max requests: 5000

**Метод тестирования:**
- Инструмент: Vegeta
- Длительность: 30 секунд на тест
- Алгоритм: Бинарный поиск максимального RPS
- Критерий успеха: Success rate ≥ 99.5%

**Результаты тестирования:**
- **Максимальный стабильный RPS:** ~1046
- **Success rate:** 99.77%
- **Latency (mean):** 11.49ms
- **Latency (p95):** 35.5ms
- **Проблемный RPS:** 1054 (98.94% success)

**Запуск тестов:**
```bash
cd django_uwsgi_nginx/load_test
./benchmark.sh
```

### 2. Django + Gunicorn + Nginx

TODO

### 3. Django + Uvicorn + Nginx

TODO

### 4. FastAPI + Uvicorn + Nginx

TODO

### 5. Golang (net/http)

TODO

### 6. Golang + fasthttp

TODO

### 7. Golang + Gin

TODO

### 8. Golang + Echo

TODO
