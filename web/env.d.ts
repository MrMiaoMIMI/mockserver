/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_API_PREFIX?: string
  readonly VITE_MOCKSERVER_PROXY_TARGET?: string
}

interface Window {
  __MOCKSERVER_API_PREFIX__?: string
}
