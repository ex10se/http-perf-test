from app.config import settings
from app.status_router import router as status_router
from fastapi import FastAPI

app = FastAPI(title=settings.APP_NAME)

app.include_router(status_router)
