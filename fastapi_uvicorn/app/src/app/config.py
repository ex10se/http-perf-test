import os

from dotenv import load_dotenv
from pydantic import BaseModel


class Settings(BaseModel):
    APP_NAME: str = 'fastapi-uvicorn'
    RABBITMQ_DSN: str

    @classmethod
    def from_env(cls) -> "Settings":
        load_dotenv()
        rabbitmq_dsn = os.getenv('DSN__RABBITMQ')
        if not rabbitmq_dsn:
            raise ValueError('Переменная окружения DSN__RABBITMQ обязательна')
        return cls(RABBITMQ_DSN=rabbitmq_dsn)


settings = Settings.from_env()
