"""OAuth client + 자동 프로비저닝."""
from __future__ import annotations

import base64
import secrets
from typing import Any
from urllib.parse import urlencode

import httpx
from fastapi import APIRouter, Depends, Request
from fastapi.responses import JSONResponse, RedirectResponse
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from .db import get_session
from .models import User
from .settings import settings

router = APIRouter()


def _basic_auth() -> str:
    raw = f"{settings.client_id}:{settings.client_secret}".encode()
    return base64.b64encode(raw).decode()


async def _exchange_token(code: str) -> dict[str, Any] | None:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{settings.oauth_internal_url}/oauth/token",
            data={
                "grant_type": "authorization_code",
                "code": code,
                "redirect_uri": settings.redirect_uri,
            },
            headers={
                "Authorization": f"Basic {_basic_auth()}",
                "Content-Type": "application/x-www-form-urlencoded",
            },
        )
    if resp.status_code != 200:
        return None
    return resp.json()


async def _refresh_access_token(refresh: str) -> dict[str, Any] | None:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{settings.oauth_internal_url}/oauth/token",
            data={"grant_type": "refresh_token", "refresh_token": refresh},
            headers={
                "Authorization": f"Basic {_basic_auth()}",
                "Content-Type": "application/x-www-form-urlencoded",
            },
        )
    if resp.status_code != 200:
        return None
    return resp.json()


async def _fetch_user_info(access_token: str) -> dict[str, Any] | None:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{settings.oauth_internal_url}/oauth/userinfo",
            headers={"Authorization": f"Bearer {access_token}"},
        )
    if resp.status_code != 200:
        return None
    return resp.json()


async def _upsert_user(session: AsyncSession, sub: str) -> User:
    existing = await session.scalar(select(User).where(User.sub == sub))
    if existing is not None:
        return existing
    user = User(sub=sub, display_name=sub, email=None)
    session.add(user)
    await session.commit()
    return user


@router.get("/login")
async def login(request: Request):
    state = secrets.token_hex(16)
    request.session["oauth_state"] = state
    qs = urlencode({
        "response_type": "code",
        "client_id": settings.client_id,
        "redirect_uri": settings.redirect_uri,
        "scope": "openid profile",
        "state": state,
    })
    return RedirectResponse(f"{settings.oauth_server_url}/oauth/authorize?{qs}")


@router.get("/callback")
async def callback(
    request: Request,
    code: str | None = None,
    state: str | None = None,
    error: str | None = None,
    session: AsyncSession = Depends(get_session),
):
    if error:
        return RedirectResponse(f"{settings.frontend_url}?error={error}")

    session_state = request.session.get("oauth_state")
    if not state or state != session_state:
        return RedirectResponse(f"{settings.frontend_url}?error=invalid_state")
    request.session.pop("oauth_state", None)

    if not code:
        return RedirectResponse(f"{settings.frontend_url}?error=missing_code")

    tokens = await _exchange_token(code)
    if tokens is None:
        return RedirectResponse(f"{settings.frontend_url}?error=token_exchange_failed")

    request.session["access_token"] = tokens["access_token"]
    request.session["refresh_token"] = tokens.get("refresh_token")
    # id_token: RP-initiated logout 의 id_token_hint 로 사용 (openid scope 일 때 IdP 가 발급)
    if isinstance(tokens.get("id_token"), str):
        request.session["id_token"] = tokens["id_token"]

    info = await _fetch_user_info(tokens["access_token"])
    if info and isinstance(info.get("sub"), str):
        await _upsert_user(session, info["sub"])
        request.session["user_sub"] = info["sub"]

    return RedirectResponse(settings.frontend_url)


@router.get("/api/me")
async def me(request: Request, session: AsyncSession = Depends(get_session)):
    access_token = request.session.get("access_token")
    if not access_token:
        return JSONResponse({"error": "Not authenticated"}, status_code=401)

    info = await _fetch_user_info(access_token)
    if info is None:
        refresh = request.session.get("refresh_token")
        if not refresh:
            request.session.clear()
            return JSONResponse({"error": "Session expired"}, status_code=401)
        refreshed = await _refresh_access_token(refresh)
        if refreshed is None:
            request.session.clear()
            return JSONResponse({"error": "Session expired"}, status_code=401)
        request.session["access_token"] = refreshed["access_token"]
        request.session["refresh_token"] = refreshed.get("refresh_token")
        info = await _fetch_user_info(refreshed["access_token"])

    if info is None:
        return JSONResponse({"error": "Failed to fetch user info"}, status_code=500)

    if isinstance(info.get("sub"), str):
        user = await _upsert_user(session, info["sub"])
        request.session["user_sub"] = info["sub"]
        info["display_name"] = user.display_name

    return JSONResponse(info)


@router.get("/api/logout")
async def logout(request: Request):
    """앱 세션 무효화 + IdP RP-initiated logout 으로 redirect.

    app 세션만 끊으면 IdP 세션이 살아있어 silent SSO 가 다시 자동 로그인시킴.
    IdP /oauth/logout 으로 보내야 IdP 세션도 같이 폐기되고 진짜 로그아웃.
    """
    id_token_hint = request.session.get("id_token")
    request.session.clear()

    params: dict[str, str] = {"post_logout_redirect_uri": settings.frontend_url}
    if isinstance(id_token_hint, str):
        params["id_token_hint"] = id_token_hint
    qs = urlencode(params)
    return RedirectResponse(f"{settings.oauth_server_url}/oauth/logout?{qs}")


async def require_user_sub(request: Request) -> str:
    sub = request.session.get("user_sub")
    if not isinstance(sub, str) or not sub:
        from fastapi import HTTPException

        raise HTTPException(status_code=401, detail="Not authenticated")
    return sub
