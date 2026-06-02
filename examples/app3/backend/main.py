"""
app3 backend: FastAPI OAuth client (group-b 멤버).

backend 세션 패턴:
- 토큰은 starlette SessionMiddleware 가 관리하는 세션에 저장
- 브라우저에는 itsdangerous 로 서명된 세션 쿠키만 전달
- 토큰 자체는 클라이언트에 노출되지 않음
"""
import base64
import os
import secrets
from urllib.parse import urlencode

import httpx
from dotenv import load_dotenv
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, RedirectResponse
from starlette.middleware.sessions import SessionMiddleware

load_dotenv()

OAUTH_SERVER_URL = os.environ["OAUTH_SERVER_URL"]
CLIENT_ID = os.environ["OAUTH_CLIENT_ID"]
CLIENT_SECRET = os.environ["OAUTH_CLIENT_SECRET"]
REDIRECT_URI = os.environ["OAUTH_REDIRECT_URI"]
FRONTEND_URL = os.environ["FRONTEND_URL"]
SESSION_SECRET = os.environ["SESSION_SECRET"]
PORT = int(os.environ.get("PORT", "8013"))

app = FastAPI()
app.add_middleware(
    SessionMiddleware,
    secret_key=SESSION_SECRET,
    same_site="lax",
    https_only=False,  # 프로덕션에서는 True
    max_age=24 * 60 * 60,
)


# ──────────────────────────────────────────
# GET /login
# ──────────────────────────────────────────
@app.get("/login")
async def login(request: Request):
    state = secrets.token_hex(16)
    request.session["oauth_state"] = state

    qs = urlencode({
        "response_type": "code",
        "client_id": CLIENT_ID,
        "redirect_uri": REDIRECT_URI,
        "scope": "openid profile",
        "state": state,
    })
    return RedirectResponse(f"{OAUTH_SERVER_URL}/oauth/authorize?{qs}")


# ──────────────────────────────────────────
# GET /callback
# ──────────────────────────────────────────
@app.get("/callback")
async def callback(
    request: Request,
    code: str | None = None,
    state: str | None = None,
    error: str | None = None,
):
    if error:
        return RedirectResponse(f"{FRONTEND_URL}?error={error}")

    session_state = request.session.get("oauth_state")
    if not state or state != session_state:
        return RedirectResponse(f"{FRONTEND_URL}?error=invalid_state")
    request.session.pop("oauth_state", None)

    if not code:
        return RedirectResponse(f"{FRONTEND_URL}?error=missing_code")

    tokens = await _exchange_token(code)
    if tokens is None:
        return RedirectResponse(f"{FRONTEND_URL}?error=token_exchange_failed")

    request.session["access_token"] = tokens["access_token"]
    request.session["refresh_token"] = tokens.get("refresh_token")

    return RedirectResponse(FRONTEND_URL)


# ──────────────────────────────────────────
# GET /api/me — 401 시 자동 refresh
# ──────────────────────────────────────────
@app.get("/api/me")
async def me(request: Request):
    access_token = request.session.get("access_token")
    if not access_token:
        return JSONResponse({"error": "Not authenticated"}, status_code=401)

    user_info = await _fetch_user_info(access_token)
    if user_info is None:
        # 만료 → refresh 시도
        if not await _refresh_access_token(request):
            request.session.clear()
            return JSONResponse(
                {"error": "Session expired. Please login again."},
                status_code=401,
            )
        access_token = request.session["access_token"]
        user_info = await _fetch_user_info(access_token)

    if user_info is None:
        return JSONResponse(
            {"error": "Failed to fetch user info"}, status_code=500
        )
    return JSONResponse(user_info)


# ──────────────────────────────────────────
# POST /api/logout
# ──────────────────────────────────────────
@app.post("/api/logout")
async def logout(request: Request):
    request.session.clear()
    return {"ok": True}


# ──────────────────────────────────────────
# 헬퍼
# ──────────────────────────────────────────
def _basic_auth() -> str:
    raw = f"{CLIENT_ID}:{CLIENT_SECRET}".encode()
    return base64.b64encode(raw).decode()


async def _exchange_token(code: str) -> dict | None:
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{OAUTH_SERVER_URL}/oauth/token",
            data={
                "grant_type": "authorization_code",
                "code": code,
                "redirect_uri": REDIRECT_URI,
            },
            headers={
                "Authorization": f"Basic {_basic_auth()}",
                "Content-Type": "application/x-www-form-urlencoded",
            },
        )
    if resp.status_code != 200:
        return None
    return resp.json()


async def _refresh_access_token(request: Request) -> bool:
    refresh = request.session.get("refresh_token")
    if not refresh:
        return False

    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{OAUTH_SERVER_URL}/oauth/token",
            data={
                "grant_type": "refresh_token",
                "refresh_token": refresh,
            },
            headers={
                "Authorization": f"Basic {_basic_auth()}",
                "Content-Type": "application/x-www-form-urlencoded",
            },
        )
    if resp.status_code != 200:
        return False

    tokens = resp.json()
    # Token Rotation: 갱신마다 새 refresh token 저장
    request.session["access_token"] = tokens["access_token"]
    request.session["refresh_token"] = tokens.get("refresh_token")
    return True


async def _fetch_user_info(access_token: str) -> dict | None:
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{OAUTH_SERVER_URL}/oauth/userinfo",
            headers={"Authorization": f"Bearer {access_token}"},
        )
    if resp.status_code != 200:
        return None
    return resp.json()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=PORT)
