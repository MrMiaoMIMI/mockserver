import type { Condition, MockEvent, Rule, RuleSet, SimulationResult } from '@/types'
import type { RuleDiagnosticState } from '@/utils/ruleCollection'

export interface SimulationDiagnosticMetric {
  label: string
  value: string
  tone?: 'ok' | 'warn' | 'muted'
}

export interface SimulationDiagnosticTrace {
  id: string
  category: 'ruleset' | 'selector' | 'condition'
  label: string
  matched: boolean
  message: string
  detail?: string
  depth?: number
}

export interface SimulationDiagnostics {
  matched: boolean
  statusLabel: string
  fallback: boolean
  fallbackReason: string
  responseStatus: string
  matchedRuleSet: string
  matchedRule: string
  candidateRules: string[]
  metrics: SimulationDiagnosticMetric[]
  rulesetTraces: SimulationDiagnosticTrace[]
  selectorTraces: SimulationDiagnosticTrace[]
  conditionTraces: SimulationDiagnosticTrace[]
}

export interface SimulationEventWarning {
  path: string
  message: string
}

type ExplainRecord = Record<string, any>
type StringConstraint = {
  field: string
  op: string
  values: string[]
  path: string
}

const defaultHost = 'demo.com'
const defaultPath = '/api/v1/debug'

export function buildDefaultSimulationEvent(ruleSet: RuleSet | null, rule?: Rule | null): MockEvent {
  if (ruleSet?.protocol === 'cache') {
    return buildDefaultCacheEvent(ruleSet, rule)
  }
  return buildDefaultHTTPEvent(ruleSet, rule)
}

function buildDefaultHTTPEvent(ruleSet: RuleSet | null, rule?: Rule | null): MockEvent {
  const selectorHost = selectorValue(ruleSet, 'request.host', ['eq', 'in'])
  const selectorPath = normalizePath(selectorValue(ruleSet, 'request.path', ['eq', 'prefix', 'in']) || defaultPath)
  const selectorConstraints = collectSelectorConstraints(ruleSet)
  const event: MockEvent = {
    protocol: ruleSet?.protocol || 'http',
    namespace: ruleSet?.namespace || 'default',
    request: {
      method: 'GET',
      host: selectorHost || defaultHost,
      path: selectorPath,
      query: {
        q1: ['qv1'],
      },
      headers: {
        accept: ['application/json'],
      },
    },
  }

  if (rule) {
    applySafeConditionHints(event, rule.when, {
      selectorConstraints,
    })
  }

  return event
}

function buildDefaultCacheEvent(ruleSet: RuleSet, rule?: Rule | null): MockEvent {
  const operation = (selectorValue(ruleSet, 'request.operation', ['eq', 'in']) || 'get').toLowerCase()
  const selectorKey = selectorValue(ruleSet, 'request.key', ['eq', 'prefix']) || 'user:'
  const key = keyFromSelectorValue(selectorKey)
  const event: MockEvent = {
    protocol: 'cache',
    namespace: ruleSet.namespace || 'default',
    request: {
      operation,
      key,
      ttl_ms: 3000,
      value: {
        id: key,
        status: 'active',
      },
    },
  }

  if (rule) {
    applySafeConditionHints(event, rule.when, {
      selectorConstraints: collectSelectorConstraints(ruleSet),
    })
  }

  return event
}

export function buildSimulationEventWarnings(
  ruleSet: RuleSet | null,
  rule?: Rule | null
): SimulationEventWarning[] {
  if (!ruleSet || !rule) return []
  const selectorConstraints = collectSelectorConstraints(ruleSet)
  if (!selectorConstraints.length) return []
  const ruleConstraints = collectConditionConstraints(rule.when, 'rule.when')
  const warnings: SimulationEventWarning[] = []
  for (const selectorConstraint of selectorConstraints) {
    for (const ruleConstraint of ruleConstraints) {
      if (selectorConstraint.field !== ruleConstraint.field) continue
      if (constraintsOverlap(selectorConstraint, ruleConstraint)) continue
      warnings.push({
        path: ruleConstraint.path,
        message: `Selected rule may be unreachable: selector ${formatConstraint(selectorConstraint)} conflicts with condition ${formatConstraint(ruleConstraint)}.`,
      })
      return warnings
    }
  }
  return warnings
}

export function buildSimulationDiagnostics(result: SimulationResult | null): SimulationDiagnostics | null {
  if (!result) return null

  const fallbackReason = result.trace?.fallback_reason || fallbackReasonFromResponse(result)
  const diagnostics: SimulationDiagnostics = {
    matched: result.matched,
    statusLabel: result.matched ? 'matched' : 'missed',
    fallback: Boolean(result.fallback || fallbackReason),
    fallbackReason: fallbackReason || '-',
    responseStatus: result.response?.status !== undefined ? String(result.response.status) : '-',
    matchedRuleSet: result.trace?.ruleset_id || '-',
    matchedRule: result.trace?.rule_id || '-',
    candidateRules: result.candidates?.length
      ? result.candidates
      : candidateRulesFromExplain(result.explain as ExplainRecord),
    metrics: [],
    rulesetTraces: [],
    selectorTraces: [],
    conditionTraces: [],
  }

  diagnostics.metrics = [
    { label: 'matched', value: diagnostics.statusLabel, tone: result.matched ? 'ok' : 'warn' },
    { label: 'ruleset', value: diagnostics.matchedRuleSet, tone: diagnostics.matchedRuleSet === '-' ? 'muted' : 'ok' },
    { label: 'rule', value: diagnostics.matchedRule, tone: diagnostics.matchedRule === '-' ? 'muted' : 'ok' },
    { label: 'fallback', value: diagnostics.fallback ? 'yes' : 'no', tone: diagnostics.fallback ? 'warn' : 'ok' },
    { label: 'reason', value: diagnostics.fallbackReason, tone: diagnostics.fallback ? 'warn' : 'muted' },
    { label: 'status', value: diagnostics.responseStatus, tone: diagnostics.responseStatus === '-' ? 'muted' : 'ok' },
    { label: 'candidates', value: String(diagnostics.candidateRules.length), tone: diagnostics.candidateRules.length ? 'ok' : 'muted' },
  ]

  diagnostics.rulesetTraces = rulesetTracesFromExplain(result.explain as ExplainRecord)
  diagnostics.selectorTraces = selectorTracesFromExplain(result.explain as ExplainRecord)
  diagnostics.conditionTraces = conditionTracesFromExplain(result.explain as ExplainRecord)

  return diagnostics
}

export function buildSimulationRuleDiagnostics(
  result: SimulationResult | null,
  rules: Rule[] = []
): Record<string, RuleDiagnosticState> {
  if (!result) return {}

  const matchedRuleId = result.trace?.rule_id || ''
  const candidateRuleIds = new Set(
    (result.candidates?.length ? result.candidates : candidateRulesFromExplain(result.explain as ExplainRecord)).map(String)
  )

  return rules.reduce<Record<string, RuleDiagnosticState>>((diagnostics, rule) => {
    if (matchedRuleId === rule.id) {
      diagnostics[rule.id] = {
        simulation: result.fallback ? 'fallback' : 'matched',
        messages: [result.fallback ? 'latest simulation reached fallback from this rule' : 'latest simulation matched this rule'],
      }
      return diagnostics
    }
    if (!candidateRuleIds.size && selectorMissed(result.explain as ExplainRecord)) {
      diagnostics[rule.id] = {
        simulation: 'not_reached',
        messages: ['latest simulation stopped at the ruleset selector before evaluating rules'],
      }
      return diagnostics
    }
    if (candidateRuleIds.has(rule.id)) {
      diagnostics[rule.id] = {
        simulation: 'candidate',
        messages: ['latest simulation considered this rule as a candidate'],
      }
      return diagnostics
    }
    diagnostics[rule.id] = {
      simulation: 'missed',
      messages: ['latest simulation did not match this rule'],
    }
    return diagnostics
  }, {})
}

function applySafeConditionHints(
  event: MockEvent,
  condition: Condition,
  context: {
    selectorConstraints: StringConstraint[]
  }
) {
  if (condition.all) {
    condition.all.forEach((child) => applySafeConditionHints(event, child, context))
    return
  }
  if (condition.any || condition.not || condition.expr !== undefined) return
  applyPredicateHint(event, condition, context)
}

function applyPredicateHint(
  event: MockEvent,
  condition: Condition,
  context: {
    selectorConstraints: StringConstraint[]
  }
) {
  const value = firstStringValue(condition.value)
  if (!value) return

  if (condition.field === 'request.method' && condition.op === 'eq') {
    const method = value.toUpperCase()
    if (hintCanReplaceSelectorValue(condition.field, condition.op, method, context.selectorConstraints)) {
      event.request.method = method
    }
    return
  }

  if (condition.field === 'request.host' && condition.op === 'eq') {
    const host = value.toLowerCase()
    if (hintCanReplaceSelectorValue(condition.field, condition.op, host, context.selectorConstraints)) {
      event.request.host = host
    }
    return
  }

  if (condition.field === 'request.path' && ['eq', 'prefix'].includes(condition.op || '')) {
    const path = normalizePath(value)
    if (hintCanReplaceSelectorValue(condition.field, condition.op || 'eq', path, context.selectorConstraints)) {
      event.request.path = path
    }
    return
  }

  const queryKey = indexedFieldKey(condition.field || '', 'request.query.')
  if (queryKey && condition.op === 'eq') {
    event.request.query = {
      ...event.request.query,
      [queryKey]: [value],
    }
    return
  }

  if (condition.field === 'request.operation' && condition.op === 'eq') {
    const operation = value.toLowerCase()
    if (hintCanReplaceSelectorValue(condition.field, condition.op, operation, context.selectorConstraints)) {
      event.request.operation = operation
    }
    return
  }

  if (condition.field === 'request.key' && ['eq', 'prefix'].includes(condition.op || '')) {
    const key = keyFromSelectorValue(value)
    if (hintCanReplaceSelectorValue(condition.field, condition.op || 'eq', key, context.selectorConstraints)) {
      event.request.key = key
    }
    return
  }

  const headerKey = indexedFieldKey(condition.field || '', 'request.headers.')
  if (headerKey && condition.op === 'eq') {
    event.request.headers = {
      ...event.request.headers,
      [headerKey.toLowerCase()]: [value],
    }
  }
}

function rulesetTracesFromExplain(explain?: ExplainRecord): SimulationDiagnosticTrace[] {
  const ruleSets = Array.isArray(explain?.rule_set_explanations) ? explain.rule_set_explanations : []
  return ruleSets.map((ruleSet: ExplainRecord, index: number) => ({
    id: `${ruleSet.ruleset_id || 'ruleset'}-outcome-${index}`,
    category: 'ruleset' as const,
    label: String(ruleSet.ruleset_id || '-'),
    matched: Boolean(ruleSet.matched),
    message: ruleSet.message || (ruleSet.matched ? 'ruleset matched' : 'ruleset missed'),
    detail: Array.isArray(ruleSet.candidate_rules)
      ? `${ruleSet.candidate_rules.length} candidate rule${ruleSet.candidate_rules.length === 1 ? '' : 's'}`
      : undefined,
  }))
}

function selectorTracesFromExplain(explain?: ExplainRecord): SimulationDiagnosticTrace[] {
  const ruleSets = Array.isArray(explain?.rule_set_explanations) ? explain.rule_set_explanations : []
  return ruleSets.flatMap((ruleSet: ExplainRecord) => {
    const checks = Array.isArray(ruleSet.selector_checks) ? ruleSet.selector_checks : []
    return checks.map((check: ExplainRecord, index: number) => ({
      id: `${ruleSet.ruleset_id || 'ruleset'}-selector-${check.name || index}`,
      category: 'selector' as const,
      label: `${ruleSet.ruleset_id || '-'} / ${check.name || 'selector'}`,
      matched: Boolean(check.matched),
      message: check.message || (check.matched ? 'matched' : 'not matched'),
      detail: formatExpectedActual(check.expected, check.actual),
    }))
  })
}

function conditionTracesFromExplain(explain?: ExplainRecord): SimulationDiagnosticTrace[] {
  const result: SimulationDiagnosticTrace[] = []
  const rules = ruleExplanationsFromExplain(explain)
  rules.forEach((rule: ExplainRecord) => {
    flattenConditionTrace(rule.condition, result, {
      prefix: rule.rule_id || 'rule',
      label: rule.rule_id || '-',
      depth: 0,
    })
  })
  return result
}

function ruleExplanationsFromExplain(explain?: ExplainRecord): ExplainRecord[] {
  const rules = new Map<string, ExplainRecord>()
  const appendRule = (rule: ExplainRecord) => {
    const key = String(rule.rule_id || `${rules.size}`)
    if (!rules.has(key)) {
      rules.set(key, rule)
    }
  }
  if (Array.isArray(explain?.rule_explanations)) {
    explain.rule_explanations.forEach(appendRule)
  }
  const ruleSets = Array.isArray(explain?.rule_set_explanations) ? explain.rule_set_explanations : []
  ruleSets.forEach((ruleSet: ExplainRecord) => {
    if (Array.isArray(ruleSet.rule_explanations)) {
      ruleSet.rule_explanations.forEach(appendRule)
    }
  })
  return Array.from(rules.values())
}

function flattenConditionTrace(
  condition: ExplainRecord | undefined,
  result: SimulationDiagnosticTrace[],
  context: {
    prefix: string
    label: string
    depth: number
  }
) {
  if (!condition) return
  const label = condition.field || condition.expr || condition.kind || 'condition'
  result.push({
    id: `${context.prefix}-condition-${result.length}`,
    category: 'condition',
    label: `${context.label} / ${label}`,
    matched: Boolean(condition.matched),
    message: condition.message || (condition.matched ? 'matched' : 'not matched'),
    detail: condition.field
      ? formatExpectedActual(condition.expected, condition.actual)
      : condition.expr || undefined,
    depth: context.depth,
  })
  const children = Array.isArray(condition.children) ? condition.children : []
  children.forEach((child: ExplainRecord) => {
    flattenConditionTrace(child, result, {
      ...context,
      depth: context.depth + 1,
    })
  })
}

function candidateRulesFromExplain(explain?: ExplainRecord) {
  if (Array.isArray(explain?.candidate_rules)) return explain.candidate_rules.map(String)
  const ruleSets = Array.isArray(explain?.rule_set_explanations) ? explain.rule_set_explanations : []
  const candidates = ruleSets.flatMap((ruleSet: ExplainRecord) => {
    return Array.isArray(ruleSet.candidate_rules) ? ruleSet.candidate_rules : []
  })
  return candidates.map(String)
}

function selectorMissed(explain?: ExplainRecord) {
  const ruleSets = Array.isArray(explain?.rule_set_explanations) ? explain.rule_set_explanations : []
  return ruleSets.some((ruleSet: ExplainRecord) => {
    const checks = Array.isArray(ruleSet.selector_checks) ? ruleSet.selector_checks : []
    return checks.some((check: ExplainRecord) => check.matched === false)
  })
}

function fallbackReasonFromResponse(result: SimulationResult) {
  const headers = result.response?.headers || {}
  return headers['x-mockserver-fallback']?.[0] || headers['X-Mockserver-Fallback']?.[0] || ''
}

function formatExpectedActual(expected: unknown, actual: unknown) {
  const left = expected === undefined ? '' : `expected ${formatInline(expected)}`
  const right = actual === undefined ? '' : `actual ${formatInline(actual)}`
  return [left, right].filter(Boolean).join(' / ')
}

function indexedFieldKey(field: string, prefix: string) {
  if (!field.startsWith(prefix)) return ''
  return field.slice(prefix.length).replace(/\[(?:\*|\d+)]$/, '')
}

function firstStringValue(value: unknown) {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return ''
}

function selectorValue(ruleSet: RuleSet | null, field: string, operators: string[]): string {
  return firstStringValue(
    ruleSet?.selector.all?.find(
      (condition) => condition.field === field && operators.includes(condition.op || '')
    )?.value
  )
}

function selectorValues(ruleSet: RuleSet | null, field: string, operators: string[]): string[] {
  return (ruleSet?.selector.all || [])
    .filter((condition) => condition.field === field && operators.includes(condition.op || ''))
    .map((condition) => firstStringValue(condition.value))
    .filter(Boolean)
}

function collectSelectorConstraints(ruleSet: RuleSet | null): StringConstraint[] {
  return collectAllConditionConstraints(ruleSet?.selector.all || [], 'selector.all')
}

function collectAllConditionConstraints(conditions: Condition[], path: string): StringConstraint[] {
  return conditions.flatMap((condition, index) => collectConditionConstraints(condition, `${path}[${index}]`))
}

function collectConditionConstraints(condition: Condition, path: string): StringConstraint[] {
  if (condition.all?.length) {
    return collectAllConditionConstraints(condition.all, `${path}.all`)
  }
  if (condition.any?.length || condition.not || condition.expr !== undefined) return []
  if (!condition.field || !['eq', 'in', 'prefix'].includes(condition.op || '')) return []
  const values = normalizeConstraintValues(condition.field, valueStrings(condition.value))
  if (!values.length) return []
  return [
    {
      field: condition.field,
      op: condition.op || 'eq',
      values,
      path,
    },
  ]
}

function valueStrings(value: unknown): string[] {
  if (typeof value === 'string') return [value]
  if (Array.isArray(value)) return value.filter((item): item is string => typeof item === 'string')
  return []
}

function normalizeConstraintValues(field: string, values: string[]) {
  return values
    .map((value) => normalizeConstraintValue(field, value))
    .filter(Boolean)
}

function normalizeConstraintValue(field: string, value: string) {
  const trimmed = value.trim()
  if (!trimmed) return ''
  if (field === 'request.method') return trimmed.toUpperCase()
  if (field === 'request.host' || field === 'request.original_host') return trimmed.toLowerCase()
  if (field === 'request.operation') return trimmed.toLowerCase()
  if (field === 'request.path') return normalizePath(trimmed)
  return trimmed
}

function hintCanReplaceSelectorValue(
  field: string,
  op: string,
  value: string,
  selectorConstraints: StringConstraint[]
) {
  const constraints = selectorConstraints.filter((constraint) => constraint.field === field)
  if (!constraints.length) return true
  const normalized = normalizeConstraintValue(field, value)
  if (!normalized) return false

  return constraints.every((constraint) => {
    const exactValues = exactConstraintValues(constraint)
    if (exactValues) {
      if (op === 'eq') return exactValues.includes(normalized)
      return false
    }
    if (constraint.op === 'prefix') {
      if (op === 'eq') {
        return constraint.values.some((prefix) =>
          prefixConstraintMatches(field, normalized, prefix)
        )
      }
      if (op === 'prefix') {
        return constraint.values.some((prefix) =>
          prefixConstraintMatches(field, normalized, prefix)
        )
      }
    }
    return true
  })
}

function constraintsOverlap(left: StringConstraint, right: StringConstraint) {
  const leftExact = exactConstraintValues(left)
  const rightExact = exactConstraintValues(right)
  if (leftExact && rightExact) return leftExact.some((value) => rightExact.includes(value))
  if (leftExact && right.op === 'prefix') {
    return leftExact.some((value) =>
      right.values.some((prefix) => prefixConstraintMatches(right.field, value, prefix))
    )
  }
  if (left.op === 'prefix' && rightExact) {
    return rightExact.some((value) =>
      left.values.some((prefix) => prefixConstraintMatches(left.field, value, prefix))
    )
  }
  if (left.op === 'prefix' && right.op === 'prefix') {
    return left.values.some((leftPrefix) =>
      right.values.some(
        (rightPrefix) =>
          prefixConstraintMatches(left.field, leftPrefix, rightPrefix) ||
          prefixConstraintMatches(left.field, rightPrefix, leftPrefix)
      )
    )
  }
  return true
}

function prefixConstraintMatches(field: string, value: string, prefix: string) {
  if (field !== 'request.path') return value.startsWith(prefix)
  return pathPrefixMatches(value, prefix)
}

function pathPrefixMatches(path: string, prefix: string) {
  if (!prefix) return true
  if (!path.startsWith(prefix)) return false
  if (prefix === '/' || path === prefix || prefix.endsWith('/')) return true
  return path.startsWith(`${prefix}/`)
}

function exactConstraintValues(constraint: StringConstraint) {
  if (constraint.op === 'eq' || constraint.op === 'in') return constraint.values
  return null
}

function formatConstraint(constraint: StringConstraint) {
  const value = constraint.values.length === 1 ? constraint.values[0] : `[${constraint.values.join(', ')}]`
  return `${constraint.field} ${constraint.op} ${value}`
}

function keyFromSelectorValue(value: string): string {
  if (!value) return 'user:123'
  return value.endsWith(':') || value.endsWith('/') || value.endsWith('-') ? `${value}sample` : value
}

function normalizePath(value: string) {
  if (!value) return defaultPath
  return value.startsWith('/') ? value : `/${value}`
}

function formatInline(value: unknown) {
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}
