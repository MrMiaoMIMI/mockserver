import { describe, expect, it } from 'vitest'
import {
  buildSampleRequestFields,
  parseOptionalSampleJSON,
  parseSampleRequestInput,
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

  it('auto-detects JSON body samples and prepares authoring metadata', () => {
    const parsed = parseSampleRequestInput('{"id":"demo","count":123}', 'http')

    expect(parsed.error).toBeUndefined()
    expect(parsed.format).toBe('json')
    expect(parsed.root).toBe('request.body')
    expect(parsed.fields.map((field) => field.path)).toEqual([
      'request.body.count',
      'request.body.id',
    ])
    expect(parsed.authoring).toMatchObject({
      format: 'json',
      root: 'request.body',
      parsed: { id: 'demo', count: 123 },
    })
  })

  it('auto-detects cURL samples and infers http request fields', () => {
    const parsed = parseSampleRequestInput(
      `curl -X POST 'https://demo.com/api/order?region=SG' \\
        -H 'content-type: application/json' \\
        -H 'x-env: test' \\
        --data-raw '{"order_id":"o1","amount":123}'`,
      'http'
    )

    expect(parsed.error).toBeUndefined()
    expect(parsed.format).toBe('curl')
    expect(parsed.authoring?.parsed).toMatchObject({
      method: 'POST',
      host: 'demo.com',
      path: '/api/order',
      query: { region: ['SG'] },
      headers: {
        'content-type': ['application/json'],
        'x-env': ['test'],
      },
      body: { order_id: 'o1', amount: 123 },
    })
    expect(parsed.fields.map((field) => field.path)).toEqual([
      'request.body.amount',
      'request.body.order_id',
      'request.headers.content-type[*]',
      'request.headers.content-type[0]',
      'request.headers.x-env[*]',
      'request.headers.x-env[0]',
      'request.host',
      'request.method',
      'request.path',
      'request.query.region[*]',
      'request.query.region[0]',
    ])
  })

  it('rejects invalid sample requests before they can be used', () => {
    expect(parseSampleRequestInput('{', 'http').error).toContain('JSON body is invalid')
    expect(parseSampleRequestInput(`curl https://demo.com -d 'a=1'`, 'http').error).toContain('valid JSON')
    expect(parseSampleRequestInput(`curl https://demo.com -H 'broken'`, 'http').error).toContain('Invalid cURL header')
    expect(parseSampleRequestInput(`curl https://demo.com`, 'spex').error).toContain('only supported for HTTP')
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
