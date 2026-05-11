import type {
  PublishedRuleSetSnapshot,
  PublishRuleSetResponse,
  Rule,
  RuleSet,
  ValidateRuleSetResponse,
  ValidationIssue,
} from '@/types'
import type { RuleDiagnosticState } from '@/utils/ruleCollection'

export type DraftPublicationState = 'draft_only' | 'published_current' | 'changed'
export type ValidationState = 'not_run' | 'valid' | 'invalid'
export type ReadinessTone = 'ok' | 'warn' | 'blocked' | 'muted'

export interface ReadinessMetric {
  label: string
  value: string
  tone: ReadinessTone
}

export interface ValidationIssueView {
  path: string
  message: string
  section: string
  scope: string
  ruleId?: string
}

export interface PublishReadinessView {
  draftState: DraftPublicationState
  draftStateLabel: string
  draftStateDetail: string
  validationState: ValidationState
  validationLabel: string
  validationDetail: string
  canPublish: boolean
  metrics: ReadinessMetric[]
  issues: ValidationIssueView[]
  warnings: ValidationIssueView[]
  publishedSnapshotLabel: string
  publishResultLabel: string
}

export function buildPublishReadinessView(
  ruleSet: RuleSet | null,
  currentPublished: PublishedRuleSetSnapshot | null,
  validation: ValidateRuleSetResponse | null,
  publishResult: PublishRuleSetResponse | null
): PublishReadinessView {
  const validationState = validationStateOf(validation)
  const issues = mapValidationIssues(validation?.result.issues || [], ruleSet)
  const warnings = mapValidationIssues(validation?.result.warnings || [], ruleSet)
  const draftState = publicationStateOf(ruleSet, currentPublished)
  const publishedSnapshotLabel = currentPublished
    ? `${currentPublished.snapshot_id} / v${currentPublished.ruleset.version || 0}`
    : 'no published snapshot'
  const publishResultLabel = publishResult
    ? `${publishResult.snapshot.snapshot_id} / v${publishResult.ruleset.version || publishResult.snapshot.ruleset.version || 0}`
    : 'no publish in this session'

  return {
    draftState,
    draftStateLabel: publicationStateLabel(draftState),
    draftStateDetail: publicationStateDetail(draftState, currentPublished),
    validationState,
    validationLabel: validationStateLabel(validationState),
    validationDetail: validationStateDetail(validationState, issues.length, warnings.length),
    canPublish: validationState === 'valid' && Boolean(ruleSet),
    metrics: [
      {
        label: 'validation',
        value: validationStateLabel(validationState),
        tone: validationStateTone(validationState),
      },
      {
        label: 'publication',
        value: publicationStateLabel(draftState),
        tone: publicationStateTone(draftState),
      },
      {
        label: 'warnings',
        value: String(warnings.length),
        tone: warnings.length ? 'warn' : 'ok',
      },
      {
        label: 'rules',
        value: String(ruleSet?.rules.length || 0),
        tone: ruleSet?.rules.length ? 'ok' : 'warn',
      },
      {
        label: 'published',
        value: currentPublished ? `v${currentPublished.ruleset.version || 0}` : '-',
        tone: currentPublished ? 'ok' : 'muted',
      },
    ],
    issues,
    warnings,
    publishedSnapshotLabel,
    publishResultLabel,
  }
}

export function mapValidationIssues(issues: ValidationIssue[], ruleSet?: RuleSet | null) {
  return issues.map((issue) => {
    const rule = ruleFromIssuePath(issue.path, ruleSet)
    return {
      path: issue.path,
      message: issue.message,
      section: sectionFromIssuePath(issue.path),
      scope: rule ? `rule ${rule.id}` : scopeFromIssuePath(issue.path),
      ruleId: rule?.id,
    }
  })
}

export function buildValidationRuleDiagnostics(
  validation: ValidateRuleSetResponse | null,
  ruleSet?: RuleSet | null
): Record<string, RuleDiagnosticState> {
  if (!validation?.result.issues?.length || !ruleSet) return {}
  return mapValidationIssues(validation.result.issues, ruleSet).reduce<Record<string, RuleDiagnosticState>>(
    (diagnostics, issue) => {
      if (!issue.ruleId) return diagnostics
      const current = diagnostics[issue.ruleId] || {}
      diagnostics[issue.ruleId] = {
        ...current,
        validation: 'invalid',
        messages: [...(current.messages || []), `${issue.path}: ${issue.message}`],
      }
      return diagnostics
    },
    {}
  )
}

function validationStateOf(validation: ValidateRuleSetResponse | null): ValidationState {
  if (!validation) return 'not_run'
  return validation.result.valid ? 'valid' : 'invalid'
}

function publicationStateOf(
  ruleSet: RuleSet | null,
  currentPublished: PublishedRuleSetSnapshot | null
): DraftPublicationState {
  if (!ruleSet || !currentPublished) return 'draft_only'
  return stableRuleSetJSON(ruleSet) === stableRuleSetJSON(currentPublished.ruleset)
    ? 'published_current'
    : 'changed'
}

function publicationStateLabel(state: DraftPublicationState) {
  if (state === 'published_current') return 'published current'
  if (state === 'changed') return 'draft changed'
  return 'draft only'
}

function publicationStateDetail(
  state: DraftPublicationState,
  currentPublished: PublishedRuleSetSnapshot | null
) {
  if (state === 'published_current') return 'draft content matches the current published snapshot'
  if (state === 'changed') return 'draft likely differs from the current published snapshot'
  return currentPublished ? 'draft has no comparable published snapshot' : 'publish to create the first snapshot'
}

function validationStateLabel(state: ValidationState) {
  if (state === 'valid') return 'valid'
  if (state === 'invalid') return 'blocked'
  return 'not run'
}

function validationStateDetail(state: ValidationState, issueCount: number, warningCount: number) {
  if (state === 'valid') {
    return warningCount
      ? `backend validation passed with ${warningCount} runtime warning${warningCount === 1 ? '' : 's'}`
      : 'backend validation passed'
  }
  if (state === 'invalid') return `${issueCount} validation issue${issueCount === 1 ? '' : 's'}`
  return 'run validation before publishing'
}

function validationStateTone(state: ValidationState): ReadinessTone {
  if (state === 'valid') return 'ok'
  if (state === 'invalid') return 'blocked'
  return 'warn'
}

function publicationStateTone(state: DraftPublicationState): ReadinessTone {
  if (state === 'published_current') return 'ok'
  if (state === 'changed') return 'warn'
  return 'muted'
}

function sectionFromIssuePath(path: string) {
  if (path.includes('.when')) return 'condition'
  if (path.includes('.action')) return 'action'
  if (path.includes('.selector')) return 'selector'
  if (path.includes('.rules[')) return 'rule'
  if (path.includes('namespace')) return 'namespace'
  return 'ruleset'
}

function scopeFromIssuePath(path: string) {
  if (path.startsWith('selector') || path.includes('.selector')) return 'ruleset selector'
  if (path.startsWith('rules[')) return 'rule stack'
  return 'ruleset'
}

function ruleFromIssuePath(path: string, ruleSet?: RuleSet | null): Rule | undefined {
  const match = path.match(/rules\[(\d+)]/)
  if (!match) return undefined
  const index = Number(match[1])
  return Number.isFinite(index) ? ruleSet?.rules[index] : undefined
}

function stableRuleSetJSON(ruleSet: RuleSet) {
  const { version: _version, ...rest } = ruleSet
  return JSON.stringify(sortValue(rest))
}

function sortValue(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sortValue)
  if (!value || typeof value !== 'object') return value
  return Object.keys(value as Record<string, unknown>)
    .sort()
    .reduce<Record<string, unknown>>((result, key) => {
      result[key] = sortValue((value as Record<string, unknown>)[key])
      return result
    }, {})
}
