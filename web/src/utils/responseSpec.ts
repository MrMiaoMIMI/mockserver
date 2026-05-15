import type { ProtocolFieldSpec, ProtocolSpec } from '@/types'

export interface ResponseValidationIssue {
  field: string
  message: string
}

export type ResponseFieldDrafts = Record<string, string>

export function responseFields(protocolSpec?: ProtocolSpec): ProtocolFieldSpec[] {
  return protocolSpec?.response?.fields || []
}

export function defaultResponsePayload(protocolSpec?: ProtocolSpec): Record<string, unknown> {
  if (isRecord(protocolSpec?.response?.defaults)) {
    return clone(protocolSpec?.response?.defaults as Record<string, unknown>)
  }
  const payload: Record<string, unknown> = {}
  for (const field of responseFields(protocolSpec)) {
    setPath(payload, field.path, clone(field.default ?? defaultValueForType(field.type)))
  }
  return payload
}

export function responseFieldDraftsFromPayload(
  protocolSpec: ProtocolSpec | undefined,
  payload: Record<string, unknown>
): ResponseFieldDrafts {
  const drafts: ResponseFieldDrafts = {}
  for (const field of responseFields(protocolSpec)) {
    if (!usesJsonEditor(field)) continue
    const value = getPath(payload, field.path)
    drafts[field.path] = prettyJSON(value === undefined ? field.default ?? defaultValueForType(field.type) : value)
  }
  return drafts
}

export function buildResponsePayload(
  protocolSpec: ProtocolSpec | undefined,
  payload: Record<string, unknown>,
  drafts: ResponseFieldDrafts
): { payload: Record<string, unknown>; issues: ResponseValidationIssue[] } {
  const nextPayload = clone(payload || {})
  const issues: ResponseValidationIssue[] = []

  for (const field of responseFields(protocolSpec)) {
    if (!usesJsonEditor(field) || !(field.path in drafts)) continue
    const raw = drafts[field.path]
    try {
      setPath(nextPayload, field.path, JSON.parse(raw))
    } catch (error) {
      issues.push({
        field: field.path,
        message: `${field.path} JSON is not valid: ${toErrorMessage(error)}`,
      })
    }
  }

  for (const field of responseFields(protocolSpec)) {
    const value = getPath(nextPayload, field.path)
    if (value === undefined) {
      if (field.required) {
        issues.push({ field: field.path, message: `${field.path} is required` })
      }
      continue
    }
    validateFieldValue(field, value, issues)
  }

  return { payload: nextPayload, issues }
}

export function usesJsonEditor(field: ProtocolFieldSpec): boolean {
  if (isHeaderField(field)) return false
  return field.type === 'json' || field.type === 'object' || field.type === 'array'
}

export function isHeaderField(field: ProtocolFieldSpec): boolean {
  return field.path === 'headers' && field.type === 'object'
}

export function getResponsePath(payload: Record<string, unknown>, path: string) {
  return getPath(payload, path)
}

export function setResponsePath(payload: Record<string, unknown>, path: string, value: unknown) {
  const nextPayload = clone(payload || {})
  setPath(nextPayload, path, value)
  return nextPayload
}

export function prettyJSON(value: unknown) {
  return JSON.stringify(value === undefined ? null : value, null, 2)
}

function validateFieldValue(field: ProtocolFieldSpec, value: unknown, issues: ResponseValidationIssue[]) {
  if (field.type === 'number') {
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      issues.push({ field: field.path, message: `${field.path} must be a number` })
      return
    }
    if (field.min !== undefined && value < field.min) {
      issues.push({ field: field.path, message: `${field.path} must be >= ${field.min}` })
    }
    if (field.max !== undefined && value > field.max) {
      issues.push({ field: field.path, message: `${field.path} must be <= ${field.max}` })
    }
    return
  }
  if (field.type === 'bool' && typeof value !== 'boolean') {
    issues.push({ field: field.path, message: `${field.path} must be a boolean` })
    return
  }
  if (field.type === 'string' && typeof value !== 'string') {
    issues.push({ field: field.path, message: `${field.path} must be a string` })
    return
  }
  if (field.type === 'object') {
    if (!isRecord(value)) {
      issues.push({ field: field.path, message: `${field.path} must be an object` })
      return
    }
    if (isHeaderField(field)) validateHeaders(field.path, value, issues)
    return
  }
  if (field.type === 'array' && !Array.isArray(value)) {
    issues.push({ field: field.path, message: `${field.path} must be an array` })
  }
}

function validateHeaders(path: string, value: unknown, issues: ResponseValidationIssue[]) {
  if (!isRecord(value)) return
  for (const [key, raw] of Object.entries(value)) {
    if (!key.trim()) {
      issues.push({ field: path, message: `${path} contains an empty header name` })
      continue
    }
    if (typeof raw === 'string') continue
    if (Array.isArray(raw) && raw.every((item) => typeof item === 'string')) continue
    issues.push({ field: path, message: `${path}.${key} must be a string or string array` })
  }
}

function defaultValueForType(type: ProtocolFieldSpec['type']) {
  if (type === 'number') return 0
  if (type === 'bool') return false
  if (type === 'array') return []
  if (type === 'object') return {}
  if (type === 'json') return null
  return ''
}

function getPath(payload: Record<string, unknown>, path: string): unknown {
  if (!path) return payload
  const parts = path.split('.')
  let current: unknown = payload
  for (const part of parts) {
    if (!isRecord(current) || !(part in current)) return undefined
    current = current[part]
  }
  return current
}

function setPath(payload: Record<string, unknown>, path: string, value: unknown) {
  const parts = path.split('.').filter(Boolean)
  if (!parts.length) return
  let current: Record<string, unknown> = payload
  for (let index = 0; index < parts.length - 1; index += 1) {
    const part = parts[index]
    if (!isRecord(current[part])) {
      current[part] = {}
    }
    current = current[part] as Record<string, unknown>
  }
  current[parts[parts.length - 1]] = value
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}
