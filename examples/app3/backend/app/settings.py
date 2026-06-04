"""환경 변수 로딩 + 검증."""
from __future__ import annotations
import os
from dataclasses import dataclass

from dotenv import load_dotenv

load_dotenv()


def _required(key: str) -> str:
    v = os.environ.get(key)
    if not v:
        raise RuntimeError(f"Missing required env: {key}")
    return v


@dataclass(frozen=True)
class Settings:
    oauth_server_url: str
    oauth_internal_url: str
    client_id: str
    client_secret: str
    redirect_uri: str
    frontend_url: str
    session_secret: str
    database_url: str
    port: int


_server_url = _required("OAUTH_SERVER_URL")
settings = Settings(
    # 브라우저 redirect 용
    oauth_server_url=_server_url,
    # 백엔드 token/userinfo 호출 용 (compose 에서는 컨테이너 hostname)
    oauth_internal_url=os.environ.get("OAUTH_INTERNAL_URL") or _server_url,
    client_id=_required("OAUTH_CLIENT_ID"),
    client_secret=_required("OAUTH_CLIENT_SECRET"),
    redirect_uri=_required("OAUTH_REDIRECT_URI"),
    frontend_url=_required("FRONTEND_URL"),
    session_secret=_required("SESSION_SECRET"),
    database_url=_required("DATABASE_URL"),
    port=int(os.environ.get("PORT", "8013")),
)
