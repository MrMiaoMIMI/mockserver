import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

function normalizeApiPrefix(value?: string) {
  const prefix = (value || '').trim()
  if (!prefix || prefix === '/') {
    return ''
  }
  const withLeadingSlash = prefix.startsWith('/') ? prefix : `/${prefix}`
  return withLeadingSlash.replace(/\/+$/, '')
}

const apiPrefix = normalizeApiPrefix(process.env.VITE_API_PREFIX)
const proxyBasePath = `${apiPrefix}/mockserver`

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "@/styles/mixins" as *;\n`,
      },
    },
  },
  server: {
    port: 6173,
    proxy: {
      [proxyBasePath]: {
        target: process.env.VITE_MOCKSERVER_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
