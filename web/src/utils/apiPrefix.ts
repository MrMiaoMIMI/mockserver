export function normalizeApiPrefix(value?: string): string {
  const prefix = (value || '').trim()
  if (!prefix || prefix === '/') {
    return ''
  }
  const withLeadingSlash = prefix.startsWith('/') ? prefix : `/${prefix}`
  return withLeadingSlash.replace(/\/+$/, '')
}

export function apiPath(path: string, prefix = getApiPrefix()): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizeApiPrefix(prefix)}${normalizedPath}`
}

export function getApiPrefix(): string {
  const runtimePrefix =
    typeof window !== 'undefined' ? window.__MOCKSERVER_API_PREFIX__ || '' : ''
  return normalizeApiPrefix(runtimePrefix || import.meta.env.VITE_API_PREFIX || '')
}
