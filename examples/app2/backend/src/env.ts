import dotenv from 'dotenv'
dotenv.config()

function required(key: string): string {
  const v = process.env[key]
  if (!v) throw new Error(`Missing required env: ${key}`)
  return v
}

export const env = {
  // 브라우저 redirect 용 (사용자가 접근하는 IdP URL)
  OAUTH_SERVER:   required('OAUTH_SERVER_URL'),
  // 백엔드 token/userinfo 호출 용 (compose 안에서는 컨테이너 hostname)
  OAUTH_INTERNAL: process.env.OAUTH_INTERNAL_URL ?? required('OAUTH_SERVER_URL'),
  CLIENT_ID:     required('OAUTH_CLIENT_ID'),
  CLIENT_SECRET: required('OAUTH_CLIENT_SECRET'),
  REDIRECT_URI:  required('OAUTH_REDIRECT_URI'),
  FRONTEND_URL:  required('FRONTEND_URL'),
  SESSION_SECRET: required('SESSION_SECRET'),
  MONGO_URL:     required('MONGO_URL'),
  PORT:          parseInt(process.env.PORT ?? '8012', 10),
}
