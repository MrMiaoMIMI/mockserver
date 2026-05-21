import { describe, expect, it } from 'vitest'
import { apiPath, normalizeApiPrefix } from '../apiPrefix'

describe('apiPrefix', () => {
  it('normalizes empty and slash-only prefixes', () => {
    expect(normalizeApiPrefix()).toBe('')
    expect(normalizeApiPrefix('')).toBe('')
    expect(normalizeApiPrefix('/')).toBe('')
  })

  it('normalizes configured prefixes', () => {
    expect(normalizeApiPrefix('tenant-a/')).toBe('/tenant-a')
    expect(normalizeApiPrefix('/tenant-a/')).toBe('/tenant-a')
  })

  it('prepends prefix to existing api path', () => {
    expect(apiPath('/mockserver/api/v1/admin', '/tenant-a')).toBe(
      '/tenant-a/mockserver/api/v1/admin'
    )
    expect(apiPath('mockserver/api/v1/admin', '')).toBe('/mockserver/api/v1/admin')
  })
})
