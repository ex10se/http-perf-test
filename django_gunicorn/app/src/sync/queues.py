DJANGO_GUNICORN_EXCHANGE = 'django-gunicorn'

QUEUE_DJANGO_GUNICORN = 'django-gunicorn'
QUEUE_SYSTEM_DJANGO_GUNICORN = 'system-django-gunicorn'

QUEUES_DECLARATION = (
    {
        'exchange': DJANGO_GUNICORN_EXCHANGE,
        'queues': [QUEUE_DJANGO_GUNICORN, QUEUE_SYSTEM_DJANGO_GUNICORN],
    },
)
