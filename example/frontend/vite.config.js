import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Vite proxy를 사용하는 이유:
// - 브라우저 입장에서 모든 요청이 localhost:5173으로 가므로 CORS 없음
// - 세션 쿠키가 localhost 도메인으로 설정되어 정상 작동
// - Express 서버(3001)의 존재를 클라이언트 코드에서 몰라도 됨
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/login':    'http://localhost:3001',
      '/callback': 'http://localhost:3001',
      '/api':      'http://localhost:3001',
    },
  },
})
