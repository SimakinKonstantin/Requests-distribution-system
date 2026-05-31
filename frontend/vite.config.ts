import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    globals: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'text-summary'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/**/*.test.{ts,tsx}', 'src/vite-env.d.ts'],
    },
  },
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
