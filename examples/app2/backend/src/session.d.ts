import 'express-session'

declare module 'express-session' {
  interface SessionData {
    oauthState?: string
    accessToken?: string
    refreshToken?: string
    userSub?: string
  }
}
