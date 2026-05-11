import type { Condition, Rule, RuleAction, RuleSet, SequenceStep } from '@/types'

export type RuleEditorMode = 'create' | 'edit'
export type RuleConditionMode = 'tree' | 'expr' | 'raw'
export type RuleFormBuildSource = 'form' | 'raw'

export interface RuleSequenceStepForm {
  key: string
  status: number
  headersJson: string
  bodyJson: string
}

export interface RuleFormState {
  id: string
  name: string
  enabled: boolean
  priority: number
  conditionMode: RuleConditionMode
  conditionTree: Condition
  conditionExpr: string
  conditionJson: string
  actionType: string
  status: number
  headersJson: string
  bodyJson: string
  bodyTemplate: string
  bodyExpression: string
  sequenceStrategy: string
  sequenceSteps: RuleSequenceStepForm[]
  webhookUrl: string
  webhookMethod: string
  webhookTimeoutMS: number
  webhookHeadersJson: string
  rawRuleJson: string
}

export interface RuleFormError {
  section: 'identity' | 'condition' | 'action' | 'advanced'
  field: string
  message: string
}

export interface BuildRuleOptions {
  source?: RuleFormBuildSource
  existingRuleIds?: string[]
  lockedRuleId?: string
}

export interface BuildRuleResult {
  rule?: Rule
  errors: RuleFormError[]
}

const statusActionTypes = new Set(['static_response', 'template_response', 'cel_response'])

export function defaultRuleForm(ruleSet?: RuleSet | null): RuleFormState {
  const rule: Rule = {
    id: nextRuleId(ruleSet),
    name: '',
    enabled: true,
    priority: nextPriority(ruleSet),
    when: defaultConditionTree(ruleSet?.protocol),
    action: {
      type: 'static_response',
      status: 200,
      headers: {
        'content-type': ['application/json'],
      },
      body: {
        message: 'hello mockserver',
      },
    },
  }
  return ruleToForm(rule)
}

export function ruleToForm(rule: Rule): RuleFormState {
  const conditionMode = conditionModeFromRule(rule.when)
  const form: RuleFormState = {
    id: rule.id,
    name: rule.name || '',
    enabled: rule.enabled,
    priority: rule.priority,
    conditionMode,
    conditionTree: conditionMode === 'tree' ? clone(rule.when) : defaultConditionTree(),
    conditionExpr: rule.when.expr || '',
    conditionJson: formatJSON(rule.when),
    actionType: rule.action.type,
    status: rule.action.status ?? 200,
    headersJson: formatJSON(rule.action.headers || {}),
    bodyJson: formatJSON(rule.action.body ?? {}),
    bodyTemplate: rule.action.body_template || '',
    bodyExpression: rule.action.body_expression || '',
    sequenceStrategy: rule.action.sequence_strategy || 'last',
    sequenceSteps: stepsToForm(rule.action.sequence || defaultSequence()),
    webhookUrl: rule.action.webhook?.url || '',
    webhookMethod: rule.action.webhook?.method || 'POST',
    webhookTimeoutMS: rule.action.webhook?.timeout_ms || 3000,
    webhookHeadersJson: formatJSON(rule.action.webhook?.headers || {}),
    rawRuleJson: '',
  }
  form.rawRuleJson = formatJSON(formToRuleUnsafe(form))
  return form
}

export function formToRule(form: RuleFormState, options: BuildRuleOptions = {}): BuildRuleResult {
  const errors: RuleFormError[] = []
  const source = options.source || 'form'
  const rule = source === 'raw' ? parseRawRule(form.rawRuleJson, errors) : buildRuleFromForm(form, errors)

  if (rule) {
    validateRuleIdentity(rule, errors, options)
  }

  return {
    rule: errors.length ? undefined : rule,
    errors,
  }
}

export function refreshRawRuleJson(form: RuleFormState) {
  const result = formToRule(form, { source: 'form' })
  return result.rule ? formatJSON(result.rule) : form.rawRuleJson
}

export function applyRawRuleJson(form: RuleFormState): BuildRuleResult {
  const result = formToRule(form, { source: 'raw' })
  if (result.rule) {
    Object.assign(form, ruleToForm(result.rule))
  }
  return result
}

export function newSequenceStepForm(index: number): RuleSequenceStepForm {
  return {
    key: `step-${Date.now()}-${index}`,
    status: 200,
    headersJson: '{}',
    bodyJson: '{\n  "message": "ok"\n}',
  }
}

export function generateRuleIdFromName(
  name: string,
  ruleSet?: RuleSet | null,
  lockedRuleId = ''
) {
  const slug = slugRuleName(name)
  const fallback = nextRuleId(ruleSet)
  return uniqueRuleId(slug || fallback, ruleSet, lockedRuleId)
}

function buildRuleFromForm(form: RuleFormState, errors: RuleFormError[]): Rule | undefined {
  const id = form.id.trim()
  const name = form.name.trim()
  if (!id) {
    errors.push({ section: 'identity', field: 'id', message: 'Rule ID is required' })
  }
  if (!name) {
    errors.push({ section: 'identity', field: 'name', message: 'Rule Name is required' })
  }
  if (!Number.isFinite(form.priority) || form.priority < 0) {
    errors.push({ section: 'identity', field: 'priority', message: 'Priority must be a non-negative number' })
  }

  const when = buildCondition(form, errors)
  const action = buildAction(form, errors)
  if (!id || !name || !when || !action || errors.length) return undefined

  return {
    id,
    name,
    enabled: form.enabled,
    priority: form.priority,
    when,
    action,
  }
}

function buildCondition(form: RuleFormState, errors: RuleFormError[]): Condition | undefined {
  if (form.conditionMode === 'expr') {
    const expr = form.conditionExpr.trim()
    if (!expr) {
      errors.push({
        section: 'condition',
        field: 'conditionExpr',
        message: 'CEL Expression is required',
      })
      return undefined
    }
    return { expr }
  }
  if (form.conditionMode === 'raw') {
    return parseJSON<Condition>(form.conditionJson, {
      section: 'condition',
      field: 'conditionJson',
      label: 'Condition JSON',
      errors,
    })
  }
  const validationErrors = validateConditionTree(form.conditionTree)
  validationErrors.forEach((message) => {
    errors.push({ section: 'condition', field: 'conditionTree', message })
  })
  return validationErrors.length ? undefined : clone(form.conditionTree)
}

function buildAction(form: RuleFormState, errors: RuleFormError[]): RuleAction | undefined {
  const action: RuleAction = {
    type: form.actionType,
  }

  if (!action.type) {
    errors.push({ section: 'action', field: 'actionType', message: 'Action Type is required' })
    return undefined
  }

  if (statusActionTypes.has(action.type)) {
    if (!isHttpStatus(form.status)) {
      errors.push({ section: 'action', field: 'status', message: 'Status must be between 100 and 599' })
    }
    action.status = form.status
    action.headers = parseJSON<Record<string, string[]>>(form.headersJson || '{}', {
      section: 'action',
      field: 'headersJson',
      label: 'Headers JSON',
      errors,
    })
  }

  if (action.type === 'static_response') {
    action.body = parseJSON<unknown>(form.bodyJson || 'null', {
      section: 'action',
      field: 'bodyJson',
      label: 'Body JSON',
      errors,
    })
  }
  if (action.type === 'template_response') {
    if (!form.bodyTemplate.trim()) {
      errors.push({ section: 'action', field: 'bodyTemplate', message: 'Body Template is required' })
    }
    action.body_template = form.bodyTemplate
  }
  if (action.type === 'cel_response') {
    if (!form.bodyExpression.trim()) {
      errors.push({ section: 'action', field: 'bodyExpression', message: 'Body Expression is required' })
    }
    action.body_expression = form.bodyExpression
  }
  if (action.type === 'sequence_response') {
    action.sequence_strategy = form.sequenceStrategy || 'last'
    action.sequence = buildSequence(form.sequenceSteps, errors)
  }
  if (action.type === 'webhook_response') {
    if (!form.webhookUrl.trim()) {
      errors.push({ section: 'action', field: 'webhookUrl', message: 'Webhook URL is required' })
    }
    if (!Number.isFinite(form.webhookTimeoutMS) || form.webhookTimeoutMS < 0) {
      errors.push({
        section: 'action',
        field: 'webhookTimeoutMS',
        message: 'Timeout MS must be a non-negative number',
      })
    }
    action.webhook = {
      url: form.webhookUrl.trim(),
      method: form.webhookMethod || 'POST',
      timeout_ms: form.webhookTimeoutMS,
      headers: parseJSON<Record<string, string[]>>(form.webhookHeadersJson || '{}', {
        section: 'action',
        field: 'webhookHeadersJson',
        label: 'Webhook Headers JSON',
        errors,
      }),
    }
  }

  return errors.some((error) => error.section === 'action') ? undefined : action
}

function buildSequence(steps: RuleSequenceStepForm[], errors: RuleFormError[]): SequenceStep[] {
  if (!steps.length) {
    errors.push({ section: 'action', field: 'sequenceSteps', message: 'At least one sequence step is required' })
    return []
  }
  return steps.map((step, index) => {
    if (!isHttpStatus(step.status)) {
      errors.push({
        section: 'action',
        field: `sequenceSteps.${index}.status`,
        message: `Step ${index + 1} status must be between 100 and 599`,
      })
    }
    return {
      status: step.status,
      headers: parseJSON<Record<string, string[]>>(step.headersJson || '{}', {
        section: 'action',
        field: `sequenceSteps.${index}.headersJson`,
        label: `Step ${index + 1} Headers JSON`,
        errors,
      }),
      body: parseJSON<unknown>(step.bodyJson || 'null', {
        section: 'action',
        field: `sequenceSteps.${index}.bodyJson`,
        label: `Step ${index + 1} Body JSON`,
        errors,
      }),
    }
  })
}

function parseRawRule(rawRuleJson: string, errors: RuleFormError[]) {
  const rule = parseJSON<Rule>(rawRuleJson, {
    section: 'advanced',
    field: 'rawRuleJson',
    label: 'Raw Rule JSON',
    errors,
  })
  if (!rule) return undefined
  if (!rule.when) {
    errors.push({ section: 'advanced', field: 'rawRuleJson', message: 'Raw Rule JSON is missing when' })
  }
  if (!rule.name?.trim()) {
    errors.push({ section: 'advanced', field: 'rawRuleJson', message: 'Raw Rule JSON is missing name' })
  }
  if (!rule.action?.type) {
    errors.push({
      section: 'advanced',
      field: 'rawRuleJson',
      message: 'Raw Rule JSON is missing action.type',
    })
  }
  if (typeof rule.enabled !== 'boolean') {
    errors.push({
      section: 'advanced',
      field: 'rawRuleJson',
      message: 'Raw Rule JSON enabled must be boolean',
    })
  }
  if (typeof rule.priority !== 'number') {
    errors.push({
      section: 'advanced',
      field: 'rawRuleJson',
      message: 'Raw Rule JSON priority must be number',
    })
  }
  return rule
}

function validateRuleIdentity(rule: Rule, errors: RuleFormError[], options: BuildRuleOptions) {
  if (!rule.id?.trim()) {
    errors.push({ section: 'identity', field: 'id', message: 'Rule ID is required' })
  }
  if (!rule.name?.trim()) {
    errors.push({ section: 'identity', field: 'name', message: 'Rule Name is required' })
  }
  if (options.lockedRuleId && rule.id !== options.lockedRuleId) {
    errors.push({
      section: 'advanced',
      field: 'rawRuleJson',
      message: `Rule ID must remain ${options.lockedRuleId} in edit mode`,
    })
  }
  const duplicateId = options.existingRuleIds?.some((id) => id === rule.id && id !== options.lockedRuleId)
  if (duplicateId) {
    errors.push({ section: 'identity', field: 'id', message: `Rule ID ${rule.id} already exists` })
  }
}

function validateConditionTree(condition: Condition): string[] {
  if (condition.expr !== undefined) {
    return condition.expr.trim() ? [] : ['CEL Expression is required']
  }
  if (condition.all) {
    return validateGroupCondition('ALL', condition.all)
  }
  if (condition.any) {
    return validateGroupCondition('ANY', condition.any)
  }
  if (condition.not) {
    return validateConditionTree(condition.not)
  }
  if (!condition.field?.trim()) return ['Predicate field is required']
  if (!condition.op?.trim()) return ['Predicate op is required']
  return []
}

function validateGroupCondition(kind: string, children: Condition[]) {
  if (!children.length) return [`${kind} requires at least one child condition`]
  return children.flatMap(validateConditionTree)
}

function conditionModeFromRule(condition: Condition): RuleConditionMode {
  if (condition.expr !== undefined) return 'expr'
  return 'tree'
}

function stepsToForm(steps: SequenceStep[]): RuleSequenceStepForm[] {
  return steps.map((step, index) => ({
    key: `step-${index + 1}`,
    status: step.status,
    headersJson: formatJSON(step.headers || {}),
    bodyJson: formatJSON(step.body ?? null),
  }))
}

function defaultSequence(): SequenceStep[] {
  return [
    {
      status: 202,
      body: {
        state: 'pending',
      },
    },
    {
      status: 200,
      body: {
        state: 'done',
      },
    },
  ]
}

function defaultConditionTree(protocol = 'http'): Condition {
  if (protocol === 'cache') {
    return {
      all: [
        { field: 'request.operation', op: 'eq', value: 'get' },
        { field: 'request.key', op: 'prefix', value: 'user:' },
      ],
    }
  }
  return {
    all: [
      { field: 'request.method', op: 'eq', value: 'GET' },
      { field: 'request.path', op: 'eq', value: '/api/v1/debug' },
    ],
  }
}

function nextRuleId(ruleSet?: RuleSet | null) {
  const existing = new Set((ruleSet?.rules || []).map((rule) => rule.id))
  for (let index = existing.size + 1; index < existing.size + 1000; index += 1) {
    const candidate = `rule-${String(index).padStart(3, '0')}`
    if (!existing.has(candidate)) return candidate
  }
  return `rule-${Date.now().toString().slice(-6)}`
}

function slugRuleName(name: string) {
  return name
    .trim()
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 48)
}

function uniqueRuleId(base: string, ruleSet?: RuleSet | null, lockedRuleId = '') {
  const existing = new Set((ruleSet?.rules || []).map((rule) => rule.id).filter((id) => id !== lockedRuleId))
  if (!existing.has(base)) return base
  for (let index = 2; index < 1000; index += 1) {
    const candidate = `${base}-${String(index).padStart(2, '0')}`
    if (!existing.has(candidate)) return candidate
  }
  return `${base}-${Date.now().toString().slice(-6)}`
}

function nextPriority(ruleSet?: RuleSet | null) {
  const priorities = (ruleSet?.rules || []).map((rule) => rule.priority)
  let candidate = priorities.length ? Math.max(...priorities) + 10 : 100
  while (priorities.includes(candidate)) {
    candidate += 10
  }
  return candidate
}

function formToRuleUnsafe(form: RuleFormState): Rule {
  return {
    id: form.id,
    name: form.name,
    enabled: form.enabled,
    priority: form.priority,
    when:
      form.conditionMode === 'expr'
        ? { expr: form.conditionExpr }
        : form.conditionMode === 'raw'
          ? JSON.parse(form.conditionJson)
          : clone(form.conditionTree),
    action: actionFromFormUnsafe(form),
  }
}

function actionFromFormUnsafe(form: RuleFormState): RuleAction {
  const action: RuleAction = {
    type: form.actionType,
  }
  if (statusActionTypes.has(action.type)) {
    action.status = form.status
    action.headers = JSON.parse(form.headersJson || '{}')
  }
  if (action.type === 'static_response') action.body = JSON.parse(form.bodyJson || 'null')
  if (action.type === 'template_response') action.body_template = form.bodyTemplate
  if (action.type === 'cel_response') action.body_expression = form.bodyExpression
  if (action.type === 'sequence_response') {
    action.sequence_strategy = form.sequenceStrategy
    action.sequence = form.sequenceSteps.map((step) => ({
      status: step.status,
      headers: JSON.parse(step.headersJson || '{}'),
      body: JSON.parse(step.bodyJson || 'null'),
    }))
  }
  if (action.type === 'webhook_response') {
    action.webhook = {
      url: form.webhookUrl,
      method: form.webhookMethod,
      timeout_ms: form.webhookTimeoutMS,
      headers: JSON.parse(form.webhookHeadersJson || '{}'),
    }
  }
  return action
}

function parseJSON<T>(
  value: string,
  context: {
    section: RuleFormError['section']
    field: string
    label: string
    errors: RuleFormError[]
  }
) {
  try {
    return JSON.parse(value) as T
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    context.errors.push({
      section: context.section,
      field: context.field,
      message: `${context.label} is not valid JSON: ${message}`,
    })
    return undefined
  }
}

function isHttpStatus(value: number) {
  return Number.isFinite(value) && value >= 100 && value <= 599
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T
}
