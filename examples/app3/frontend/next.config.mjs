/** @type {import('next').NextConfig} */
const nextConfig = {
  // Next.js dev 서버(5183) 가 FastAPI backend(8013) 로 프록시.
  // 같은 origin 으로 묶여 세션 쿠키가 정상 동작.
  async rewrites() {
    return [
      { source: '/login',         destination: 'http://localhost:8013/login' },
      { source: '/callback',      destination: 'http://localhost:8013/callback' },
      { source: '/api/me',        destination: 'http://localhost:8013/api/me' },
      { source: '/api/logout',    destination: 'http://localhost:8013/api/logout' },
    ]
  },
}

export default nextConfig
