/** @type {import('next').NextConfig} */
const BACKEND = process.env.BACKEND_URL || 'http://localhost:8013'

const nextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/login',         destination: `${BACKEND}/login` },
      { source: '/callback',      destination: `${BACKEND}/callback` },
      { source: '/api/:path*',    destination: `${BACKEND}/api/:path*` },
    ]
  },
}

export default nextConfig
