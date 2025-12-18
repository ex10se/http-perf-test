DJANGO_UWSGI_EXCHANGE = 'django-uwsgi'

QUEUE_DJANGO_UWSGI = 'django-uwsgi'
QUEUE_SYSTEM_DJANGO_UWSGI = 'system-django-uwsgi'

QUEUES_DECLARATION = (
    {
        'exchange': DJANGO_UWSGI_EXCHANGE,
        'queues': [QUEUE_DJANGO_UWSGI, QUEUE_SYSTEM_DJANGO_UWSGI],
    },
)
