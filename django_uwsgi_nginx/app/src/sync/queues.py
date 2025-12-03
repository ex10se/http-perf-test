DJANGO_UWSGI_NGINX_EXCHANGE = 'django-uwsgi-nginx'

QUEUE_DJANGO_UWSGI_NGINX = 'django-uwsgi-nginx'
QUEUE_SYSTEM_DJANGO_UWSGI_NGINX = 'system-django-uwsgi-nginx'

QUEUES_DECLARATION = (
    {
        'exchange': DJANGO_UWSGI_NGINX_EXCHANGE,
        'queues': [QUEUE_DJANGO_UWSGI_NGINX, QUEUE_SYSTEM_DJANGO_UWSGI_NGINX],
    },
)
