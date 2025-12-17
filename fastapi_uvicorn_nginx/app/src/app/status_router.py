import json
from typing import List

from app.schemas import StatusEventSchema
from fastapi import APIRouter, status
from fastapi.responses import JSONResponse
from sync.base import get_rabbit_client
from sync.queues import (
    FASTAPI_UVICORN_NGINX_EXCHANGE,
    QUEUE_FASTAPI_UVICORN_NGINX,
    QUEUE_SYSTEM_FASTAPI_UVICORN_NGINX,
)

router = APIRouter(prefix="/status", tags=["status"])


@router.post("/status/", status_code=status.HTTP_200_OK)
async def status_endpoint(events: List[StatusEventSchema]):
    if not events:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"error": "Request body is required"},
        )

    base_sync = await get_rabbit_client()
    errors = []

    for event in events:
        event_dict = event.model_dump()
        try:
            track_data = event_dict.get("trackData") or {}
            is_system = track_data.get("is_system", False)
            queue_name = (
                QUEUE_SYSTEM_FASTAPI_UVICORN_NGINX
                if is_system
                else QUEUE_FASTAPI_UVICORN_NGINX
            )

            await base_sync.publish_once(
                exchange=FASTAPI_UVICORN_NGINX_EXCHANGE,
                routing_key=queue_name,
                body=json.dumps(event_dict),
            )
        except Exception as exc:
            errors.append({"event": event_dict.get("txId"), "error": str(exc)})

    if errors:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={
                "status": "PARTIAL_SUCCESS",
                "processed": len(events) - len(errors),
                "errors": errors,
            },
        )

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={
            "status": "SUCCESS",
            "processed": len(events),
        },
    )

