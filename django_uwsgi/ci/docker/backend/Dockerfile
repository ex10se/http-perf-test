FROM python:3.8-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gcc \
    python3-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

COPY app/build/pip/requirements.txt /tmp/requirements.txt

RUN pip install --no-cache-dir -r /tmp/requirements.txt

RUN mkdir -p /etc/uwsgi

COPY ci/docker/backend/uwsgi.ini /etc/uwsgi/uwsgi.ini
COPY ci/docker/backend/entrypoint.sh /entrypoint.sh

COPY app/src /app/

RUN mkdir -p /tmp/uwsgi

RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
    tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.6.1.tar.gz && \
    rm dockerize-linux-amd64-v0.6.1.tar.gz && \
    chmod +x /usr/local/bin/dockerize

RUN chmod +x /entrypoint.sh

RUN useradd -m -u 1000 uwsgi && \
    chown -R uwsgi:uwsgi /app && \
    chown -R uwsgi:uwsgi /tmp/uwsgi && \
    chown -R uwsgi:uwsgi /etc/uwsgi

USER uwsgi

EXPOSE 8000

CMD ["/entrypoint.sh"]
