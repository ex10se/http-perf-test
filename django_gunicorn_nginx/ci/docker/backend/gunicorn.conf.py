workers = 5
threads = 2
worker_class = 'gthread'
bind = 'unix:/tmp/gunicorn/app.sock'
max_requests = 5000
max_requests_jitter = 100
timeout = 30
