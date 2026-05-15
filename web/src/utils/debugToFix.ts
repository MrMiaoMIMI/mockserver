import type { Condition, MockEvent, Rule, RuleSet } from '@/types'
import type { RuntimeDiagnosticRow, RuntimeTone } from '@/utils/runtimeDiagnostics'

export type DebugToFixAction = 'simulate' | 'create_rule'
export type DebugToFixTargetReason = 'runtime_match' | 'namespace_protocol'

export interface DebugToFixTarget {
  rulesetId: string
  reason: DebugToFixTargetReason
  label: string
}

export interface DebugToFixRequestFacts {
  recordId: number
  observedAt: string
  namespace: string
  method: string
  host: string
  path: string
  traceId: string
  outcome: string
  fallbackReason: string
  status: string
  duration: string
}

export interface DebugToFixPayload {
  id: string
  source: 'runtime-request'
  action: DebugToFixAction
  targetRulesetId: string
  targetReason: DebugToFixTargetReason
  createdAt: string
  request: DebugToFixRequestFacts
  event: MockEvent
}

export interface RuntimeDiagnosis {
  title: string
  detail: string
  nextAction: string
  tone: RuntimeTone
}

export const debugToFixStoragePrefix = 'mockserver.debugToFix.'

export function inferDebugToFixTarget(
  row: RuntimeDiagnosticRow,
  drafts: RuleSet[]
): DebugToFixTarget | null {
  if (row.record.ruleset_id) {
    return {
      rulesetId: row.record.ruleset_id,
      reason: 'runtime_match',
      label: `runtime matched ${row.record.ruleset_id}`,
    }
  }

  const namespace = row.record.namespace || row.record.event?.namespace || 'default'
  const protocol = row.record.event?.protocol || 'http'
  const candidates = drafts
    .filter((draft) => draft.namespace === namespace && draft.protocol === protocol)
    .sort((left, right) => {
      if (left.enabled !== right.enabled) return left.enabled ? -1 : 1
      return right.rules.length - left.rules.length || left.name.localeCompare(right.name)
    })

  const candidate = candidates[0]
  if (!candidate) return null
  return {
    rulesetId: candidate.id,
    reason: 'namespace_protocol',
    label: `best ${namespace}/${protocol} draft`,
  }
}

export function buildRuntimeDiagnosis(row: RuntimeDiagnosticRow): RuntimeDiagnosis {
  if (row.outcome === 'matched') {
    return {
      title: 'Matched a rule',
      detail: `Runtime selected ${row.record.ruleset_id || 'a ruleset'} / ${row.record.rule_id || 'a rule'} and returned ${row.statusLabel}.`,
      nextAction: 'Open the owning rule or replay the request against the current draft.',
      tone: 'ok',
    }
  }

  if (row.outcome === 'error') {
    return {
      title: 'Runtime error',
      detail: row.record.message || `Runtime returned ${row.statusLabel}. Inspect the captured event and replay it against a draft.`,
      nextAction: 'Use the request as simulation input before editing rules.',
      tone: 'danger',
    }
  }

  if (row.record.fallback_reason === 'ruleset_miss') {
    return {
      title: 'Ruleset selector miss',
      detail: 'No published ruleset selector accepted this request. Check namespace, protocol, host, path, and selector conditions.',
      nextAction: 'Replay the request in the closest draft or create a request-derived rule after selector coverage is confirmed.',
      tone: 'warn',
    }
  }

  if (row.record.fallback_reason === 'rule_miss') {
    return {
      title: 'Rule condition miss',
      detail: 'A ruleset was reached, but no rule condition matched this request.',
      nextAction: 'Create a rule from this request or replay it to inspect condition misses.',
      tone: 'warn',
    }
  }

  if (row.outcome === 'fallback') {
    return {
      title: 'Fallback response',
      detail: `Runtime used fallback${row.record.fallback_reason ? `: ${row.record.fallback_reason}` : ''}.`,
      nextAction: 'Replay the event and compare selector and condition results.',
      tone: 'warn',
    }
  }

  return {
    title: 'No matching mock',
    detail: 'The request did not resolve to a mock rule.',
    nextAction: 'Use the request as a simulation event and create a matching rule if needed.',
    tone: 'neutral',
  }
}

export function createDebugToFixPayload(
  row: RuntimeDiagnosticRow,
  action: DebugToFixAction,
  target: DebugToFixTarget,
  now = new Date()
): DebugToFixPayload | null {
  if (!row.record.event) return null
  const createdAt = now.toISOString()
  return {
    id: `runtime-${row.record.id}-${now.getTime()}`,
    source: 'runtime-request',
    action,
    targetRulesetId: target.rulesetId,
    targetReason: target.reason,
    createdAt,
    request: {
      recordId: row.record.id,
      observedAt: row.record.observed_at,
      namespace: row.record.namespace || row.record.event.namespace || 'default',
      method: row.record.method || String(row.record.event.request.method || '-'),
      host: row.record.host || String(row.record.event.request.host || '-'),
      path: row.pathLabel,
      traceId: row.record.trace_id || row.record.event.meta?.trace_id || '',
      outcome: row.outcomeLabel,
      fallbackReason: row.record.fallback_reason || '',
      status: row.statusLabel,
      duration: row.durationLabel,
    },
    event: row.record.event,
  }
}

export function storeDebugToFixPayload(payload: DebugToFixPayload, storage: Storage = sessionStorage) {
  storage.setItem(storageKey(payload.id), JSON.stringify(payload))
}

export function loadDebugToFixPayload(id: string, storage: Storage = sessionStorage) {
  const raw = storage.getItem(storageKey(id))
  if (!raw) return null
  try {
    return JSON.parse(raw) as DebugToFixPayload
  } catch {
    return null
  }
}

export function clearDebugToFixPayload(id: string, storage: Storage = sessionStorage) {
  storage.removeItem(storageKey(id))
}

export function buildDebugToFixRoute(payload: DebugToFixPayload) {
  return {
    path: `/rulesets/${encodeURIComponent(payload.targetRulesetId)}/rules`,
    query: {
      workbench: payload.action === 'create_rule' ? 'editor' : 'simulate',
      debug: payload.id,
    },
  }
}

export function buildRuleSeedFromDebugPayload(
  payload: DebugToFixPayload,
  existingRules: Rule[] = []
): Rule {
  return {
    id: nextRuntimeRuleId(payload.request.recordId, existingRules),
    name: `${payload.request.method || 'Request'} ${payload.request.path || 'runtime request'}`.slice(0, 96),
    enabled: true,
    priority: nextPriority(existingRules),
    when: buildRuntimeCondition(payload.event),
    action: {
      type: 'respond',
      renderer: 'static',
      response: {
        payload: {
          status: 200,
          headers: {
            'content-type': ['application/json'],
          },
          body: {
            source: 'runtime-debug',
            request_id: payload.request.recordId,
            trace_id: payload.request.traceId || undefined,
            message: 'mock response generated from runtime request',
          },
        },
      },
    },
  }
}

export function debugPayloadSummary(payload: DebugToFixPayload) {
  const request = payload.request
  return [
    `#${request.recordId}`,
    request.method,
    request.host,
    request.path,
    request.traceId ? `trace ${request.traceId}` : '',
  ].filter(Boolean).join(' / ')
}

function storageKey(id: string) {
  return `${debugToFixStoragePrefix}${id}`
}

function buildRuntimeCondition(event: MockEvent): Condition {
  const conditions: Condition[] = []
  const request = event.request || {}

  pushPredicate(conditions, 'request.method', 'eq', normalizeString(request.method).toUpperCase())
  pushPredicate(conditions, 'request.host', 'eq', normalizeString(request.host).toLowerCase())
  pushPredicate(conditions, 'request.path', 'eq', normalizeString(request.path))

  Object.entries(request.query || {}).slice(0, 4).forEach(([key, values]) => {
    const value = Array.isArray(values) ? values[0] : ''
    pushPredicate(conditions, `request.query.${key}[0]`, 'eq', value)
  })

  if (!conditions.length) {
    pushPredicate(conditions, 'request.path', 'eq', '/')
  }

  return conditions.length === 1 ? conditions[0] : { all: conditions }
}

function pushPredicate(conditions: Condition[], field: string, op: string, value: unknown) {
  if (value === undefined || value === null || value === '') return
  conditions.push({ field, op, value })
}

function normalizeString(value: unknown) {
  return typeof value === 'string' ? value.trim() : ''
}

function nextRuntimeRuleId(recordId: number, existingRules: Rule[]) {
  const existingIds = new Set(existingRules.map((rule) => rule.id))
  const base = `runtime-${recordId}`
  if (!existingIds.has(base)) return base
  for (let index = 2; index < 1000; index += 1) {
    const id = `${base}-${index}`
    if (!existingIds.has(id)) return id
  }
  return `${base}-${Date.now()}`
}

function nextPriority(existingRules: Rule[]) {
  const max = existingRules.reduce((result, rule) => Math.max(result, rule.priority || 0), 0)
  return max + 10
}
