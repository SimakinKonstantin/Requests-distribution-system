import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/employees': 'http://localhost:8080',
      '/slots':     'http://localhost:8080',
      '/appeals':   'http://localhost:8080',
      '/subthemes': 'http://localhost:8080',
    },
  },
})
