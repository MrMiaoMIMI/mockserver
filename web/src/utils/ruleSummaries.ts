import type { Condition, Rule, RuleAction } from '@/types'

export function conditionSummary(condition: Condition): string {
  if (condition.expr !== undefined) return `CEL ${truncateValue(condition.expr || 'empty expression')}`
  if (condition.all) return groupConditionSummary('ALL', condition.all)
  if (condition.any) return groupConditionSummary('ANY', condition.any)
  if (condition.not) return `NOT (${conditionSummary(condition.not)})`
  return predicateSummary(condition)
}

export function actionSummary(action: RuleAction): string {
  const status = action.status ? `HTTP ${action.status}` : ''
  if (action.type === 'static_response') {
    return ['static response', status, actionBodySummary(action.body)].filter(Boolean).join(' / ')
  }
  if (action.type === 'template_response') {
    return ['template response', status, truncateValue(action.body_template || 'empty template')]
      .filter(Boolean)
      .join(' / ')
  }
  if (action.type === 'cel_response') {
    return ['CEL response', status, truncateValue(action.body_expression || 'empty expression')]
      .filter(Boolean)
      .join(' / ')
  }
  if (action.type === 'sequence_response') {
    const strategy = action.sequence_strategy || 'last'
    return `sequence response / ${(action.sequence || []).length} steps / ${strategy}`
  }
  if (action.type === 'webhook_response') {
    const method = action.webhook?.method || 'POST'
    const url = truncateValue(action.webhook?.url || 'missing url')
    const timeout = action.webhook?.timeout_ms ? `${action.webhook.timeout_ms}ms` : ''
    return ['webhook response', method, url, timeout].filter(Boolean).join(' / ')
  }
  return [action.type || 'unknown action', status].filter(Boolean).join(' / ')
}

export function actionTypeLabel(actionType: string) {
  if (actionType === 'static_response') return 'static'
  if (actionType === 'template_response') return 'template'
  if (actionType === 'cel_response') return 'CEL'
  if (actionType === 'sequence_response') return 'sequence'
  if (actionType === 'webhook_response') return 'webhook'
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

function actionBodySummary(body: unknown) {
  if (body === undefined) return ''
  if (body === null) return 'null body'
  if (Array.isArray(body)) return `${body.length} item body`
  if (typeof body === 'object') return `${Object.keys(body as Record<string, unknown>).length} field body`
  return truncateValue(String(body))
}
