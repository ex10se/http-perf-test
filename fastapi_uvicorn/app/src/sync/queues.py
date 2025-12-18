FASTAPI_UVICORN_EXCHANGE = 'fastapi-uvicorn'

QUEUE_FASTAPI_UVICORN = 'fastapi-uvicorn'
QUEUE_SYSTEM_FASTAPI_UVICORN = 'system-fastapi-uvicorn'

QUEUES_DECLARATION = (
    {
        'exchange': FASTAPI_UVICORN_EXCHANGE,
        'queues': [QUEUE_FASTAPI_UVICORN, QUEUE_SYSTEM_FASTAPI_UVICORN_NGINX],
    },
)
