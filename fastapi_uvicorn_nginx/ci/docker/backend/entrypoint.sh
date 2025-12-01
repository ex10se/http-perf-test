#!/bin/sh
set -e

dockerize -wait tcp://rabbitmq:5672 -timeout 60s

python -m sync.queue_declare

umask 0000

exec uvicorn app.main:app \
  --uds /tmp/uvicorn/app.sock \
  --workers 10 \
  --backlog 2048 \
  --timeout-keep-alive 5
