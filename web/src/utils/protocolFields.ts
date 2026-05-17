import type { ProtocolFieldSpec, ProtocolFieldType, ProtocolSpec } from '@/types'

export type DynamicFieldKind = 'header' | 'query' | 'json'
export type DynamicIndexMode = 'first' | 'any'

export interface DynamicFieldParts {
  rootPath: string
  kind: DynamicFieldKind
  suffix: string
  indexMode: DynamicIndexMode
  label: string
  placeholder: string
  indexed: boolean
}

const jsonOperators = [
  'eq',
  'ne',
  'in',
  'not_in',
  'contains',
  'not_contains',
  'exists',
  'not_exists',
  'is_null',
  'is_not_null',
  'prefix',
  'suffix',
  'regex',
  'gt',
  'gte',
  'lt',
  'lte',
]

const hiddenDefaultRuleConditionFields = new Set([
  'request.scheme',
  'request.host',
  'request.original_host',
  'request.client_ip',
])

export function fieldForPath(protocolSpec: ProtocolSpec | undefined, fieldPath: string): ProtocolFieldSpec | undefined {
  if (!fieldPath) return undefined
  for (const field of protocolSpec?.fields || []) {
    if (field.path === fieldPath) return field
    if (!field.dynamic_path) continue
    if (fieldPath.startsWith(`${field.path}.`) || fieldPath.startsWith(`${field.path}[`)) {
      return { ...field, type: 'json', operators: undefined }
    }
  }
  return undefined
}

export function operatorsForField(field: ProtocolFieldSpec | undefined): string[] {
  if (!field) return []
  return field.operators?.length ? field.operators : defaultOperatorsForType(field.type)
}

export function ruleConditionFields(fields: ProtocolFieldSpec[]): ProtocolFieldSpec[] {
  return sortRuleConditionFields(fields.filter((field) => isRuleConditionFieldPath(field.path)))
}

export function isRuleConditionFieldPath(path: string): boolean {
  const normalized = path.trim()
  if (!normalized) return false
  return normalized !== 'protocol'
    && normalized !== 'namespace'
    && normalized !== 'meta'
    && !normalized.startsWith('meta.')
    && !hiddenDefaultRuleConditionFields.has(normalized)
}

export function sortRuleConditionFields<T extends Pick<ProtocolFieldSpec, 'path'>>(fields: T[]): T[] {
  return [...fields].sort((left, right) => {
    return ruleConditionFieldSortKey(left.path).localeCompare(ruleConditionFieldSortKey(right.path))
  })
}

export function dynamicFieldParts(
  fieldPath: string,
  field: ProtocolFieldSpec | undefined
): DynamicFieldParts | null {
  if (!field?.dynamic_path) return null
  const rootPath = field.path
  const suffix = fieldPath === rootPath ? '' : fieldPath.slice(rootPath.length + 1)
  if (rootPath === 'request.headers') {
    const parsed = parseIndexedSuffix(suffix)
    return {
      rootPath,
      kind: 'header',
      suffix: parsed.key,
      indexMode: parsed.indexMode,
      label: 'Header key',
      placeholder: 'x-env',
      indexed: true,
    }
  }
  if (rootPath === 'request.query') {
    const parsed = parseIndexedSuffix(suffix)
    return {
      rootPath,
      kind: 'query',
      suffix: parsed.key,
      indexMode: parsed.indexMode,
      label: 'Query key',
      placeholder: 'state',
      indexed: true,
    }
  }
  return {
    rootPath,
    kind: 'json',
    suffix,
    indexMode: 'first',
    label: rootPath === 'request.value' ? 'Value path' : 'JSON path',
    placeholder: rootPath === 'request.value' ? 'user.id' : 'status',
    indexed: false,
  }
}

export function buildDynamicFieldPath(
  rootPath: string,
  suffix: string,
  indexMode: DynamicIndexMode = 'first'
): string {
  const normalizedSuffix = normalizeDynamicSuffix(rootPath, suffix)
  if (!normalizedSuffix) return rootPath
  if (rootPath === 'request.headers' || rootPath === 'request.query') {
    const index = indexMode === 'any' ? '*' : '0'
    return `${rootPath}.${normalizedSuffix}[${index}]`
  }
  return `${rootPath}.${normalizedSuffix}`
}

export function defaultOperatorsForType(type: ProtocolFieldType): string[] {
  if (type === 'number') {
    return ['eq', 'ne', 'in', 'not_in', 'gt', 'gte', 'lt', 'lte', 'exists', 'not_exists', 'is_null', 'is_not_null']
  }
  if (type === 'bool') {
    return ['eq', 'ne', 'exists', 'not_exists', 'is_null', 'is_not_null']
  }
  if (type === 'object') {
    return ['exists', 'not_exists', 'is_null', 'is_not_null']
  }
  if (type === 'array') {
    return ['contains', 'not_contains', 'exists', 'not_exists', 'is_null', 'is_not_null']
  }
  if (type === 'json') {
    return jsonOperators
  }
  return [
    'eq',
    'ne',
    'in',
    'not_in',
    'contains',
    'not_contains',
    'prefix',
    'suffix',
    'regex',
    'exists',
    'not_exists',
    'is_null',
    'is_not_null',
  ]
}

function parseIndexedSuffix(value: string): { key: string; indexMode: DynamicIndexMode } {
  const match = value.match(/^(.*)\[(\*|0|\d+)]$/)
  if (!match) return { key: value, indexMode: 'first' }
  return {
    key: match[1],
    indexMode: match[2] === '*' ? 'any' : 'first',
  }
}

function normalizeDynamicSuffix(rootPath: string, suffix: string): string {
  let normalized = suffix.trim().replace(/^\.+|\.+$/g, '')
  if (rootPath === 'request.headers') {
    normalized = normalized.toLowerCase()
  }
  if (rootPath === 'request.headers' || rootPath === 'request.query') {
    normalized = normalized.replace(/\[(?:\*|0|\d+)]$/, '')
  }
  return normalized
}

function ruleConditionFieldSortKey(path: string): string {
  return path.trim()
    .toLowerCase()
    .replace(/\[(\d+)]/g, (_match, index: string) => `[${index.padStart(8, '0')}]`)
    .replace(/\[\*]/g, '[99999999]')
}
