import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Vite proxy 를 쓰는 이유:
// - 브라우저는 모든 요청을 :5182 로 보내고, vite 가 :8012(backend) 로 전달 → CORS 회피
// - 백엔드 세션 쿠키가 localhost 단일 도메인으로 정상 동작
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5182,
    proxy: {
      '/login':    'http://localhost:8012',
      '/callback': 'http://localhost:8012',
      '/api':      'http://localhost:8012',
    },
  },
})
