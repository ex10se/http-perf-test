import asyncio
import logging

from sync.base import BaseSync
from sync.queues import QUEUES_DECLARATION

logger = logging.getLogger(__name__)


async def async_declare() -> None:
    base_sync = BaseSync()
    try:
        await base_sync.connect()
        for queues_group in QUEUES_DECLARATION:
            await base_sync.declare_queues(**queues_group)
        await base_sync.close()
    except BaseException as exc:
        logger.error(f"queue declare error: {exc}")


def main() -> None:
    asyncio.run(async_declare())


if __name__ == "__main__":
    main()
