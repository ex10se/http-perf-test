import json

import pika
from rest_framework import status
from rest_framework.decorators import action
from rest_framework.parsers import JSONParser
from rest_framework.request import Request
from rest_framework.response import Response
from rest_framework.viewsets import ViewSet
from sync.base import BaseSync
from sync.queues import (
    DJANGO_UWSGI_NGINX_EXCHANGE,
    QUEUE_DJANGO_UWSGI_NGINX,
    QUEUE_SYSTEM_DJANGO_UWSGI_NGINX,
)
from sync.serializers import StatusEventSerializer


class StatusViewSet(ViewSet):
    parser_classes = (JSONParser,)

    @action(methods=['POST'], detail=False, url_path='status')
    def status(self, request: Request) -> Response:
        """
        POST status/ - принимает JSON-массив событий и отправляет их в rabbitmq очереди
        """
        if not request.data:
            return Response(
                {'error': 'Request body is required'}, 
                status=status.HTTP_400_BAD_REQUEST
            )

        if not isinstance(request.data, list):
            return Response(
                {'error': 'Request body must be a JSON array'}, 
                status=status.HTTP_400_BAD_REQUEST
            )

        serializer = StatusEventSerializer(data=request.data, many=True)
        if not serializer.is_valid():
            return Response(
                {'error': 'Validation failed', 'details': serializer.errors}, 
                status=status.HTTP_400_BAD_REQUEST
            )

        base_sync = BaseSync()
        errors = []
        for event in serializer.validated_data:
            try:
                track_data = event.get('trackData') or {}
                is_system = track_data.get('is_system', False)
                queue_name = (
                    QUEUE_SYSTEM_DJANGO_UWSGI_NGINX 
                    if is_system 
                    else QUEUE_DJANGO_UWSGI_NGINX
                )

                # Отправляем событие в очередь
                base_sync.publish_once(
                    exchange=DJANGO_UWSGI_NGINX_EXCHANGE,
                    routing_key=queue_name,
                    body=json.dumps(event),
                    properties=pika.BasicProperties(
                        content_type='application/json',
                    ),
                )
            except Exception as e:
                errors.append({'event': event.get('txId'), 'error': str(e)})

        if errors:
            return Response(
                {
                    'status': 'PARTIAL_SUCCESS',
                    'processed': len(serializer.validated_data) - len(errors),
                    'errors': errors
                },
                status=status.HTTP_400_BAD_REQUEST,
            )

        return Response(
            {
                'status': 'SUCCESS',
                'processed': len(serializer.validated_data)
            },
            status=status.HTTP_200_OK
        )
