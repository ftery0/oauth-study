require('dotenv').config()

const express = require('express')
const session = require('express-session')
const crypto  = require('crypto')

const app = express()
app.use(express.json())
app.use(express.urlencoded({ extended: true }))

app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: { httpOnly: true, maxAge: 24 * 60 * 60 * 1000 }, // 24시간
}))


const OAUTH_SERVER  = process.env.OAUTH_SERVER_URL
const CLIENT_ID     = process.env.OAUTH_CLIENT_ID
const CLIENT_SECRET = process.env.OAUTH_CLIENT_SECRET
// Vite(5173) → Vite proxy → 이 서버(3001)로 전달되므로 5173 기준
const REDIRECT_URI  = process.env.OAUTH_REDIRECT_URI
const FRONTEND_URL  = process.env.FRONTEND_URL

// ─────────────────────────────────────────
// GET /login
// Authorization Code Flow 시작:
// state 생성 → OAuth 서버의 /authorize로 redirect
// ─────────────────────────────────────────
app.get('/login', (req, res) => {
  // state: CSRF 방어용 랜덤값 (로그인 전 세션에 저장, 콜백에서 검증)
  const state = crypto.randomBytes(16).toString('hex')
  req.session.oauthState = state

  const params = new URLSearchParams({
    response_type: 'code',
    client_id:     CLIENT_ID,
    redirect_uri:  REDIRECT_URI,
    scope:         'openid profile',
    state,
  })

  res.redirect(`${OAUTH_SERVER}/oauth/authorize?${params}`)
})

// ─────────────────────────────────────────
// GET /callback
// OAuth 서버가 auth code를 들고 여기로 redirect
// → code를 access token + refresh token으로 교환
// ─────────────────────────────────────────
app.get('/callback', async (req, res) => {
  const { code, state, error } = req.query

  if (error) {
    return res.redirect(`${FRONTEND_URL}?error=${error}`)
  }

  // state 검증: 세션에 저장한 값과 일치해야 함 (CSRF 방어)
  if (!state || state !== req.session.oauthState) {
    return res.redirect(`${FRONTEND_URL}?error=invalid_state`)
  }
  delete req.session.oauthState

  // auth code → tokens 교환 (서버-투-서버 요청)
  // 브라우저가 아닌 Express가 직접 OAuth 서버에 요청하는 이유:
  // client_secret을 브라우저에 노출하지 않기 위함
  const response = await fetch(`${OAUTH_SERVER}/oauth/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      // HTTP Basic Auth: base64(client_id:client_secret)
      'Authorization': 'Basic ' + Buffer.from(`${CLIENT_ID}:${CLIENT_SECRET}`).toString('base64'),
    },
    body: new URLSearchParams({
      grant_type:   'authorization_code',
      code,
      redirect_uri: REDIRECT_URI,
    }).toString(),
  })

  if (!response.ok) {
    console.error('Token exchange failed:', await response.text())
    return res.redirect(`${FRONTEND_URL}?error=token_exchange_failed`)
  }

  const tokens = await response.json()

  // 토큰을 세션에 저장 (브라우저에는 세션 ID 쿠키만 전달됨)
  req.session.accessToken  = tokens.access_token
  req.session.refreshToken = tokens.refresh_token

  res.redirect(FRONTEND_URL)
})

// ─────────────────────────────────────────
// GET /api/me
// 저장된 access token으로 OAuth 서버에서 유저 정보 조회
// access token 만료시 refresh token으로 자동 갱신
// ─────────────────────────────────────────
app.get('/api/me', async (req, res) => {
  if (!req.session.accessToken) {
    return res.status(401).json({ error: 'Not authenticated' })
  }

  let userInfo = await fetchUserInfo(req.session.accessToken)

  if (userInfo === null) {
    // access token 만료 → refresh token으로 새 토큰 발급
    console.log('[/api/me] access token 만료, refresh 시도...')
    const refreshed = await refreshAccessToken(req)
    if (!refreshed) {
      req.session.destroy()
      return res.status(401).json({ error: 'Session expired. Please login again.' })
    }
    userInfo = await fetchUserInfo(req.session.accessToken)
  }

  if (userInfo === null) {
    return res.status(500).json({ error: 'Failed to fetch user info' })
  }

  res.json(userInfo)
})

// ─────────────────────────────────────────
// POST /api/logout
// ─────────────────────────────────────────
app.post('/api/logout', (req, res) => {
  req.session.destroy()
  res.json({ ok: true })
})

// ─────────────────────────────────────────
// 헬퍼 함수
// ─────────────────────────────────────────

async function fetchUserInfo(accessToken) {
  const res = await fetch(`${OAUTH_SERVER}/oauth/userinfo`, {
    headers: { 'Authorization': `Bearer ${accessToken}` },
  })
  if (!res.ok) return null
  return res.json()
}

async function refreshAccessToken(req) {
  if (!req.session.refreshToken) return false

  const response = await fetch(`${OAUTH_SERVER}/oauth/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      'Authorization': 'Basic ' + Buffer.from(`${CLIENT_ID}:${CLIENT_SECRET}`).toString('base64'),
    },
    body: new URLSearchParams({
      grant_type:    'refresh_token',
      refresh_token: req.session.refreshToken,
    }).toString(),
  })

  if (!response.ok) return false

  const tokens = await response.json()
  // Token Rotation: 갱신마다 새 refresh token 저장
  req.session.accessToken  = tokens.access_token
  req.session.refreshToken = tokens.refresh_token
  console.log('[refreshAccessToken] 토큰 갱신 완료')
  return true
}

app.listen(3001, () => {
  console.log('OAuth example backend: http://localhost:3001')
  console.log(`  Client ID:    ${CLIENT_ID}`)
  console.log(`  Redirect URI: ${REDIRECT_URI}`)
  console.log(`  OAuth Server: ${OAUTH_SERVER}`)
})
