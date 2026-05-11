import type { ListResponse } from './common'

export interface Selector {
  all?: Condition[]
}

export interface Condition {
  all?: Condition[]
  any?: Condition[]
  not?: Condition
  expr?: string
  field?: string
  op?: string
  value?: unknown
}

export interface RuleAction {
  type: string
  status?: number
  headers?: Record<string, string[]>
  body?: unknown
  body_template?: string
  body_expression?: string
  sequence?: SequenceStep[]
  sequence_strategy?: string
  webhook?: WebhookConfig
}

export interface SequenceStep {
  status: number
  headers?: Record<string, string[]>
  body?: unknown
}

export interface WebhookConfig {
  url: string
  method?: string
  headers?: Record<string, string[]>
  timeout_ms?: number
}

export interface Rule {
  id: string
  name: string
  enabled: boolean
  priority: number
  when: Condition
  action: RuleAction
}

export interface RuleSet {
  id: string
  name: string
  enabled: boolean
  protocol: string
  namespace: string
  selector: Selector
  rules: Rule[]
  version?: number
}

export type NamespaceFallbackType = 'response' | 'forward'

export interface NamespaceResponseFallback {
  status: number
  headers?: Record<string, string[]>
  body?: unknown
}

export interface NamespaceForwardFallback {
  timeout_ms?: number
}

export interface NamespaceFallbackAction {
  type: NamespaceFallbackType
  response?: NamespaceResponseFallback
  forward?: NamespaceForwardFallback
}

export interface NamespaceConfig {
  id: string
  name?: string
  description?: string
  ruleset_miss_action: NamespaceFallbackAction
  rule_miss_action: NamespaceFallbackAction
}

export interface NamespaceFallbackForm {
  type: NamespaceFallbackType
  responseStatus: number
  responseHeaders: string
  responseBody: string
  forwardTimeoutMs: number
}

export interface AuditInfo {
  action?: string
  operator?: string
  reason?: string
  trace_id?: string
  source_snapshot_id?: string
}

export interface PublishedRuleSetSnapshot {
  snapshot_id: string
  published_at: string
  ruleset: RuleSet
  audit?: AuditInfo
}

export type PublishedRuleSetResponse = PublishedRuleSetSnapshot

export interface PublishRuleSetResponse {
  ruleset: RuleSet
  snapshot: PublishedRuleSetSnapshot
}

export interface RollbackRuleSetResponse {
  ruleset: RuleSet
  snapshot: PublishedRuleSetSnapshot
}

export interface EventRequest {
  [key: string]: unknown
  method?: string
  scheme?: string
  host?: string
  original_host?: string
  path?: string
  query?: Record<string, string[]>
  headers?: Record<string, string[]>
  body?: unknown
  raw_body?: string
  client_ip?: string
}

export interface MockEvent {
  protocol: string
  operation?: string
  namespace: string
  request: EventRequest
  meta?: {
    trace_id?: string
  }
}

export interface MatchTrace {
  ruleset_id?: string
  rule_id?: string
  fallback_reason?: string
}

export interface ActionExecution {
  status: number
  headers?: Record<string, string[]>
  body?: unknown
}

export interface SimulationResult {
  matched: boolean
  fallback?: boolean
  trace: MatchTrace
  candidates?: string[]
  response?: ActionExecution
  explain?: Record<string, unknown>
}

export interface SimulateRuleSetRequest {
  event: MockEvent
  draft_override?: RuleSet
  explain_only?: boolean
  explain_max_depth?: number
  explain_compact?: boolean
  explain_summary?: boolean
}

export interface SimulateRuleSetResponse {
  result: SimulationResult
}

export interface ValidationIssue {
  path: string
  message: string
}

export interface ValidationResult {
  valid: boolean
  issues?: ValidationIssue[]
  warnings?: ValidationIssue[]
}

export interface ValidateRuleSetResponse {
  result: ValidationResult
}

export type ProtocolFieldType = 'string' | 'number' | 'bool' | 'object' | 'array' | 'json'

export interface ProtocolFieldSpec {
  path: string
  type: ProtocolFieldType
  dynamic_path?: boolean
  operators?: string[]
}

export interface ProtocolSelectorSpec {
  path: string
  operators?: string[]
  dynamic_path?: boolean
}

export interface ProtocolSpec {
  name: string
  fields: ProtocolFieldSpec[]
  selectors?: ProtocolSelectorSpec[]
}

export type ListProtocolsResponse = ListResponse<ProtocolSpec>

export interface RollbackPreviewRequest {
  snapshot_id: string
  event?: MockEvent
  explain_only?: boolean
  explain_max_depth?: number
  explain_compact?: boolean
  explain_summary?: boolean
}

export interface FieldDiffSummary {
  path: string
  message?: string
  current?: unknown
  target?: unknown
}

export interface RuleDiffSummary {
  rule_id: string
  change_type: string
  condition_changed?: boolean
  action_changed?: boolean
  message?: string
  field_diffs?: FieldDiffSummary[]
}

export interface RollbackDiffSummary {
  current_snapshot_id?: string
  target_snapshot_id?: string
  changed: boolean
  messages?: string[]
  ruleset_field_diffs?: FieldDiffSummary[]
  rule_diffs?: RuleDiffSummary[]
}

export interface RollbackPreviewResponse {
  result: {
    snapshot: PublishedRuleSetSnapshot
    valid: boolean
    validation: ValidationResult
    diff: RollbackDiffSummary
    simulation?: SimulationResult
  }
}

export type ListRuleSetsResponse = ListResponse<RuleSet>
export type ListPublishedRuleSetsResponse = ListResponse<PublishedRuleSetSnapshot>
export type ListNamespacesResponse = ListResponse<NamespaceConfig>

export interface RuntimeMetrics {
  total_requests: number
  matched_requests: number
  unmatched_requests: number
  error_requests: number
  total_duration_ms?: number
  average_duration_ms: number
  ruleset_matches?: Record<string, number>
  rule_matches?: Record<string, number>
  last_matched_at_by_rule?: Record<string, string>
  fallback_reasons?: Record<string, number>
  status_codes?: Record<string, number>
  recent_requests?: RuntimeRequestRecord[]
}

export type RuntimeRequestOutcome = 'matched' | 'fallback' | 'unmatched' | 'error'

export interface RuntimeRequestRecord {
  id: number
  observed_at: string
  namespace?: string
  method?: string
  scheme?: string
  host?: string
  path?: string
  raw_query?: string
  trace_id?: string
  outcome?: RuntimeRequestOutcome | string
  matched: boolean
  fallback?: boolean
  error?: boolean
  status?: number
  ruleset_id?: string
  rule_id?: string
  fallback_reason?: string
  message?: string
  duration_ms: number
  event?: MockEvent
}
