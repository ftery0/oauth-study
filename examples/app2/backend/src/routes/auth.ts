import { Router } from 'express'
import crypto from 'crypto'
import { env } from '../env'
import { UserModel } from '../models/User'

export const authRouter = Router()

interface TokenResponse {
  access_token: string
  refresh_token?: string
  token_type?: string
  expires_in?: number
}

function basicAuth(): string {
  return 'Basic ' + Buffer.from(`${env.CLIENT_ID}:${env.CLIENT_SECRET}`).toString('base64')
}

async function exchangeToken(code: string): Promise<TokenResponse | null> {
  const r = await fetch(`${env.OAUTH_INTERNAL}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded', Authorization: basicAuth() },
    body: new URLSearchParams({
      grant_type:   'authorization_code',
      code,
      redirect_uri: env.REDIRECT_URI,
    }).toString(),
  })
  if (!r.ok) return null
  return (await r.json()) as TokenResponse
}

async function fetchUserInfo(accessToken: string): Promise<Record<string, unknown> | null> {
  const r = await fetch(`${env.OAUTH_INTERNAL}/oauth/userinfo`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
  if (!r.ok) return null
  return (await r.json()) as Record<string, unknown>
}

async function refreshAccessToken(refresh: string): Promise<TokenResponse | null> {
  const r = await fetch(`${env.OAUTH_INTERNAL}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded', Authorization: basicAuth() },
    body: new URLSearchParams({ grant_type: 'refresh_token', refresh_token: refresh }).toString(),
  })
  if (!r.ok) return null
  return (await r.json()) as TokenResponse
}

async function upsertUser(sub: string): Promise<void> {
  await UserModel.updateOne(
    { sub },
    { $setOnInsert: { sub, displayName: sub, email: null } },
    { upsert: true },
  )
}

authRouter.get('/login', (req, res) => {
  const state = crypto.randomBytes(16).toString('hex')
  req.session.oauthState = state

  const qs = new URLSearchParams({
    response_type: 'code',
    client_id:     env.CLIENT_ID,
    redirect_uri:  env.REDIRECT_URI,
    scope:         'openid profile',
    state,
  })
  res.redirect(`${env.OAUTH_SERVER}/oauth/authorize?${qs}`)
})

authRouter.get('/callback', async (req, res) => {
  const { code, state, error } = req.query as Record<string, string | undefined>

  if (error) return res.redirect(`${env.FRONTEND_URL}?error=${error}`)
  if (!state || state !== req.session.oauthState) {
    return res.redirect(`${env.FRONTEND_URL}?error=invalid_state`)
  }
  delete req.session.oauthState

  if (!code) return res.redirect(`${env.FRONTEND_URL}?error=missing_code`)

  const tokens = await exchangeToken(code)
  if (!tokens) return res.redirect(`${env.FRONTEND_URL}?error=token_exchange_failed`)

  req.session.accessToken  = tokens.access_token
  req.session.refreshToken = tokens.refresh_token

  const info = await fetchUserInfo(tokens.access_token)
  if (info && typeof info.sub === 'string') {
    await upsertUser(info.sub)
    req.session.userSub = info.sub
  }

  res.redirect(env.FRONTEND_URL)
})

authRouter.get('/api/me', async (req, res) => {
  if (!req.session.accessToken) {
    return res.status(401).json({ error: 'Not authenticated' })
  }

  let info = await fetchUserInfo(req.session.accessToken)

  if (!info) {
    if (!req.session.refreshToken) {
      req.session.destroy(() => undefined)
      return res.status(401).json({ error: 'Session expired' })
    }
    const refreshed = await refreshAccessToken(req.session.refreshToken)
    if (!refreshed) {
      req.session.destroy(() => undefined)
      return res.status(401).json({ error: 'Session expired' })
    }
    req.session.accessToken  = refreshed.access_token
    req.session.refreshToken = refreshed.refresh_token
    info = await fetchUserInfo(refreshed.access_token)
  }

  if (!info) return res.status(500).json({ error: 'Failed to fetch user info' })

  if (typeof info.sub === 'string') {
    await upsertUser(info.sub)
    req.session.userSub = info.sub
    const user = await UserModel.findOne({ sub: info.sub }).lean()
    if (user) (info as Record<string, unknown>).display_name = user.displayName
  }

  res.json(info)
})

authRouter.post('/api/logout', (req, res) => {
  req.session.destroy(() => res.json({ ok: true }))
})
