"""app3 backend: HelpDesk · FastAPI + PostgreSQL OAuth client (group-b)."""
from __future__ import annotations

from contextlib import asynccontextmanager

from fastapi import FastAPI
from starlette.middleware.sessions import SessionMiddleware

from app.db import Base, engine
from app.oauth import router as oauth_router
from app.settings import settings
from app.tickets import router as tickets_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield


app = FastAPI(lifespan=lifespan)
app.add_middleware(
    SessionMiddleware,
    secret_key=settings.session_secret,
    same_site="lax",
    https_only=False,
    max_age=24 * 60 * 60,
)

app.include_router(oauth_router)
app.include_router(tickets_router)


if __name__ == "__main__":
    import os
    import uvicorn

    host = os.environ.get("HOST", "127.0.0.1")
    uvicorn.run(app, host=host, port=settings.port)
