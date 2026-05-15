import type { Condition, Rule, RuleAction } from '@/types'

export function conditionSummary(condition: Condition): string {
  if (condition.expr !== undefined) return `CEL ${truncateValue(condition.expr || 'empty expression')}`
  if (condition.all) return groupConditionSummary('ALL', condition.all)
  if (condition.any) return groupConditionSummary('ANY', condition.any)
  if (condition.not) return `NOT (${conditionSummary(condition.not)})`
  return predicateSummary(condition)
}

export function actionSummary(action: RuleAction): string {
  const renderer = action.renderer || action.type || 'static'
  if (renderer === 'static') {
    return ['static response', responsePayloadLabel(action.response?.payload), payloadSummary(action.response?.payload)]
      .filter(Boolean)
      .join(' / ')
  }
  if (renderer === 'template') {
    return ['template response', truncateValue(action.response_template || 'empty template')]
      .filter(Boolean)
      .join(' / ')
  }
  if (renderer === 'cel') {
    return ['CEL response', truncateValue(action.response_expression || 'empty expression')]
      .filter(Boolean)
      .join(' / ')
  }
  if (renderer === 'sequence') {
    const strategy = action.sequence_strategy || 'last'
    return `sequence response / ${(action.sequence || []).length} steps / ${strategy}`
  }
  if (renderer === 'webhook') {
    const method = action.webhook?.method || 'POST'
    const url = truncateValue(action.webhook?.url || 'missing url')
    const timeout = action.webhook?.timeout_ms ? `${action.webhook.timeout_ms}ms` : ''
    return ['webhook response', method, url, timeout].filter(Boolean).join(' / ')
  }
  return [action.type || 'unknown action', payloadSummary(action.response?.payload)].filter(Boolean).join(' / ')
}

export function actionTypeLabel(actionType: string) {
  if (actionType === 'static') return 'static'
  if (actionType === 'template') return 'template'
  if (actionType === 'cel') return 'CEL'
  if (actionType === 'sequence') return 'sequence'
  if (actionType === 'webhook') return 'webhook'
  return actionType || 'unknown'
}

export function ruleDisplayName(rule: Rule | null | undefined) {
  return rule?.name?.trim() || 'Untitled rule'
}

export function ruleTechnicalLabel(rule: Rule | null | undefined) {
  return rule?.id ? `id: ${rule.id}` : 'id pending'
}

export function cloneRule(rule: Rule): Rule {
  return JSON.parse(JSON.stringify(rule)) as Rule
}

export function nextDuplicateRuleId(ruleId: string, rules: Rule[]) {
  const base = `${ruleId}-copy`
  const existing = new Set(rules.map((rule) => rule.id))
  if (!existing.has(base)) return base
  for (let index = 2; index < 1000; index += 1) {
    const candidate = `${base}-${index}`
    if (!existing.has(candidate)) return candidate
  }
  return `${base}-${Date.now().toString().slice(-6)}`
}

export function nextDuplicatePriority(priority: number, rules: Rule[]) {
  const priorities = rules.map((rule) => rule.priority)
  let candidate = priority + 1
  while (priorities.includes(candidate)) {
    candidate += 1
  }
  return candidate
}

export function formatInlineValue(value: unknown) {
  if (value === undefined) return ''
  if (typeof value === 'string') return truncateValue(value)
  return truncateValue(JSON.stringify(value))
}

export function truncateValue(value: string, max = 72) {
  return value.length > max ? `${value.slice(0, max - 1)}...` : value
}

function groupConditionSummary(kind: string, children: Condition[]) {
  const count = children.length
  const preview = children.slice(0, 2).map((child) => conditionSummary(child))
  const suffix = count > preview.length ? ` +${count - preview.length} more` : ''
  return `${kind} ${count}: ${preview.join(' ; ')}${suffix}`
}

function predicateSummary(condition: Condition) {
  const field = condition.field || 'field'
  const op = condition.op || 'eq'
  const value = formatInlineValue(condition.value)
  return value ? `${field} ${op} ${value}` : `${field} ${op}`
}

function payloadSummary(payload: unknown) {
  if (payload === undefined) return ''
  if (payload === null) return 'null payload'
  if (Array.isArray(payload)) return `${payload.length} item payload`
  if (typeof payload === 'object') return `${Object.keys(payload as Record<string, unknown>).length} field payload`
  return truncateValue(String(payload))
}

function responsePayloadLabel(payload?: Record<string, unknown>) {
  if (!payload) return ''
  if (payload.status !== undefined) return `HTTP ${payload.status}`
  if (payload.code !== undefined) return `SPEX ${payload.code}`
  if (payload.hit !== undefined) return `cache ${payload.hit ? 'hit' : 'miss'}`
  return ''
}
