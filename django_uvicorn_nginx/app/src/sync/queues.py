DJANGO_UVICORN_NGINX_EXCHANGE = 'django-uvicorn-nginx'

QUEUE_DJANGO_UVICORN_NGINX = 'django-uvicorn-nginx'
QUEUE_SYSTEM_DJANGO_UVICORN_NGINX = 'system-django-uvicorn-nginx'

QUEUES_DECLARATION = (
    {
        'exchange': DJANGO_UVICORN_NGINX_EXCHANGE,
        'queues': [QUEUE_DJANGO_UVICORN_NGINX, QUEUE_SYSTEM_DJANGO_UVICORN_NGINX],
    },
)
