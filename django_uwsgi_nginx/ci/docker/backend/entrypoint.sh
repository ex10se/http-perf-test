#!/bin/sh
set -e

dockerize -wait tcp://rabbitmq:5672 -timeout 60s

python manage.py queue_declare

exec uwsgi --ini /etc/uwsgi/uwsgi.ini
