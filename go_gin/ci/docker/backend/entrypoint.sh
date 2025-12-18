#!/bin/sh
set -e

dockerize -wait tcp://rabbitmq:5672 -timeout 60s

exec ./app
