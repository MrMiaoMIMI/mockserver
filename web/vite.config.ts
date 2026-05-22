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

function apiPrefixFromBaseURL(value?: string) {
  const baseURL = (value || '').trim()
  if (!baseURL) {
    return ''
  }
  if (baseURL.startsWith('/')) {
    return normalizeApiPrefix(baseURL)
  }
  try {
    return normalizeApiPrefix(new URL(baseURL).pathname)
  } catch {
    return ''
  }
}

const apiPrefix = normalizeApiPrefix(
  process.env.VITE_API_PREFIX || apiPrefixFromBaseURL(process.env.VITE_API_BASE_URL)
)
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
