import type { ProtocolFieldSpec, ProtocolFieldType } from '@/types'

export type ValueEditorKind = 'none' | 'text' | 'number' | 'boolean' | 'json' | 'list' | 'regex'
export type ValueScalarKind = 'string' | 'number' | 'boolean' | 'json'

export interface ValueInputSpec {
  fieldPath: string
  fieldType: ProtocolFieldType | 'unknown'
  operator: string
  editor: ValueEditorKind
  valueRequired: boolean
  scalarKind: ValueScalarKind
  jsonLiteral: boolean
  placeholder: string
  helperText: string
  examples: unknown[]
}

export interface BuildValueInputSpecOptions {
  fieldPath: string
  field?: ProtocolFieldSpec
  operator: string
  dynamicJSONLiteral?: boolean
}

export interface InvalidJSONLiteralValue {
  __mockserver_invalid_json_literal__: true
  raw: string
  message: string
}

const noValueOperators = new Set(['exists', 'not_exists', 'is_null', 'is_not_null'])
const listOperators = new Set(['in', 'not_in'])
const numericOperators = new Set(['gt', 'gte', 'lt', 'lte'])

export function buildValueInputSpec(options: BuildValueInputSpecOptions): ValueInputSpec {
  const fieldPath = options.fieldPath || ''
  const operator = options.operator || ''
  const fieldType = options.field?.type || 'unknown'
  const profile = fieldProfile(fieldPath)
  const dynamicChild = Boolean(options.field?.dynamic_path && options.field.path !== fieldPath)

  if (noValueOperators.has(operator)) {
    return {
      fieldPath,
      fieldType,
      operator,
      editor: 'none',
      valueRequired: false,
      scalarKind: 'string',
      jsonLiteral: false,
      placeholder: '',
      helperText: noValueHelper(operator),
      examples: [],
    }
  }

  const jsonLiteral = usesJSONLiteralEditor(fieldPath, options.field, options.dynamicJSONLiteral ?? true)
  if (jsonLiteral) {
    return {
      fieldPath,
      fieldType,
      operator,
      editor: 'json',
      valueRequired: true,
      scalarKind: 'json',
      jsonLiteral: true,
      placeholder: jsonLiteralPlaceholder(operator),
      helperText: jsonLiteralHelper(operator),
      examples: profile.examples.length ? profile.examples : ['demo', 123, true, ['a', 'b'], { id: 'u1' }],
    }
  }

  if (listOperators.has(operator)) {
    const scalarKind = scalarKindForField(fieldType, operator, dynamicChild)
    return {
      fieldPath,
      fieldType,
      operator,
      editor: 'list',
      valueRequired: true,
      scalarKind,
      jsonLiteral: false,
      placeholder: profile.listPlaceholder || profile.placeholder,
      helperText: `${operatorLabel(operator)} requires one or more values. Each chip becomes one array item.`,
      examples: profile.listExamples.length ? profile.listExamples : [profile.examples.slice(0, 2).filter(Boolean)],
    }
  }

  if (operator === 'regex') {
    return {
      fieldPath,
      fieldType,
      operator,
      editor: 'regex',
      valueRequired: true,
      scalarKind: 'string',
      jsonLiteral: false,
      placeholder: profile.regexPlaceholder || '^/api/v[0-9]+/',
      helperText: 'Enter a regular expression. It is checked before the rule is saved.',
      examples: profile.regexExamples.length ? profile.regexExamples : ['^/api/.*'],
    }
  }

  if (numericOperators.has(operator)) {
    return {
      fieldPath,
      fieldType,
      operator,
      editor: 'number',
      valueRequired: true,
      scalarKind: 'number',
      jsonLiteral: false,
      placeholder: profile.numberPlaceholder || '100',
      helperText: `${operatorLabel(operator)} compares numeric values. The value will be saved as a number.`,
      examples: profile.numberExamples.length ? profile.numberExamples : [100, 1000],
    }
  }

  const scalarKind = scalarKindForField(fieldType, operator, dynamicChild)
  return {
    fieldPath,
    fieldType,
    operator,
    editor: editorForScalar(fieldType, scalarKind, dynamicChild),
    valueRequired: true,
    scalarKind,
    jsonLiteral: false,
    placeholder: profile.placeholder || placeholderForScalar(scalarKind),
    helperText: helperForScalar(fieldPath, operator, scalarKind, dynamicChild),
    examples: profile.examples.length ? profile.examples : examplesForScalar(scalarKind),
  }
}

export function valueInputNeedsValue(spec: ValueInputSpec): boolean {
  return spec.valueRequired && spec.editor !== 'none'
}

export function isSmartValueEmpty(value: unknown, spec: ValueInputSpec): boolean {
  if (!valueInputNeedsValue(spec)) return false
  const normalized = normalizeValueForSpec(value, spec)
  if (spec.editor === 'list') return Array.isArray(normalized) && normalized.length === 0
  if (normalized === undefined || normalized === null) return true
  if (typeof normalized === 'string') return !normalized.trim()
  return false
}

export function normalizeValueForSpec(value: unknown, spec: ValueInputSpec): unknown {
  if (!valueInputNeedsValue(spec)) return undefined
  if (isInvalidJSONLiteralValue(value)) return value
  if (spec.editor === 'list') {
    if (Array.isArray(value)) return value
    if (value === undefined || value === null) return []
    if (typeof value === 'string' && !value.trim()) return []
    return [value]
  }
  if (spec.editor === 'number' && typeof value === 'string' && value.trim()) {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : value
  }
  if (spec.editor === 'boolean') {
    if (value === 'true') return true
    if (value === 'false') return false
  }
  return value
}

export function defaultValueForSpec(spec: ValueInputSpec): unknown {
  if (spec.editor === 'none') return undefined
  if (spec.editor === 'list') return []
  if (spec.editor === 'boolean') return true
  if (spec.editor === 'number') return undefined
  if (spec.editor === 'json') return null
  return ''
}

export function coerceListItem(value: string, spec: ValueInputSpec): unknown {
  const trimmed = value.trim()
  if (!trimmed) return ''
  if (spec.scalarKind === 'number') {
    const parsed = Number(trimmed)
    return Number.isFinite(parsed) ? parsed : trimmed
  }
  if (spec.scalarKind === 'boolean') {
    if (trimmed.toLowerCase() === 'true') return true
    if (trimmed.toLowerCase() === 'false') return false
  }
  if (spec.scalarKind === 'json') {
    try {
      return JSON.parse(trimmed)
    } catch {
      return trimmed
    }
  }
  return trimmed
}

export function invalidJSONLiteralValue(raw: string, message: string): InvalidJSONLiteralValue {
  return {
    __mockserver_invalid_json_literal__: true,
    raw,
    message,
  }
}

export function isInvalidJSONLiteralValue(value: unknown): value is InvalidJSONLiteralValue {
  return Boolean(
    value
      && typeof value === 'object'
      && (value as Partial<InvalidJSONLiteralValue>).__mockserver_invalid_json_literal__ === true
  )
}

export function operatorLabel(operator: string): string {
  if (operator === 'eq') return 'equals'
  if (operator === 'ne') return 'not equals'
  if (operator === 'in') return 'in list'
  if (operator === 'not_in') return 'not in list'
  if (operator === 'contains') return 'contains'
  if (operator === 'not_contains') return 'does not contain'
  if (operator === 'prefix') return 'starts with'
  if (operator === 'suffix') return 'ends with'
  if (operator === 'gt') return 'greater than'
  if (operator === 'gte') return 'greater than or equal'
  if (operator === 'lt') return 'less than'
  if (operator === 'lte') return 'less than or equal'
  return operator
}

function scalarKindForField(
  fieldType: ProtocolFieldType | 'unknown',
  operator: string,
  dynamicChild: boolean
): ValueScalarKind {
  if (numericOperators.has(operator)) return 'number'
  if (fieldType === 'number') return 'number'
  if (fieldType === 'bool') return 'boolean'
  if (fieldType === 'object' || fieldType === 'array') return 'json'
  if (fieldType === 'json' && !dynamicChild) return 'json'
  return 'string'
}

function editorForScalar(
  fieldType: ProtocolFieldType | 'unknown',
  scalarKind: ValueScalarKind,
  dynamicChild: boolean
): ValueEditorKind {
  if (scalarKind === 'number') return 'number'
  if (scalarKind === 'boolean') return 'boolean'
  if (scalarKind === 'json') return fieldType === 'json' && dynamicChild ? 'text' : 'json'
  return 'text'
}

function noValueHelper(operator: string): string {
  if (operator === 'exists') return 'Matches when the path exists. No value is needed.'
  if (operator === 'not_exists') return 'Matches when the path is missing. No value is needed.'
  if (operator === 'is_null') return 'Matches when the path is missing, or exists with null / Go nil.'
  if (operator === 'is_not_null') return 'Matches when the path exists and its value is not null / Go nil.'
  return 'No value is needed.'
}

function helperForScalar(
  fieldPath: string,
  operator: string,
  scalarKind: ValueScalarKind,
  dynamicChild: boolean
): string {
  if (scalarKind === 'number') return `${operatorLabel(operator)} compares numeric values.`
  if (scalarKind === 'boolean') return 'Choose a boolean value. The value will be saved as true or false.'
  if (scalarKind === 'json') return 'Enter valid JSON. Objects, arrays, strings, numbers, booleans, and null are supported.'
  if (fieldPath.includes('.headers.')) return 'Header names are normalized to lower case before matching.'
  if (fieldPath.includes('.query.')) return 'Query values are arrays. Use [0] for the first value or [*] for any value.'
  if (dynamicChild) return 'This dynamic JSON child is saved as a plain value unless you switch to raw JSON.'
  return `${operatorLabel(operator)} compares this field as text.`
}

function placeholderForScalar(scalarKind: ValueScalarKind): string {
  if (scalarKind === 'number') return '100'
  if (scalarKind === 'boolean') return 'true'
  if (scalarKind === 'json') return '{"status":"active"}'
  return 'value'
}

function examplesForScalar(scalarKind: ValueScalarKind): unknown[] {
  if (scalarKind === 'number') return [100, 1000]
  if (scalarKind === 'boolean') return [true, false]
  if (scalarKind === 'json') return [{ status: 'active' }, ['a', 'b'], null]
  return ['active', 'test']
}

function usesJSONLiteralEditor(
  fieldPath: string,
  field: ProtocolFieldSpec | undefined,
  dynamicJSONLiteral: boolean
): boolean {
  if (dynamicJSONLiteral && isDynamicJSONLiteralPath(fieldPath)) return true
  if (!field) return false
  if (dynamicJSONLiteral && isDynamicJSONLiteralRoot(field.path)) return true
  return field.path === fieldPath && field.type === 'json'
}

function isDynamicJSONLiteralPath(path: string): boolean {
  return ['request.body', 'request.req', 'request.value', 'meta.extra'].some((root) => (
    path === root || path.startsWith(`${root}.`) || path.startsWith(`${root}[`)
  ))
}

function isDynamicJSONLiteralRoot(path: string): boolean {
  return path === 'request.body'
    || path === 'request.req'
    || path === 'request.value'
    || path === 'meta.extra'
}

function jsonLiteralPlaceholder(operator: string) {
  if (listOperators.has(operator)) return '["a", "b"]'
  if (numericOperators.has(operator)) return '123'
  return '"demo", 123, true, null, ["a", "b"], {"id":"u1"}'
}

function jsonLiteralHelper(operator: string) {
  if (listOperators.has(operator)) {
    return `${operatorLabel(operator)} expects one JSON array. Strings inside the array must use double quotes.`
  }
  if (numericOperators.has(operator)) {
    return `${operatorLabel(operator)} expects a JSON number. Numeric-looking strings such as "123" will not match numbers.`
  }
  return 'Enter one JSON value. Strings must use double quotes, for example "demo". Numbers, booleans, null, arrays, and objects keep their JSON type.'
}

function fieldProfile(fieldPath: string) {
  const normalized = fieldPath.toLowerCase()
  if (normalized === 'request.method') {
    return profile({
      placeholder: 'GET',
      examples: ['GET', 'POST', 'PUT', 'DELETE'],
      listExamples: [['GET', 'POST']],
    })
  }
  if (normalized === 'request.operation') {
    return profile({
      placeholder: 'get',
      examples: ['get', 'set', 'del'],
      listExamples: [['get', 'set']],
    })
  }
  if (normalized === 'request.path') {
    return profile({
      placeholder: '/api/v1/debug',
      examples: ['/api/', '/api/v1/debug'],
      regexPlaceholder: '^/api/v[0-9]+/',
      regexExamples: ['^/api/v[0-9]+/', '/orders/[0-9]+$'],
    })
  }
  if (normalized === 'request.host' || normalized === 'request.original_host') {
    return profile({
      placeholder: 'demo.com',
      examples: ['demo.com', 'api.demo.com'],
      regexPlaceholder: '(^|\\.)demo\\.com$',
    })
  }
  if (normalized === 'request.key') {
    return profile({
      placeholder: 'user:123',
      examples: ['user:', 'user:123', 'order:'],
      listExamples: [['user:123', 'user:456']],
    })
  }
  if (normalized === 'request.ttl_ms') {
    return profile({
      placeholder: '60000',
      examples: [60000, 300000],
      numberPlaceholder: '60000',
      numberExamples: [60000, 300000],
    })
  }
  if (normalized.includes('.headers.')) {
    return profile({
      placeholder: 'test',
      examples: ['test', 'staging', 'prod'],
    })
  }
  if (normalized.includes('.query.')) {
    return profile({
      placeholder: 'qv1',
      examples: ['qv1', 'active', 'vip'],
      listExamples: [['active', 'pending']],
    })
  }
  if (normalized.startsWith('request.body')) {
    return profile({
      placeholder: '"doing"',
      examples: ['doing', 'done', 1, true],
      listExamples: [['doing', 'done']],
    })
  }
  if (normalized.startsWith('request.req')) {
    return profile({
      placeholder: '"demo"',
      examples: ['demo', 123, true, ['a', 'b'], { id: 'u1' }],
      listExamples: [['demo', 'active']],
    })
  }
  if (normalized.startsWith('request.value')) {
    return profile({
      placeholder: '"active"',
      examples: ['active', 'inactive', 1, true],
      listExamples: [['active', 'inactive']],
    })
  }
  if (normalized === 'meta.trace_id') {
    return profile({
      placeholder: 'trace-001',
      examples: ['trace-001', 'debug-trace'],
    })
  }
  return profile({})
}

function profile(input: Partial<{
  placeholder: string
  listPlaceholder: string
  regexPlaceholder: string
  numberPlaceholder: string
  examples: unknown[]
  listExamples: unknown[]
  regexExamples: unknown[]
  numberExamples: unknown[]
}>) {
  return {
    placeholder: input.placeholder || '',
    listPlaceholder: input.listPlaceholder || '',
    regexPlaceholder: input.regexPlaceholder || '',
    numberPlaceholder: input.numberPlaceholder || '',
    examples: input.examples || [],
    listExamples: input.listExamples || [],
    regexExamples: input.regexExamples || [],
    numberExamples: input.numberExamples || [],
  }
}
