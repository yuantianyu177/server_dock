import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { readFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

function loadHttpsOptions() {
  const certPath = process.env.TLS_CERT_PATH
  const keyPath = process.env.TLS_KEY_PATH

  if (!certPath && !keyPath) return undefined
  if (!certPath || !keyPath) {
    throw new Error('TLS_CERT_PATH and TLS_KEY_PATH must be configured together')
  }

  return {
    cert: readFileSync(certPath),
    key: readFileSync(keyPath)
  }
}

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) }
  },
  server: {
    port: 3000,
    https: loadHttpsOptions(),
    proxy: {
      '/api': {
        target: process.env.BACKEND_URL || 'http://localhost:8000',
        changeOrigin: true,
        ws: true
      }
    }
  }
})
