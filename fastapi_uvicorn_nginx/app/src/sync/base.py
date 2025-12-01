import asyncio
import gzip
import random
import typing

import aio_pika
from aio_pika.exceptions import AMQPException
from app.config import settings


class BaseSync:
    def __init__(self):
        self.mq_connection: typing.Union[None, aio_pika.RobustConnection] = None
        self.mq_channel: typing.Union[None, aio_pika.Channel] = None
        self.exchanges = {}

    async def close(self) -> None:
        """
        Отключаем коннекты
        """
        if self.mq_channel:
            await self.mq_channel.close()
        if self.mq_connection:
            await self.mq_connection.close()

    async def connect(self) -> None:
        self.mq_connection = await aio_pika.connect_robust(
            settings.RABBITMQ_DSN,
            heartbeat=60,
        )
        self.mq_channel = await self.mq_connection.channel()

    async def connect_once(self):
        if self.mq_connection is None or self.mq_connection.is_closed or self.mq_channel is None or self.mq_channel.is_closed:
            await self.connect()

    async def publish_once(
        self,
        exchange: str,
        routing_key: str,
        body: typing.Union[str, int, float, bool, None, typing.Dict[str, typing.Any], typing.List[typing.Any]],
        mandatory: bool = False,
    ) -> None:
        """
        Разовая публикация сообщения без явного создания подключения к rabbitmq.
        Аналогично basic_publish()
        """
        back_off_ms = BackOffMs(retries=5)

        if isinstance(body, str):
            body = gzip.compress(body.encode('utf-8'))

        while True:
            try:
                exchange_obj = self.exchanges.get(exchange)
                if exchange_obj is None:
                    exchange_obj = await self.mq_channel.declare_exchange(
                        name=exchange,
                        type=aio_pika.ExchangeType.DIRECT,
                        durable=True,
                        passive=False,
                    )
                    self.exchanges[exchange] = exchange_obj
                message = aio_pika.Message(
                    body=body,
                    content_type='application/json',
                    content_encoding='gzip',
                    delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                )
                await exchange_obj.publish(
                    message,
                    routing_key=routing_key,
                    mandatory=mandatory,
                )
                return
            except AMQPException as e:
                timeout = back_off_ms()
                if timeout == BackOffMs.MAX_RETRIES_EXCEED:
                    raise e
                await asyncio.sleep(timeout / 1000)  # Задержка

    async def declare_queues(self, exchange: str, queues: list = None, arguments: dict = None):
        exchange_obj = await self.mq_channel.declare_exchange(
            name=exchange, type=aio_pika.ExchangeType.DIRECT, durable=True, auto_delete=False,
        )
        if not queues:
            queues = list()
        for queue in queues:
            queue_obj = await self.mq_channel.declare_queue(name=queue, durable=True, arguments=arguments)
            await queue_obj.bind(exchange=exchange_obj, routing_key=queue)


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


rabbit_client = None


async def get_rabbit_client():
    """Получить или создать RabbitMQ клиент (singleton)"""
    global rabbit_client
    if rabbit_client is None:
        rabbit_client = BaseSync()
        await rabbit_client.connect()
    return rabbit_client
