import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// vite proxy 로 backend(:8011) 호출을 같은 origin 으로 묶음
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5181,
    proxy: {
      '/login':    'http://localhost:8011',
      '/callback': 'http://localhost:8011',
      '/api':      'http://localhost:8011',
    },
  },
})
