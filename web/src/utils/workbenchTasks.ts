import type { Rule } from '@/types'
import type { RuleEditorMode } from '@/utils/ruleFormAdapter'
import { ruleDisplayName } from '@/utils/ruleSummaries'

export type WorkbenchMode = 'inspect' | 'editor' | 'simulate' | 'readiness' | 'snapshots' | 'result'
export type WorkbenchIntent = 'rules' | 'test' | 'release'

export interface WorkbenchTask {
  name: WorkbenchMode | WorkbenchIntent
  label: string
  description: string
}

export const workbenchTasks: WorkbenchTask[] = [
  { name: 'inspect', label: 'Rule', description: 'Selected rule overview' },
  { name: 'editor', label: 'Edit', description: 'Create or edit rule payload' },
  { name: 'simulate', label: 'Simulate', description: 'Run draft or published request simulation' },
  { name: 'readiness', label: 'Readiness', description: 'Validate before publishing draft state' },
  { name: 'snapshots', label: 'Snapshots', description: 'Review snapshots and rollback impact' },
  { name: 'result', label: 'Result', description: 'Structured diagnostics and raw JSON' },
]

export const workbenchIntentTasks: WorkbenchTask[] = [
  { name: 'rules', label: 'Rules', description: 'Inspect, create, and edit rules' },
  { name: 'test', label: 'Test', description: 'Simulate requests and inspect results' },
  { name: 'release', label: 'Release', description: 'Validate, publish, and rollback' },
]

export function isWorkbenchMode(value: string): value is WorkbenchMode {
  return workbenchTasks.some((task) => task.name === value)
}

export function isWorkbenchIntent(value: string): value is WorkbenchIntent {
  return workbenchIntentTasks.some((task) => task.name === value)
}

export function workbenchIntentForMode(mode: WorkbenchMode): WorkbenchIntent {
  if (mode === 'simulate' || mode === 'result') return 'test'
  if (mode === 'readiness' || mode === 'snapshots') return 'release'
  return 'rules'
}

export function defaultWorkbenchModeForIntent(intent: WorkbenchIntent): WorkbenchMode {
  if (intent === 'test') return 'simulate'
  if (intent === 'release') return 'readiness'
  return 'inspect'
}

export function workbenchTitle(mode: WorkbenchMode, rule: Rule | null, editorMode: RuleEditorMode) {
  if (mode === 'inspect') return rule ? `Inspect ${ruleDisplayName(rule)}` : 'Rule overview'
  if (mode === 'editor') return editorMode === 'create' ? 'Create rule' : rule ? `Edit ${ruleDisplayName(rule)}` : 'Rule editor'
  if (mode === 'readiness') return 'Publish readiness'
  if (mode === 'snapshots') return 'Snapshot and rollback'
  if (mode === 'result') return 'Diagnostics result'
  return rule ? `Simulate ${ruleDisplayName(rule)}` : 'Request simulation'
}
