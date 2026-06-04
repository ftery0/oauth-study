import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

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
