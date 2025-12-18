from typing import Optional

from pydantic import BaseModel


class ErrorSchema(BaseModel):
    code: Optional[str] = None
    message: Optional[str] = None


class TrackDataSchema(BaseModel):
    priority: Optional[int] = None
    is_system: bool = False


class StatusEventSchema(BaseModel):
    state: str
    error: Optional[ErrorSchema] = None
    trackData: Optional[TrackDataSchema] = None
    updatedAt: str
    txId: str
    email: Optional[str] = None
    channel_id: Optional[str] = None
    channel: Optional[str] = None

