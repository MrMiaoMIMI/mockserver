import type { Condition, MockEvent, Rule, RuleSet } from '@/types'
import { buildDefaultSimulationEvent } from '@/utils/simulationDiagnostics'
import { actionSummary, conditionSummary } from '@/utils/ruleSummaries'
import { formToRule, type BuildRuleOptions, type RuleFormError, type RuleFormState } from '@/utils/ruleFormAdapter'

export type ConditionPresetId = 'method' | 'path' | 'query' | 'header' | 'body' | 'host' | 'trace' | 'custom'
export type ReadinessTone = 'ok' | 'warn' | 'danger' | 'muted'

export interface ConditionPreset {
  id: ConditionPresetId
  label: string
  description: string
  defaultField: string
  defaultOp: string
  defaultValue: unknown
  fieldPrefix?: string
  keyLabel?: string
  keyPlaceholder?: string
  valuePlaceholder?: string
  operators: string[]
}

export interface ActionAuthoringProfile {
  type: string
  label: string
  detail: string
}

export interface SectionReadiness {
  section: RuleFormError['section'] | 'preview'
  label: string
  tone: ReadinessTone
  status: string
  detail: string
  errorCount: number
}

export interface RuleAuthoringView {
  rule?: Rule
  errors: RuleFormError[]
  sectionReadiness: SectionReadiness[]
  conditionSummary: string
  actionSummary: string
  actionDescription: string
  event?: MockEvent
  eventJson: string
  draftOverride?: RuleSet
  canSimulate: boolean
}

export const CONDITION_PRESETS: ConditionPreset[] = [
  {
    id: 'method',
    label: 'Method',
    description: 'Match the HTTP method.',
    defaultField: 'request.method',
    defaultOp: 'eq',
    defaultValue: 'GET',
    valuePlaceholder: 'GET',
    operators: ['eq', 'ne', 'in', 'not_in'],
  },
  {
    id: 'path',
    label: 'Path',
    description: 'Match the request path.',
    defaultField: 'request.path',
    defaultOp: 'eq',
    defaultValue: '/api/v1/debug',
    valuePlaceholder: '/api/v1/debug',
    operators: ['eq', 'prefix', 'suffix', 'contains', 'regex'],
  },
  {
    id: 'query',
    label: 'Query',
    description: 'Match a query parameter value.',
    defaultField: 'request.query.q1[0]',
    fieldPrefix: 'request.query.',
    keyLabel: 'Query key',
    keyPlaceholder: 'q1',
    defaultOp: 'eq',
    defaultValue: 'qv1',
    valuePlaceholder: 'qv1',
    operators: ['eq', 'ne', 'contains', 'not_contains', 'exists', 'not_exists', 'regex'],
  },
  {
    id: 'header',
    label: 'Header',
    description: 'Match a request header value.',
    defaultField: 'request.headers.x-env[0]',
    fieldPrefix: 'request.headers.',
    keyLabel: 'Header name',
    keyPlaceholder: 'x-env',
    defaultOp: 'eq',
    defaultValue: 'test',
    valuePlaceholder: 'test',
    operators: ['eq', 'ne', 'contains', 'not_contains', 'exists', 'not_exists', 'regex'],
  },
  {
    id: 'body',
    label: 'Body field',
    description: 'Match a JSON request body field.',
    defaultField: 'request.body.user.id',
    fieldPrefix: 'request.body.',
    keyLabel: 'Body path',
    keyPlaceholder: 'user.id',
    defaultOp: 'eq',
    defaultValue: 'u-1',
    valuePlaceholder: 'u-1',
    operators: ['eq', 'ne', 'contains', 'not_contains', 'exists', 'not_exists', 'gt', 'gte', 'lt', 'lte'],
  },
  {
    id: 'host',
    label: 'Host',
    description: 'Match the normalized request host.',
    defaultField: 'request.host',
    defaultOp: 'eq',
    defaultValue: 'demo.com',
    valuePlaceholder: 'demo.com',
    operators: ['eq', 'ne', 'contains', 'prefix', 'suffix', 'regex'],
  },
  {
    id: 'trace',
    label: 'Trace ID',
    description: 'Match the request trace ID.',
    defaultField: 'meta.trace_id',
    defaultOp: 'eq',
    defaultValue: 'trace-001',
    valuePlaceholder: 'trace-001',
    operators: ['eq', 'ne', 'contains', 'prefix', 'suffix', 'regex'],
  },
]

export const CONDITION_OPERATORS = [
  { value: 'eq', label: 'equals' },
  { value: 'ne', label: 'not equals' },
  { value: 'in', label: 'in list' },
  { value: 'not_in', label: 'not in list' },
  { value: 'contains', label: 'contains' },
  { value: 'not_contains', label: 'not contains' },
  { value: 'exists', label: 'exists' },
  { value: 'not_exists', label: 'not exists' },
  { value: 'is_null', label: 'is null' },
  { value: 'is_not_null', label: 'is not null' },
  { value: 'prefix', label: 'starts with' },
  { value: 'suffix', label: 'ends with' },
  { value: 'regex', label: 'matches regex' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
]

export const ACTION_AUTHORING_PROFILES: ActionAuthoringProfile[] = [
  {
    type: 'static',
    label: 'Static',
    detail: 'Return a fixed protocol response payload.',
  },
  {
    type: 'template',
    label: 'Template',
    detail: 'Render a protocol response payload from request fields.',
  },
  {
    type: 'cel',
    label: 'CEL',
    detail: 'Build a response from an expression.',
  },
  {
    type: 'sequence',
    label: 'Sequence',
    detail: 'Return ordered responses across repeated calls.',
  },
  {
    type: 'webhook',
    label: 'Webhook',
    detail: 'Call another service and use its response.',
  },
]

const sectionLabels: Record<SectionReadiness['section'], string> = {
  identity: 'Identity',
  condition: 'Condition',
  action: 'Action',
  advanced: 'Raw JSON',
  preview: 'Preview',
}

export function buildRuleAuthoringView(
  form: RuleFormState,
  ruleSet: RuleSet | null,
  options: BuildRuleOptions = {}
): RuleAuthoringView {
  const result = formToRule(form, options)
  const rawJsonError = rawRuleJsonError(form.rawRuleJson)
  const errors = rawJsonError ? [...result.errors, rawJsonError] : result.errors
  const rule = result.rule
  const event = rule ? buildDefaultSimulationEvent(ruleSet, rule) : undefined
  const draftOverride = rule && ruleSet ? buildDraftOverride(ruleSet, rule, options.lockedRuleId) : undefined

  return {
    rule,
    errors,
    sectionReadiness: buildSectionReadiness(form, errors, rule, event),
    conditionSummary: rule ? conditionSummary(rule.when) : fallbackConditionSummary(form),
    actionSummary: rule ? actionSummary(rule.action) : fallbackActionSummary(form),
    actionDescription: actionDescription(form.actionType),
    event,
    eventJson: event ? formatJSON(event) : '',
    draftOverride,
    canSimulate: Boolean(rule && event && draftOverride && !errors.length),
  }
}

export function buildDraftOverride(ruleSet: RuleSet, rule: Rule, lockedRuleId = ''): RuleSet {
  const rules = [...ruleSet.rules]
  const index = lockedRuleId
    ? rules.findIndex((item) => item.id === lockedRuleId)
    : rules.findIndex((item) => item.id === rule.id)

  if (index >= 0) {
    rules[index] = rule
  } else {
    rules.push(rule)
  }

  return {
    ...ruleSet,
    rules,
  }
}

export function resolveConditionPreset(condition: Condition): ConditionPreset {
  const field = condition.field || ''
  return (
    CONDITION_PRESETS.find((preset) => {
      if (preset.fieldPrefix) return field.startsWith(preset.fieldPrefix)
      return field === preset.defaultField
    }) || customConditionPreset(condition)
  )
}

export function applyConditionPreset(condition: Condition, presetId: ConditionPresetId): Condition {
  if (presetId === 'custom') {
    return {
      field: 'request.custom',
      op: condition.op || 'eq',
      value: condition.value ?? '',
    }
  }
  const preset = presetById(presetId)
  if (!preset) return condition
  return {
    field: preset.defaultField,
    op: preset.defaultOp,
    value: preset.defaultValue,
  }
}

export function updateConditionPresetKey(condition: Condition, key: string): Condition {
  const preset = resolveConditionPreset(condition)
  if (!preset.fieldPrefix) return condition
  const normalizedKey = normalizePresetKey(key, preset.id)
  return {
    ...condition,
    field: fieldForPresetKey(preset, normalizedKey),
  }
}

export function conditionPresetKey(condition: Condition) {
  const preset = resolveConditionPreset(condition)
  if (!preset.fieldPrefix) return ''
  return (condition.field || '').slice(preset.fieldPrefix.length).replace(/\[(?:\*|\d+)]$/, '')
}

export function operatorLabel(operator: string) {
  return CONDITION_OPERATORS.find((item) => item.value === operator)?.label || operator
}

export function predicateValueInput(value: unknown) {
  if (value === undefined) return ''
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

export function parsePredicateValueInput(value: string): unknown {
  const trimmed = value.trim()
  if (!trimmed) return ''
  try {
    return JSON.parse(trimmed)
  } catch {
    return value
  }
}

export function validatePredicateCondition(condition: Condition): string[] {
  const errors: string[] = []
  const field = condition.field?.trim() || ''
  const operator = condition.op?.trim() || ''
  if (!field) errors.push('Field is required')
  if (!operator) errors.push('Operator is required')
  const preset = resolveConditionPreset(condition)
  if (preset.fieldPrefix && !conditionPresetKey(condition).trim()) {
    errors.push(`${preset.keyLabel || 'Key'} is required`)
  }
  if (!['exists', 'not_exists', 'is_null', 'is_not_null'].includes(operator) && condition.value === undefined) {
    errors.push('Value is required')
  }
  return errors
}

export function actionDescription(actionType: string) {
  return ACTION_AUTHORING_PROFILES.find((profile) => profile.type === actionType)?.detail || 'Custom action payload.'
}

function buildSectionReadiness(
  form: RuleFormState,
  errors: RuleFormError[],
  rule?: Rule,
  event?: MockEvent
): SectionReadiness[] {
  return (['identity', 'condition', 'action', 'advanced', 'preview'] as const).map((section) => {
    const sectionErrors = errors.filter((error) => error.section === section)
    if (sectionErrors.length) {
      return {
        section,
        label: sectionLabels[section],
        tone: 'danger',
        status: `${sectionErrors.length} issue${sectionErrors.length > 1 ? 's' : ''}`,
        detail: sectionErrors[0].message,
        errorCount: sectionErrors.length,
      }
    }
    if (section === 'preview') {
      return {
        section,
        label: sectionLabels[section],
        tone: rule && event ? 'ok' : 'muted',
        status: rule && event ? 'ready' : 'waiting',
        detail: rule && event ? 'Simulation event is ready.' : 'Fix the rule before preview.',
        errorCount: 0,
      }
    }
    if (section === 'advanced') {
      return {
        section,
        label: sectionLabels[section],
        tone: form.rawRuleJson.trim() ? 'ok' : 'muted',
        status: form.rawRuleJson.trim() ? 'synced' : 'empty',
        detail: form.rawRuleJson.trim() ? 'Raw rule JSON is parseable.' : 'Raw JSON can be generated from the form.',
        errorCount: 0,
      }
    }
    return {
      section,
      label: sectionLabels[section],
      tone: 'ok',
      status: 'ready',
      detail: readyDetail(section),
      errorCount: 0,
    }
  })
}

function readyDetail(section: RuleFormError['section']) {
  if (section === 'identity') return 'Rule Name, generated Rule ID, priority, and enabled state are valid.'
  if (section === 'condition') return 'Condition can be converted to a rule payload.'
  if (section === 'action') return 'Action can be converted to a rule payload.'
  return 'Raw JSON is available for advanced edits.'
}

function rawRuleJsonError(rawRuleJson: string): RuleFormError | null {
  if (!rawRuleJson.trim()) return null
  try {
    JSON.parse(rawRuleJson)
    return null
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    return {
      section: 'advanced',
      field: 'rawRuleJson',
      message: `Raw Rule JSON is not valid JSON: ${message}`,
    }
  }
}

function fallbackConditionSummary(form: RuleFormState) {
  if (form.conditionMode === 'expr') return form.conditionExpr.trim() || 'CEL expression needs input'
  if (form.conditionMode === 'raw') return form.conditionJson.trim() || 'Raw condition JSON needs input'
  return conditionSummary(form.conditionTree)
}

function fallbackActionSummary(form: RuleFormState) {
  return `${actionProfileLabel(form.actionType)} response needs valid fields`
}

function actionProfileLabel(actionType: string) {
  return ACTION_AUTHORING_PROFILES.find((profile) => profile.type === actionType)?.label || actionType || 'Unknown'
}

function customConditionPreset(condition: Condition): ConditionPreset {
  return {
    id: 'custom',
    label: 'Custom field',
    description: 'Match a custom event field path.',
    defaultField: condition.field || 'request.path',
    defaultOp: condition.op || 'eq',
    defaultValue: condition.value ?? '',
    operators: CONDITION_OPERATORS.map((operator) => operator.value),
  }
}

function presetById(presetId: ConditionPresetId) {
  return CONDITION_PRESETS.find((preset) => preset.id === presetId)
}

function normalizePresetKey(key: string, presetId: ConditionPresetId) {
  const trimmed = key.trim()
  if (presetId === 'header') return trimmed.toLowerCase()
  return trimmed
}

function fieldForPresetKey(preset: ConditionPreset, key: string) {
  const base = `${preset.fieldPrefix}${key}`
  if (preset.id === 'query' || preset.id === 'header') return `${base}[0]`
  return base
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}
