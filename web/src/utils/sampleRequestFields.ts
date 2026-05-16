import type { ProtocolFieldSpec, ProtocolFieldType } from '@/types'
import { isInvalidJSONLiteralValue } from '@/utils/valueInputSpec'

export type SampleValueType = 'string' | 'number' | 'boolean' | 'null' | 'object' | 'array' | 'mixed' | 'unknown'

export interface SampleRequestField extends ProtocolFieldSpec {
  source: 'sample'
  sample_value_type: SampleValueType
  preview: string
}

const noValueOperators = new Set(['exists', 'not_exists', 'is_null', 'is_not_null'])
const listOperators = new Set(['in', 'not_in'])
const numericOperators = new Set(['gt', 'gte', 'lt', 'lte'])

export function sampleRootForProtocol(protocol: string) {
  const normalized = protocol.toLowerCase()
  if (normalized === 'spex') return 'request.req'
  if (normalized === 'cache') return 'request.value'
  return 'request.body'
}

export function buildSampleRequestFields(rootPath: string, sample: unknown): SampleRequestField[] {
  if (!rootPath || sample === undefined) return []
  const fields = new Map<string, SampleRequestField>()
  visitValue(fields, rootPath, sample, true)
  return [...fields.values()].sort((left, right) => left.path.localeCompare(right.path))
}

export function parseOptionalSampleJSON(raw: string): { value?: unknown; error?: string } {
  if (!raw.trim()) return {}
  try {
    return { value: JSON.parse(raw) }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    return { error: `Sample request JSON is invalid: ${message}` }
  }
}

export function sampleTypeWarning(
  fieldPath: string,
  operator: string,
  value: unknown,
  sampleFields: SampleRequestField[]
) {
  if (!fieldPath || noValueOperators.has(operator) || isInvalidJSONLiteralValue(value)) return ''
  const sampleField = sampleFields.find((field) => field.path === fieldPath)
  if (!sampleField) return ''

  if (listOperators.has(operator)) {
    if (!Array.isArray(value)) {
      return `${fieldPath} looks like ${sampleTypeLabel(sampleField.sample_value_type)} in the sample. ${operator} expects a JSON array of candidate values.`
    }
    const mismatch = value.find((item) => !sampleTypesCompatible(sampleField.sample_value_type, valueTypeOf(item)))
    if (mismatch !== undefined) {
      return `${fieldPath} looks like ${sampleTypeLabel(sampleField.sample_value_type)} in the sample, but one candidate is ${sampleTypeLabel(valueTypeOf(mismatch))}.`
    }
    return ''
  }

  const valueType = valueTypeOf(value)
  if (numericOperators.has(operator) && valueType !== 'number') {
    return `${fieldPath} uses a numeric operator, but the condition value is ${sampleTypeLabel(valueType)}.`
  }
  if (!sampleTypesCompatible(sampleField.sample_value_type, valueType)) {
    return `${fieldPath} looks like ${sampleTypeLabel(sampleField.sample_value_type)} in the sample, but the condition value is ${sampleTypeLabel(valueType)}.`
  }
  return ''
}

export function valueTypeOf(value: unknown): SampleValueType {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  if (typeof value === 'string') return 'string'
  if (typeof value === 'number') return 'number'
  if (typeof value === 'boolean') return 'boolean'
  if (typeof value === 'object') return 'object'
  return 'unknown'
}

function visitValue(fields: Map<string, SampleRequestField>, path: string, value: unknown, isRoot = false) {
  const sampleType = valueTypeOf(value)
  if (!isRoot) {
    addField(fields, path, sampleType, value)
  }

  if (sampleType === 'object' && value && !Array.isArray(value)) {
    Object.entries(value as Record<string, unknown>).forEach(([key, child]) => {
      if (!isSupportedPathKey(key)) return
      visitValue(fields, `${path}.${key}`, child)
    })
    return
  }

  if (sampleType === 'array' && Array.isArray(value)) {
    const itemPath = `${path}[*]`
    const itemType = arrayItemType(value)
    addField(fields, itemPath, itemType, value[0])
    collectArrayObjectChildren(fields, itemPath, value)
  }
}

function collectArrayObjectChildren(fields: Map<string, SampleRequestField>, itemPath: string, items: unknown[]) {
  items.forEach((item) => {
    if (!item || typeof item !== 'object' || Array.isArray(item)) return
    Object.entries(item as Record<string, unknown>).forEach(([key, child]) => {
      if (!isSupportedPathKey(key)) return
      visitValue(fields, `${itemPath}.${key}`, child)
    })
  })
}

function addField(fields: Map<string, SampleRequestField>, path: string, sampleType: SampleValueType, value: unknown) {
  if (fields.has(path)) return
  fields.set(path, {
    path,
    type: protocolTypeForSample(sampleType),
    source: 'sample',
    sample_value_type: sampleType,
    preview: previewValue(value),
  })
}

function arrayItemType(items: unknown[]): SampleValueType {
  if (!items.length) return 'unknown'
  const itemTypes = new Set(items.map(valueTypeOf))
  return itemTypes.size === 1 ? [...itemTypes][0] : 'mixed'
}

function protocolTypeForSample(sampleType: SampleValueType): ProtocolFieldType {
  if (sampleType === 'number') return 'number'
  if (sampleType === 'boolean') return 'bool'
  if (sampleType === 'object') return 'object'
  if (sampleType === 'array') return 'array'
  if (sampleType === 'string') return 'string'
  return 'json'
}

function sampleTypesCompatible(expected: SampleValueType, actual: SampleValueType) {
  if (expected === 'unknown' || expected === 'mixed') return true
  return expected === actual
}

function sampleTypeLabel(type: SampleValueType) {
  if (type === 'boolean') return 'boolean'
  return type
}

function isSupportedPathKey(key: string) {
  return Boolean(key) && !/[.[\]]/.test(key)
}

function previewValue(value: unknown) {
  if (typeof value === 'string') return value
  if (value === undefined) return ''
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}
