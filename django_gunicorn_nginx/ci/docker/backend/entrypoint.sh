#!/bin/sh
set -e

dockerize -wait tcp://rabbitmq:5672 -timeout 60s

python manage.py queue_declare

exec gunicorn app.wsgi:application -c /etc/gunicorn/gunicorn.conf.py
