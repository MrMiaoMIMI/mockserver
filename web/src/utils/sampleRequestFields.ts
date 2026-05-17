import type {
  ProtocolFieldSpec,
  ProtocolFieldType,
  SampleRequestAuthoring,
  SampleRequestFormat,
} from '@/types'
import { isInvalidJSONLiteralValue } from '@/utils/valueInputSpec'

export type SampleValueType = 'string' | 'number' | 'boolean' | 'null' | 'object' | 'array' | 'mixed' | 'unknown'

export interface SampleRequestField extends ProtocolFieldSpec {
  source: 'sample'
  sample_value_type: SampleValueType
  preview: string
}

export interface ParsedSampleRequest {
  raw: string
  format?: SampleRequestFormat
  root: string
  fields: SampleRequestField[]
  authoring?: SampleRequestAuthoring
  error?: string
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

export function parseSampleRequestInput(raw: string, protocol: string): ParsedSampleRequest {
  const root = sampleRootForProtocol(protocol)
  const trimmed = raw.trim()
  if (!trimmed) {
    return { raw, root, fields: [] }
  }
  if (looksLikeCurl(trimmed)) {
    return parseCurlSampleRequest(raw, protocol, root)
  }
  const parsed = parseOptionalSampleJSON(raw)
  if (parsed.error) {
    return {
      raw,
      root,
      fields: [],
      error: `Sample Request JSON body is invalid: ${parsed.error}`,
    }
  }
  const fields = buildSampleRequestFields(root, parsed.value)
  return {
    raw,
    format: 'json',
    root,
    fields,
    authoring: {
      format: 'json',
      root,
      raw,
      parsed: parsed.value,
    },
  }
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

function addHTTPScalarField(fields: Map<string, SampleRequestField>, path: string, value: string) {
  addField(fields, path, 'string', value)
}

function addHTTPIndexedField(fields: Map<string, SampleRequestField>, path: string, values: string[]) {
  if (!values.length) return
  addField(fields, `${path}[0]`, 'string', values[0])
  addField(fields, `${path}[*]`, arrayItemType(values), values[0])
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

function parseCurlSampleRequest(raw: string, protocol: string, root: string): ParsedSampleRequest {
  if (protocol.toLowerCase() !== 'http') {
    return {
      raw,
      root,
      fields: [],
      error: 'cURL Sample Request is only supported for HTTP rulesets.',
    }
  }

  const tokenized = tokenizeShell(raw)
  if (tokenized.error) {
    return { raw, root, fields: [], error: tokenized.error }
  }
  const tokens = tokenized.tokens
  if (!tokens.length || tokens[0] !== 'curl') {
    return { raw, root, fields: [], error: 'cURL Sample Request must start with curl.' }
  }

  let method = ''
  let urlText = ''
  const headers: Record<string, string[]> = {}
  const dataParts: string[] = []

  const addHeader = (value: string) => {
    const separator = value.indexOf(':')
    if (separator <= 0) {
      throw new Error(`Invalid cURL header "${value}". Expected "Name: value".`)
    }
    const key = value.slice(0, separator).trim().toLowerCase()
    const headerValue = value.slice(separator + 1).trim()
    if (!key) {
      throw new Error(`Invalid cURL header "${value}". Header name is required.`)
    }
    headers[key] = [...(headers[key] || []), headerValue]
  }

  try {
    for (let index = 1; index < tokens.length; index += 1) {
      const token = tokens[index]
      if (token === '-X' || token === '--request') {
        method = readOptionValue(tokens, ++index, token)
        continue
      }
      if (token.startsWith('--request=')) {
        method = token.slice('--request='.length)
        continue
      }
      if (token.startsWith('-X') && token.length > 2) {
        method = token.slice(2)
        continue
      }
      if (token === '-H' || token === '--header') {
        addHeader(readOptionValue(tokens, ++index, token))
        continue
      }
      if (token.startsWith('--header=')) {
        addHeader(token.slice('--header='.length))
        continue
      }
      if (token.startsWith('-H') && token.length > 2) {
        addHeader(token.slice(2))
        continue
      }
      if (['-d', '--data', '--data-raw', '--data-binary'].includes(token)) {
        dataParts.push(readOptionValue(tokens, ++index, token))
        continue
      }
      const dataPrefix = ['--data=', '--data-raw=', '--data-binary='].find((prefix) => token.startsWith(prefix))
      if (dataPrefix) {
        dataParts.push(token.slice(dataPrefix.length))
        continue
      }
      if (token.startsWith('-d') && token.length > 2) {
        dataParts.push(token.slice(2))
        continue
      }
      if (token.startsWith('-')) {
        continue
      }
      urlText = token
    }
  } catch (error) {
    return { raw, root, fields: [], error: error instanceof Error ? error.message : String(error) }
  }

  if (!urlText) {
    return { raw, root, fields: [], error: 'cURL Sample Request is missing a URL.' }
  }

  let parsedURL: URL
  try {
    parsedURL = new URL(urlText)
  } catch {
    return { raw, root, fields: [], error: `cURL Sample Request URL is invalid: ${urlText}` }
  }

  if (dataParts.length > 1) {
    return { raw, root, fields: [], error: 'Only one JSON --data payload is supported for cURL Sample Request.' }
  }

  let body: unknown
  if (dataParts.length) {
    try {
      body = JSON.parse(dataParts[0])
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      return { raw, root, fields: [], error: `cURL --data must be valid JSON: ${message}` }
    }
  }

  const effectiveMethod = (method || (dataParts.length ? 'POST' : 'GET')).toUpperCase()
  const query = queryObjectFromURL(parsedURL)
  const parsed = {
    method: effectiveMethod,
    url: parsedURL.toString(),
    host: parsedURL.host,
    path: parsedURL.pathname || '/',
    query,
    headers,
    body,
  }

  const fieldMap = new Map<string, SampleRequestField>()
  addHTTPScalarField(fieldMap, 'request.method', effectiveMethod)
  addHTTPScalarField(fieldMap, 'request.host', parsedURL.host)
  addHTTPScalarField(fieldMap, 'request.path', parsedURL.pathname || '/')
  Object.entries(query).forEach(([key, values]) => {
    if (isSupportedPathKey(key)) addHTTPIndexedField(fieldMap, `request.query.${key}`, values)
  })
  Object.entries(headers).forEach(([key, values]) => {
    if (isSupportedPathKey(key)) addHTTPIndexedField(fieldMap, `request.headers.${key}`, values)
  })
  if (body !== undefined) {
    buildSampleRequestFields('request.body', body).forEach((field) => fieldMap.set(field.path, field))
  }

  return {
    raw,
    format: 'curl',
    root: 'request',
    fields: [...fieldMap.values()].sort((left, right) => left.path.localeCompare(right.path)),
    authoring: {
      format: 'curl',
      root: 'request',
      raw,
      parsed,
    },
  }
}

function queryObjectFromURL(parsedURL: URL): Record<string, string[]> {
  const query: Record<string, string[]> = {}
  parsedURL.searchParams.forEach((value, key) => {
    query[key] = [...(query[key] || []), value]
  })
  return query
}

function readOptionValue(tokens: string[], index: number, option: string) {
  const value = tokens[index]
  if (value === undefined) {
    throw new Error(`${option} requires a value.`)
  }
  return value
}

function looksLikeCurl(value: string) {
  return /^curl(?:\s|$)/.test(value)
}

function tokenizeShell(raw: string): { tokens: string[]; error?: string } {
  const input = raw.replace(/\\\r?\n/g, ' ')
  const tokens: string[] = []
  let current = ''
  let quote: '"' | "'" | '' = ''
  let escaping = false

  const pushCurrent = () => {
    if (current.length) {
      tokens.push(current)
      current = ''
    }
  }

  for (const char of input) {
    if (escaping) {
      current += char
      escaping = false
      continue
    }
    if (char === '\\' && quote !== "'") {
      escaping = true
      continue
    }
    if (quote) {
      if (char === quote) {
        quote = ''
      } else {
        current += char
      }
      continue
    }
    if (char === '"' || char === "'") {
      quote = char
      continue
    }
    if (/\s/.test(char)) {
      pushCurrent()
      continue
    }
    current += char
  }

  if (escaping) current += '\\'
  if (quote) {
    return { tokens: [], error: `cURL Sample Request has an unclosed ${quote} quote.` }
  }
  pushCurrent()
  return { tokens }
}
