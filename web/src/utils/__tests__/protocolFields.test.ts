import { describe, expect, it } from 'vitest'
import type { ProtocolSpec } from '@/types'
import {
  buildDynamicFieldPath,
  dynamicFieldParts,
  fieldForPath,
  isRuleConditionFieldPath,
  operatorsForField,
  ruleConditionFields,
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
      { path: 'request.cmd', type: 'string' },
      { path: 'request.req', type: 'json', dynamic_path: true },
    ])

    expect(fields.map((field) => field.path)).toEqual(['request.cmd', 'request.req'])
    expect(isRuleConditionFieldPath('meta.extra.debug')).toBe(false)
    expect(isRuleConditionFieldPath('request.body.id')).toBe(true)
  })
})
