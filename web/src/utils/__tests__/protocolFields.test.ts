import { describe, expect, it } from 'vitest'
import type { ProtocolFieldSpec, ProtocolSpec } from '@/types'
import {
  buildDynamicFieldPath,
  dynamicFieldParts,
  fieldForPath,
  isRuleConditionFieldPath,
  operatorsForField,
  ruleConditionFields,
  sortRuleConditionFields,
} from '@/utils/protocolFields'

const cacheSpec: ProtocolSpec = {
  name: 'cache',
  fields: [
    { path: 'request.operation', type: 'string' },
    { path: 'request.key', type: 'string' },
    { path: 'request.value', type: 'json', dynamic_path: true },
    { path: 'request.headers', type: 'object', dynamic_path: true },
  ],
}

describe('protocol field helpers', () => {
  it('resolves dynamic child paths and derives default operators', () => {
    const valueField = fieldForPath(cacheSpec, 'request.value.user.id')
    const keyField = fieldForPath(cacheSpec, 'request.key')

    expect(valueField?.path).toBe('request.value')
    expect(valueField?.type).toBe('json')
    expect(operatorsForField(valueField)).toContain('regex')
    expect(operatorsForField(keyField)).toContain('prefix')
  })

  it('prefers explicit effective operators from protocol metadata', () => {
    expect(operatorsForField({ path: 'request.key', type: 'string', operators: ['eq', 'prefix'] })).toEqual([
      'eq',
      'prefix',
    ])
  })

  it('builds dynamic query and header paths from form parts', () => {
    const headerField = fieldForPath(cacheSpec, 'request.headers.x-env[*]')
    const parts = dynamicFieldParts('request.headers.x-env[*]', headerField)

    expect(parts).toMatchObject({
      rootPath: 'request.headers',
      suffix: 'x-env',
      indexMode: 'any',
      indexed: true,
    })
    expect(buildDynamicFieldPath('request.headers', 'X-Env', 'any')).toBe('request.headers.x-env[*]')
    expect(buildDynamicFieldPath('request.headers', 'X-Env', 'first')).toBe('request.headers.x-env[0]')
  })

  it('builds dynamic JSON child paths without array index controls', () => {
    const valueField = fieldForPath(cacheSpec, 'request.value.user.id')
    const parts = dynamicFieldParts('request.value.user.id', valueField)

    expect(parts).toMatchObject({
      rootPath: 'request.value',
      suffix: 'user.id',
      indexed: false,
    })
    expect(buildDynamicFieldPath('request.value', 'user.id')).toBe('request.value.user.id')
  })

  it('filters ruleset-level and internal metadata fields from rule conditions', () => {
    const fields = ruleConditionFields([
      { path: 'protocol', type: 'string' },
      { path: 'namespace', type: 'string' },
      { path: 'meta.trace_id', type: 'string' },
      { path: 'request.scheme', type: 'string' },
      { path: 'request.host', type: 'string' },
      { path: 'request.original_host', type: 'string' },
      { path: 'request.client_ip', type: 'string' },
      { path: 'request.cmd', type: 'string' },
      { path: 'request.req', type: 'json', dynamic_path: true },
    ])

    expect(fields.map((field) => field.path)).toEqual(['request.cmd', 'request.req'])
    expect(isRuleConditionFieldPath('meta.extra.debug')).toBe(false)
    expect(isRuleConditionFieldPath('request.host')).toBe(false)
    expect(isRuleConditionFieldPath('request.original_host')).toBe(false)
    expect(isRuleConditionFieldPath('request.body.id')).toBe(true)
  })

  it('sorts rule condition fields by normalized path name', () => {
    const fields: ProtocolFieldSpec[] = [
      { path: 'request.body.order_id', type: 'string' },
      { path: 'request.raw_body', type: 'string' },
      { path: 'request.headers.x-env[*]', type: 'string' },
      { path: 'request.query.region[*]', type: 'string' },
      { path: 'request.method', type: 'string' },
      { path: 'request.headers.x-env[0]', type: 'string' },
      { path: 'request.query.region[0]', type: 'string' },
      { path: 'request.path', type: 'string' },
      { path: 'request.body.amount', type: 'number' },
    ]

    expect(sortRuleConditionFields(fields).map((field) => field.path)).toEqual([
      'request.body.amount',
      'request.body.order_id',
      'request.headers.x-env[0]',
      'request.headers.x-env[*]',
      'request.method',
      'request.path',
      'request.query.region[0]',
      'request.query.region[*]',
      'request.raw_body',
    ])
  })

  it('sorts SPEX and cache rule condition fields without protocol-specific priority', () => {
    expect(sortRuleConditionFields([
      { path: 'request.req.user_id', type: 'number' },
      { path: 'request.param', type: 'string' },
      { path: 'request.cmd', type: 'string' },
    ]).map((field) => field.path)).toEqual([
      'request.cmd',
      'request.param',
      'request.req.user_id',
    ])

    expect(sortRuleConditionFields([
      { path: 'request.value.user_id', type: 'number' },
      { path: 'request.ttl_ms', type: 'number' },
      { path: 'request.key', type: 'string' },
      { path: 'request.operation', type: 'string' },
    ]).map((field) => field.path)).toEqual([
      'request.key',
      'request.operation',
      'request.ttl_ms',
      'request.value.user_id',
    ])
  })
})
