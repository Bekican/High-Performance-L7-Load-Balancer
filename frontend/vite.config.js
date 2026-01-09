import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/stats': 'http://localhost:9091',
      '/api': 'http://localhost:9091'
    }
  }
})
