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
  renderer?: string
  response?: ProtocolResponse
  response_template?: string
  response_expression?: string
  sequence?: SequenceStep[]
  sequence_strategy?: string
  webhook?: WebhookConfig
  forward?: NamespaceForwardFallback
}

export interface ProtocolResponse {
  protocol?: string
  payload?: Record<string, unknown>
}

export interface SequenceStep {
  response: ProtocolResponse
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
  authoring?: RuleAuthoring
}

export interface RuleAuthoring {
  sample_request?: SampleRequestAuthoring
}

export type SampleRequestFormat = 'json' | 'curl'

export interface SampleRequestAuthoring {
  format: SampleRequestFormat
  root?: string
  raw: string
  parsed?: unknown
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

export interface DebugLoginRequest {
  email: string
}

export interface UserInfo {
  email: string
}

export interface LoginResponse {
  token: string
  user: UserInfo
}

export type NamespaceFallbackType = 'respond' | 'forward'

export interface NamespaceForwardFallback {
  timeout_ms?: number
}

export interface NamespaceAction extends RuleAction {
  type: NamespaceFallbackType
}

export type NamespaceFallbackAction = NamespaceAction

export interface NamespacePolicy {
  ruleset_miss_action: NamespaceAction
  rule_miss_action: NamespaceAction
}

export interface NamespaceConfig {
  id: string
  name?: string
  description?: string
  policies: Record<string, NamespacePolicy>
  version?: number
}

export interface NamespaceFallbackForm {
  type: NamespaceFallbackType
  responsePayload: Record<string, unknown>
  responseFieldDrafts: Record<string, string>
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

export interface RuntimeDecision {
  kind: string
  matched: boolean
  fallback?: boolean
  protocol?: string
  trace: MatchTrace
  response?: ProtocolResponse
  forward?: {
    timeout_ms?: number
  }
  meta?: {
    trace_id?: string
  }
}

export interface SimulationResult {
  matched: boolean
  fallback?: boolean
  trace: MatchTrace
  candidates?: string[]
  response?: ProtocolResponse
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
  required?: boolean
  default?: unknown
  min?: number
  max?: number
  format?: string
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
  response?: {
    fields?: ProtocolFieldSpec[]
    defaults?: Record<string, unknown>
  }
  actions?: Array<{
    type: string
    renderers?: string[]
  }>
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
  sources?: Record<string, number>
  protocols?: Record<string, number>
  operations?: Record<string, number>
  ruleset_matches?: Record<string, number>
  rule_matches?: Record<string, number>
  last_matched_at_by_rule?: Record<string, string>
  fallback_reasons?: Record<string, number>
  fallback_stats?: Record<string, RuntimeFallbackStat>
  status_codes?: Record<string, number>
  recent_requests?: RuntimeRequestRecord[]
}

export interface RuntimeFallbackStat {
  total: number
  by_protocol?: Record<string, number>
  by_namespace?: Record<string, number>
  by_operation?: Record<string, number>
}

export type RuntimeRequestOutcome = 'matched' | 'fallback' | 'unmatched' | 'error'

export interface RuntimeRequestRecord {
  id: number
  observed_at: string
  source?: string
  protocol?: string
  operation?: string
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

export interface TrafficEventIndex {
  id: number
  traffic_event_id: number
  protocol_name: string
  field_path: string
  field_value_preview: string
  field_value_hash: number
  field_value_text?: string
  event_time: number
}

export interface TrafficEvent {
  id: number
  event_id: string
  trace_id?: string
  traffic_source: string
  protocol_name: string
  namespace_id: string
  operation_name?: string
  outcome: 'matched' | 'fallback' | 'unmatched' | 'error' | string
  decision_kind?: string
  ruleset_id?: string
  rule_id?: string
  snapshot_id?: string
  fallback_reason?: string
  duration_ms: number
  event_time: number
  expire_time: number
  event?: MockEvent
  decision?: RuntimeDecision | Record<string, unknown>
  explain?: Record<string, unknown>
  error_message?: string
  indexes?: TrafficEventIndex[]
}

export interface TrafficStats {
  total: number
  by_outcome: Record<string, number>
  by_protocol: Record<string, number>
  by_namespace: Record<string, number>
}

export interface ListTrafficEventsResponse {
  items: TrafficEvent[]
  total: number
  stats: TrafficStats
}

export interface TrafficQueryParams {
  limit?: number
  offset?: number
  start_time?: number
  end_time?: number
  event_id?: string
  trace_id?: string
  protocol_name?: string
  namespace_id?: string
  operation_name?: string
  outcome?: string
  decision_kind?: string
  ruleset_id?: string
  rule_id?: string
  fallback_reason?: string
  field_path?: string
  field_value?: string
}
