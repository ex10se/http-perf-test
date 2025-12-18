import asyncio
import logging

from django.core.management import BaseCommand
from sync.base import BaseSync
from sync.queues import QUEUES_DECLARATION

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = 'Queue declare task'

    def handle(self, *args, **kwargs):
        async def async_declare():
            base_sync = BaseSync()
            try:
                await base_sync.connect()
                for queues_group in QUEUES_DECLARATION:
                    await base_sync.declare_queues(**queues_group)
                await base_sync.close()
            except BaseException as e:
                logger.error(f'queue declare error: {e}')

        asyncio.run(async_declare())
