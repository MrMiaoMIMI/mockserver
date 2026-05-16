import { describe, expect, it } from 'vitest'
import {
  buildSampleRequestFields,
  parseOptionalSampleJSON,
  sampleRootForProtocol,
  sampleTypeWarning,
} from '@/utils/sampleRequestFields'

describe('sample request fields', () => {
  it('maps protocol samples to dynamic JSON roots', () => {
    expect(sampleRootForProtocol('http')).toBe('request.body')
    expect(sampleRootForProtocol('spex')).toBe('request.req')
    expect(sampleRootForProtocol('cache')).toBe('request.value')
  })

  it('infers nested scalar and array wildcard fields', () => {
    const fields = buildSampleRequestFields('request.req', {
      id: 'demo',
      count: 123,
      enabled: true,
      tags: ['a', 'b'],
      items: [
        { sku: 'A-1', qty: 2 },
        { sku: 'B-2', qty: 5 },
      ],
    })

    expect(fields.map((field) => field.path)).toEqual([
      'request.req.count',
      'request.req.enabled',
      'request.req.id',
      'request.req.items',
      'request.req.items[*]',
      'request.req.items[*].qty',
      'request.req.items[*].sku',
      'request.req.tags',
      'request.req.tags[*]',
    ])
    expect(fields.find((field) => field.path === 'request.req.count')).toMatchObject({
      type: 'number',
      sample_value_type: 'number',
    })
    expect(fields.find((field) => field.path === 'request.req.items[*].sku')).toMatchObject({
      type: 'string',
      sample_value_type: 'string',
    })
  })

  it('parses optional sample JSON without requiring a sample', () => {
    expect(parseOptionalSampleJSON('')).toEqual({})
    expect(parseOptionalSampleJSON('{"id":"demo"}').value).toEqual({ id: 'demo' })
    expect(parseOptionalSampleJSON('{').error).toContain('invalid')
  })

  it('warns when a JSON literal value type differs from the sample', () => {
    const fields = buildSampleRequestFields('request.req', {
      id: 'demo',
      count: 123,
    })

    expect(sampleTypeWarning('request.req.count', 'eq', '123', fields)).toContain('number')
    expect(sampleTypeWarning('request.req.count', 'eq', 123, fields)).toBe('')
    expect(sampleTypeWarning('request.req.id', 'in', ['demo', 123], fields)).toContain('one candidate')
    expect(sampleTypeWarning('request.req.id', 'in', 'demo', fields)).toContain('JSON array')
  })
})
