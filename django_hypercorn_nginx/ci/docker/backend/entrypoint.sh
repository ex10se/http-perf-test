#!/bin/sh
set -e

dockerize -wait tcp://rabbitmq:5672 -timeout 60s

python manage.py queue_declare

umask 0000

exec hypercorn app.asgi:application \
  --bind unix:/tmp/hypercorn/app.sock \
  --workers 10 \
  --worker-class asyncio \
  --backlog 2048 \
  --keep-alive 5
