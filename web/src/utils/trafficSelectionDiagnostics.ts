import type { TrafficEvent } from '@/types'

export interface TrafficRuleSetCandidateDiagnostic {
  rulesetId: string
  snapshotId: string
  selectorMatched: boolean
  selected: boolean
  selectorSpecificity: number
  message: string
}

export interface TrafficSelectionDiagnostics {
  hasDiagnostics: boolean
  winnerRuleSetId: string
  winnerSnapshotId: string
  rulesetCandidates: TrafficRuleSetCandidateDiagnostic[]
  candidateRuleIds: string[]
  winnerRuleId: string
}

type ExplainRecord = Record<string, unknown>

export function buildTrafficSelectionDiagnostics(event: TrafficEvent | null | undefined): TrafficSelectionDiagnostics {
  const explain = objectRecord(event?.explain)
  const rulesetSelection = objectRecord(explain?.ruleset_selection)
  const ruleSelection = objectRecord(explain?.rule_selection)
  const candidates = arrayRecords(rulesetSelection?.candidates).map((candidate) => ({
    rulesetId: stringValue(candidate.ruleset_id),
    snapshotId: stringValue(candidate.snapshot_id),
    selectorMatched: booleanValue(candidate.selector_matched),
    selected: booleanValue(candidate.selected),
    selectorSpecificity: numberValue(candidate.selector_specificity),
    message: stringValue(candidate.message),
  })).filter((candidate) => candidate.rulesetId)
  const candidateRuleIds = arrayValues(ruleSelection?.candidate_rule_ids).map(String).filter(Boolean)
  const diagnostics = {
    hasDiagnostics: false,
    winnerRuleSetId: stringValue(rulesetSelection?.winner_ruleset_id),
    winnerSnapshotId: stringValue(rulesetSelection?.winner_snapshot_id),
    rulesetCandidates: candidates,
    candidateRuleIds,
    winnerRuleId: stringValue(ruleSelection?.winner_rule_id),
  }
  diagnostics.hasDiagnostics = Boolean(
    diagnostics.winnerRuleSetId ||
    diagnostics.winnerSnapshotId ||
    diagnostics.rulesetCandidates.length ||
    diagnostics.candidateRuleIds.length ||
    diagnostics.winnerRuleId
  )
  return diagnostics
}

function objectRecord(value: unknown): ExplainRecord | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  return value as ExplainRecord
}

function arrayRecords(value: unknown): ExplainRecord[] {
  if (!Array.isArray(value)) return []
  return value.map(objectRecord).filter((item): item is ExplainRecord => Boolean(item))
}

function arrayValues(value: unknown): unknown[] {
  return Array.isArray(value) ? value : []
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function booleanValue(value: unknown) {
  return value === true
}

function numberValue(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}
