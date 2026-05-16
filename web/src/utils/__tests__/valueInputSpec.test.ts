import { describe, expect, it } from 'vitest'
import type { ProtocolFieldSpec } from '@/types'
import {
  buildValueInputSpec,
  coerceListItem,
  defaultValueForSpec,
  isSmartValueEmpty,
  normalizeValueForSpec,
} from '@/utils/valueInputSpec'

const stringField: ProtocolFieldSpec = { path: 'request.method', type: 'string' }
const numberField: ProtocolFieldSpec = { path: 'request.ttl_ms', type: 'number' }
const boolField: ProtocolFieldSpec = { path: 'request.enabled', type: 'bool' }
const bodyRoot: ProtocolFieldSpec = { path: 'request.body', type: 'json', dynamic_path: true }

describe('value input spec', () => {
  it('uses no-value editors for existence and null operators', () => {
    const exists = buildValueInputSpec({ fieldPath: 'request.body.result', field: bodyRoot, operator: 'exists' })
    const isNull = buildValueInputSpec({ fieldPath: 'request.body.result', field: bodyRoot, operator: 'is_null' })

    expect(exists.editor).toBe('none')
    expect(exists.jsonLiteral).toBe(false)
    expect(exists.valueRequired).toBe(false)
    expect(isNull.helperText).toContain('missing')
    expect(defaultValueForSpec(exists)).toBeUndefined()
  })

  it('uses typed editors from field type and operator', () => {
    expect(buildValueInputSpec({ fieldPath: 'request.method', field: stringField, operator: 'eq' })).toMatchObject({
      editor: 'text',
      placeholder: 'GET',
    })
    expect(buildValueInputSpec({ fieldPath: 'request.ttl_ms', field: numberField, operator: 'gte' })).toMatchObject({
      editor: 'number',
      scalarKind: 'number',
    })
    expect(buildValueInputSpec({ fieldPath: 'request.enabled', field: boolField, operator: 'eq' })).toMatchObject({
      editor: 'boolean',
      scalarKind: 'boolean',
    })
  })

  it('uses list editor and coerces list values', () => {
    const methodList = buildValueInputSpec({ fieldPath: 'request.method', field: stringField, operator: 'in' })
    const ttlList = buildValueInputSpec({ fieldPath: 'request.ttl_ms', field: numberField, operator: 'in' })

    expect(methodList.editor).toBe('list')
    expect(methodList.examples[0]).toEqual(['GET', 'POST'])
    expect(coerceListItem('42', ttlList)).toBe(42)
    expect(isSmartValueEmpty([], methodList)).toBe(true)
    expect(normalizeValueForSpec('GET', methodList)).toEqual(['GET'])
  })

  it('uses JSON literal input for dynamic JSON roots and children', () => {
    const root = buildValueInputSpec({ fieldPath: 'request.body', field: bodyRoot, operator: 'eq' })
    const child = buildValueInputSpec({ fieldPath: 'request.body.status', field: bodyRoot, operator: 'eq' })
    const childIn = buildValueInputSpec({ fieldPath: 'request.body.status', field: bodyRoot, operator: 'in' })

    expect(root.editor).toBe('json')
    expect(root.jsonLiteral).toBe(true)
    expect(child.editor).toBe('json')
    expect(child.jsonLiteral).toBe(true)
    expect(child.placeholder).toContain('"demo"')
    expect(childIn.editor).toBe('json')
    expect(childIn.placeholder).toBe('["a", "b"]')
  })

  it('can keep dynamic JSON child fields on legacy text input for non-rule condition surfaces', () => {
    const child = buildValueInputSpec({
      fieldPath: 'request.body.status',
      field: bodyRoot,
      operator: 'eq',
      dynamicJSONLiteral: false,
    })

    expect(child.editor).toBe('text')
    expect(child.jsonLiteral).toBe(false)
  })

  it('uses regex editor with field-specific examples', () => {
    const spec = buildValueInputSpec({
      fieldPath: 'request.path',
      field: { path: 'request.path', type: 'string' },
      operator: 'regex',
    })

    expect(spec.editor).toBe('regex')
    expect(spec.placeholder).toContain('/api')
    expect(spec.examples[0]).toContain('/api')
  })
})
