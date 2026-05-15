import { describe, expect, it } from 'vitest'
import type { ProtocolSpec } from '@/types'
import {
  buildResponsePayload,
  defaultResponsePayload,
  responseFieldDraftsFromPayload,
} from '@/utils/responseSpec'

const spexSpec: ProtocolSpec = {
  name: 'spex',
  fields: [],
  response: {
    defaults: { code: 0, resp: {} },
    fields: [
      { path: 'code', type: 'number', required: true, default: 0 },
      { path: 'resp', type: 'json', required: true, default: {} },
    ],
  },
}

const httpSpec: ProtocolSpec = {
  name: 'http',
  fields: [],
  response: {
    defaults: { status: 200, headers: { 'content-type': ['application/json'] }, body: {} },
    fields: [
      { path: 'status', type: 'number', required: true, default: 200, min: 100, max: 599 },
      { path: 'headers', type: 'object', default: { 'content-type': ['application/json'] } },
      { path: 'body', type: 'json', default: {} },
    ],
  },
}

describe('response spec helpers', () => {
  it('uses protocol metadata defaults for direct JSON response fields', () => {
    const payload = defaultResponsePayload(spexSpec)
    const drafts = responseFieldDraftsFromPayload(spexSpec, payload)

    expect(payload).toEqual({ code: 0, resp: {} })
    expect(drafts.resp).toBe('{}')
  })

  it('builds response payloads from JSON field drafts', () => {
    const result = buildResponsePayload(spexSpec, { code: 0, resp: {} }, {
      resp: '{"order_status":"mocked"}',
    })

    expect(result.issues).toEqual([])
    expect(result.payload).toEqual({ code: 0, resp: { order_status: 'mocked' } })
  })

  it('reports invalid JSON and protocol field validation issues', () => {
    const result = buildResponsePayload(httpSpec, {
      status: 99,
      headers: { 'x-bad': [1] },
      body: {},
    }, {
      body: '{"unterminated"',
    })

    expect(result.issues.map((issue) => issue.field)).toEqual(['body', 'status', 'headers'])
  })
})
