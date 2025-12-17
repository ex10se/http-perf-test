import json

from adrf.viewsets import ViewSet
from rest_framework import status
from rest_framework.decorators import action
from rest_framework.parsers import JSONParser
from rest_framework.request import Request
from rest_framework.response import Response
from sync.base import get_rabbit_client
from sync.queues import (
    DJANGO_UVICORN_NGINX_EXCHANGE,
    QUEUE_DJANGO_UVICORN_NGINX,
    QUEUE_SYSTEM_DJANGO_UVICORN_NGINX,
)
from sync.serializers import StatusEventSerializer


class StatusViewSet(ViewSet):
    parser_classes = (JSONParser,)

    @action(methods=['POST'], detail=False, url_path='status')
    async def status(self, request: Request) -> Response:
        """
        POST status/ - принимает JSON-массив событий и отправляет их в rabbitmq очереди
        """
        if not request.data:
            return Response(
                {'error': 'Request body is required'},
                status=status.HTTP_400_BAD_REQUEST,
            )

        if not isinstance(request.data, list):
            return Response(
                {'error': 'Request body must be a JSON array'},
                status=status.HTTP_400_BAD_REQUEST,
            )

        serializer = StatusEventSerializer(data=request.data, many=True)
        if not serializer.is_valid():
            return Response(
                {'error': 'Validation failed', 'details': serializer.errors},
                status=status.HTTP_400_BAD_REQUEST,
            )

        base_sync = await get_rabbit_client()
        errors = []
        serializer_data = await serializer.adata
        for event in serializer_data:
            try:
                track_data = event.get('trackData') or {}
                is_system = track_data.get('is_system', False)
                queue_name = (
                    QUEUE_SYSTEM_DJANGO_UVICORN_NGINX
                    if is_system
                    else QUEUE_DJANGO_UVICORN_NGINX
                )

                # Отправляем событие в очередь
                await base_sync.publish_once(
                    exchange=DJANGO_UVICORN_NGINX_EXCHANGE,
                    routing_key=queue_name,
                    body=json.dumps(event),
                )
            except Exception as e:
                errors.append({'event': event.get('txId'), 'error': str(e)})

        if errors:
            return Response(
                {
                    'status': 'PARTIAL_SUCCESS',
                    'processed': len(serializer_data) - len(errors),
                    'errors': errors
                },
                status=status.HTTP_400_BAD_REQUEST,
            )

        return Response(
            {
                'status': 'SUCCESS',
                'processed': len(serializer_data)
            },
            status=status.HTTP_200_OK,
        )
