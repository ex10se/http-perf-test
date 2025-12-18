import gzip
import random
import socket
import threading
import typing
from time import sleep

import pika
from django.conf import settings
from pika.exceptions import (
    ChannelClosed,
    ConnectionClosed,
    StreamLostError,
    ChannelWrongStateError,
)


class BaseSync:
    def __init__(self):
        self.mq_connection: typing.Union[None, pika.BlockingConnection] = None
        self.mq_channel: typing.Union[None, pika.adapters.blocking_connection.BlockingChannel] = None
        self.mq_lock = threading.Lock()

    def close(self) -> None:
        """
        Отключаем коннекты
        """
        with self.mq_lock:
            if self.mq_channel:
                self.mq_channel.close()
            if self.mq_connection:
                self.mq_connection.close()

    def connect(self) -> None:
        params = pika.URLParameters(settings.RABBITMQ['default'])
        parameters = pika.ConnectionParameters(
            host=params.host,
            port=params.port or 5672,
            virtual_host=params.virtual_host,
            connection_attempts=5,
            retry_delay=1,
            credentials=params.credentials,
            heartbeat=60,
            blocked_connection_timeout=60,
        )

        with self.mq_lock:
            self.mq_connection = pika.BlockingConnection(parameters)
            self.mq_channel = self.mq_connection.channel()

    def connect_once(self):
        if self.mq_connection is None or self.mq_connection.is_closed or self.mq_channel is None or self.mq_channel.is_closed:
            self.connect()

    def publish_once(
        self,
        exchange: str,
        routing_key: str,
        body: typing.Union[str, int, float, bool, None, typing.Dict[str, typing.Any], typing.List[typing.Any]],
        properties: pika.BasicProperties = None,
        mandatory: bool = False,
    ) -> None:
        """
        Разовая публикация сообщения без явного создания подключения к rabbitmq.
        Аналогично basic_publish()
        """
        self.connect_once()
        back_off_ms = BackOffMs(retries=5)

        if isinstance(body, str):
            body = gzip.compress(body.encode('utf-8'))

        while True:
            try:
                with self.mq_lock:
                    self.mq_channel.basic_publish(exchange, routing_key, body, properties, mandatory)
                return
            except (
                    ChannelClosed, ConnectionClosed, ConnectionError, StreamLostError,
                    ChannelWrongStateError, IOError, socket.error,
            ) as e:
                timeout = back_off_ms()
                if timeout == BackOffMs.MAX_RETRIES_EXCEED:
                    raise e
                sleep(timeout / 1000)  # Задержка
                self.connect_once()

    def declare_queues(self, exchange: str, queues: list = None, arguments: dict = None):
        self.mq_channel.exchange_declare(exchange=exchange, exchange_type='direct', durable=True, auto_delete=False)
        if not queues:
            queues = list()
        for queue in queues:
            self.mq_channel.queue_declare(queue=queue, durable=True, arguments=arguments)
            self.mq_channel.queue_bind(queue=queue, exchange=exchange, routing_key=queue)


class BackOffMs:
    """
    Вспомогательный класс для получения значений timeout согласно логике backoff

    Возвращает миллисекунды
    """
    MAX_RETRIES_EXCEED = -1

    def __init__(
        self,
        retries: int = 5,
        min_timeout: float = 1000,
        max_timeout: float = 5000,
        min_jitter: float = .1,
        max_jitter: float = .4,
        factor: float = 2,
    ):
        self.retries = retries
        self.min_timeout = min_timeout
        self.max_timeout = max_timeout
        self.min_jitter = min_jitter
        self.max_jitter = max_jitter
        self.factor = factor
        self.timeout = None
        self.retry = 0

    def __call__(self):
        if self.retry == 0:
            self.timeout = self.min_timeout * random.uniform(self.min_jitter, self.max_jitter)
        elif self.retry > self.retries:
            self.timeout = self.MAX_RETRIES_EXCEED
        else:
            jitter = random.uniform(self.min_jitter, self.max_jitter)
            if self.timeout >= self.max_timeout:
                self.timeout *= jitter
            else:
                self.timeout = self.timeout * self.factor + self.timeout * jitter
        self.retry += 1
        return self.timeout
