import type { Rule, RuleSet } from '@/types'
import { actionSummary, actionTypeLabel, conditionSummary } from '@/utils/ruleSummaries'

export type RuleEnabledFilter = 'all' | 'enabled' | 'disabled'
export type RuleStatusFilter =
  | 'all'
  | 'matched'
  | 'missed'
  | 'not_reached'
  | 'candidate'
  | 'fallback'
  | 'invalid'
export type RuleStatusTone = 'neutral' | 'ok' | 'warn' | 'caution' | 'danger' | 'accent'

export interface RuleFilterState {
  query: string
  enabled: RuleEnabledFilter
  actionType: string
  status: RuleStatusFilter
}

export interface RuleStatusTag {
  key: string
  label: string
  tone: RuleStatusTone
  title?: string
}

export interface RuleDiagnosticState {
  simulation?: 'matched' | 'missed' | 'not_reached' | 'candidate' | 'fallback'
  validation?: 'invalid'
  messages?: string[]
}

export interface RuleRowView {
  rule: Rule
  selected: boolean
  condition: string
  action: string
  actionType: string
  statuses: RuleStatusTag[]
  searchableText: string
}

export function defaultRuleFilters(): RuleFilterState {
  return {
    query: '',
    enabled: 'all',
    actionType: '',
    status: 'all',
  }
}

export function orderedRules(rules: Rule[]) {
  return [...rules].sort((left, right) => {
    if (left.priority !== right.priority) return left.priority - right.priority
    return left.id.localeCompare(right.id)
  })
}

export function buildRuleRows(
  ruleSet: RuleSet | null,
  filters: RuleFilterState,
  selectedRuleId = '',
  diagnostics: Record<string, RuleDiagnosticState> = {}
): RuleRowView[] {
  return orderedRules(ruleSet?.rules || [])
    .map((rule) => buildRuleRow(rule, selectedRuleId, diagnostics[rule.id]))
    .filter((row) => rowMatchesFilters(row, filters))
}

export function buildRuleRow(
  rule: Rule,
  selectedRuleId = '',
  diagnostic?: RuleDiagnosticState
): RuleRowView {
  const condition = conditionSummary(rule.when)
  const action = actionSummary(rule.action)
  const statuses = buildRuleStatusTags(rule, selectedRuleId, diagnostic)
  const searchableText = [
    rule.id,
    rule.name,
    String(rule.priority),
    condition,
    action,
    ruleActionKind(rule),
    actionTypeLabel(ruleActionKind(rule)),
    ...statuses.map((status) => status.label),
    ...(diagnostic?.messages || []),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  return {
    rule,
    selected: selectedRuleId === rule.id,
    condition,
    action,
    actionType: actionTypeLabel(ruleActionKind(rule)),
    statuses,
    searchableText,
  }
}

export function buildRuleStatusCounts(rows: RuleRowView[]) {
  return rows.reduce(
    (result, row) => {
      result.total += 1
      if (row.rule.enabled) result.enabled += 1
      else result.disabled += 1
      if (row.statuses.some((status) => status.key === 'invalid')) result.invalid += 1
      if (row.statuses.some((status) => status.key === 'matched')) result.matched += 1
      if (row.statuses.some((status) => status.key === 'missed')) result.missed += 1
      return result
    },
    { total: 0, enabled: 0, disabled: 0, invalid: 0, matched: 0, missed: 0 }
  )
}

export function availableActionTypes(ruleSet: RuleSet | null) {
  return Array.from(new Set((ruleSet?.rules || []).map(ruleActionKind).filter(Boolean))).sort()
}

export function resolveStableSelectedRuleId(
  ruleSet: RuleSet | null,
  preferredRuleId = '',
  visibleRows?: RuleRowView[]
) {
  const allRules = orderedRules(ruleSet?.rules || [])
  if (!allRules.length) return ''
  if (preferredRuleId && allRules.some((rule) => rule.id === preferredRuleId)) return preferredRuleId
  return visibleRows?.[0]?.rule.id || allRules[0].id
}

export function mergeRuleDiagnostics(
  ...sources: Array<Record<string, RuleDiagnosticState> | undefined>
) {
  return sources.reduce<Record<string, RuleDiagnosticState>>((result, source) => {
    Object.entries(source || {}).forEach(([ruleId, diagnostic]) => {
      const current = result[ruleId] || {}
      result[ruleId] = {
        simulation: diagnostic.simulation || current.simulation,
        validation: diagnostic.validation || current.validation,
        messages: [...(current.messages || []), ...(diagnostic.messages || [])],
      }
    })
    return result
  }, {})
}

function buildRuleStatusTags(
  rule: Rule,
  selectedRuleId: string,
  diagnostic?: RuleDiagnosticState
): RuleStatusTag[] {
  const statuses: RuleStatusTag[] = []
  if (selectedRuleId === rule.id) statuses.push({ key: 'selected', label: 'selected', tone: 'accent' })
  statuses.push({
    key: rule.enabled ? 'enabled' : 'disabled',
    label: rule.enabled ? 'enabled' : 'disabled',
    tone: rule.enabled ? 'ok' : 'neutral',
  })
  statuses.push({
    key: `action:${ruleActionKind(rule) || 'unknown'}`,
    label: actionTypeLabel(ruleActionKind(rule)),
    tone: 'neutral',
  })
  if (diagnostic?.validation === 'invalid') {
    statuses.push({
      key: 'invalid',
      label: 'invalid',
      tone: 'danger',
      title: diagnostic.messages?.join('\n'),
    })
  }
  if (diagnostic?.simulation === 'matched') {
    statuses.push({ key: 'matched', label: 'matched', tone: 'ok' })
  } else if (diagnostic?.simulation === 'fallback') {
    statuses.push({ key: 'fallback', label: 'fallback', tone: 'warn' })
  } else if (diagnostic?.simulation === 'missed') {
    statuses.push({ key: 'missed', label: 'missed', tone: 'warn' })
  } else if (diagnostic?.simulation === 'not_reached') {
    statuses.push({ key: 'not_reached', label: 'not reached', tone: 'caution' })
  } else if (diagnostic?.simulation === 'candidate') {
    statuses.push({ key: 'candidate', label: 'candidate', tone: 'accent' })
  }
  return statuses
}

function rowMatchesFilters(row: RuleRowView, filters: RuleFilterState) {
  const query = filters.query.trim().toLowerCase()
  if (query && !row.searchableText.includes(query)) return false
  if (filters.enabled === 'enabled' && !row.rule.enabled) return false
  if (filters.enabled === 'disabled' && row.rule.enabled) return false
  if (filters.actionType && ruleActionKind(row.rule) !== filters.actionType) return false
  if (filters.status !== 'all' && !row.statuses.some((status) => status.key === filters.status)) {
    return false
  }
  return true
}

function ruleActionKind(rule: Rule) {
  return rule.action.renderer || rule.action.type || ''
}
