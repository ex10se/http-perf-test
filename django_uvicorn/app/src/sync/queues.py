DJANGO_UVICORN_EXCHANGE = 'django-uvicorn'

QUEUE_DJANGO_UVICORN = 'django-uvicorn'
QUEUE_SYSTEM_DJANGO_UVICORN = 'system-django-uvicorn'

QUEUES_DECLARATION = (
    {
        'exchange': DJANGO_UVICORN_EXCHANGE,
        'queues': [QUEUE_DJANGO_UVICORN, QUEUE_SYSTEM_DJANGO_UVICORN],
    },
)
