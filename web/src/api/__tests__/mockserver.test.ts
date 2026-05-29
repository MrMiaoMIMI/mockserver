import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { NamespaceConfig } from '@/types'

const httpMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/utils/request', () => ({
  http: httpMock,
}))

describe('mockserverApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates a namespace with POST', async () => {
    const { mockserverApi } = await import('@/api/mockserver')
    const namespace = namespacePayload('orders')

    await mockserverApi.saveNamespace(namespace)

    expect(httpMock.post).toHaveBeenCalledWith('/mockserver/api/v1/admin/namespaces', namespace)
    expect(httpMock.put).not.toHaveBeenCalled()
  })

  it('updates an existing namespace with PUT when update is explicit', async () => {
    const { mockserverApi } = await import('@/api/mockserver')
    const namespace = namespacePayload('orders')

    await mockserverApi.saveNamespace(namespace, true)

    expect(httpMock.put).toHaveBeenCalledWith('/mockserver/api/v1/admin/namespaces/orders', namespace)
    expect(httpMock.post).not.toHaveBeenCalled()
  })
})

function namespacePayload(name: string): NamespaceConfig {
  return {
    name,
    policies: {
      http: {
        ruleset_miss_action: {
          type: 'forward',
          forward: {
            timeout_ms: 5000,
          },
        },
        rule_miss_action: {
          type: 'forward',
          forward: {
            timeout_ms: 5000,
          },
        },
      },
    },
  }
}
