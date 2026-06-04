import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

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
