import { http } from '@/utils/request'
import { apiPath } from '@/utils/apiPrefix'
import type {
  DebugLoginRequest,
  CreateScenarioRequest,
  HTTPQuickRuleRequest,
  ListPublishedRuleSetsResponse,
  ListNamespacesResponse,
  ListProtocolsResponse,
  ListRuleSetsResponse,
  ListScenarioRulesResponse,
  ListScenariosResponse,
  ListTrafficEventsResponse,
  LoginResponse,
  NamespaceConfig,
  PublishedRuleSetResponse,
  PublishRuleSetResponse,
  RollbackPreviewRequest,
  RollbackPreviewResponse,
  RollbackRuleSetResponse,
  Rule,
  Scenario,
  ScenarioRule,
  RuleSet,
  RuntimeMetrics,
  SimulateScenarioRequest,
  SimulateScenarioResponse,
  SimulateRuleSetRequest,
  SimulateRuleSetResponse,
  TrafficEvent,
  TrafficQueryParams,
  UpdateScenarioRequest,
  UpsertScenarioRuleRequest,
  UserInfo,
  ValidateRuleSetResponse,
} from '@/types'

const adminBase = apiPath('/mockserver/api/v1/admin')
const authBase = apiPath('/mockserver/api/v1/auth')
const agentBase = apiPath('/mockserver/api/v1/agent')

export const mockserverApi = {
  debugLogin(data: DebugLoginRequest): Promise<LoginResponse> {
    return http.post(`${authBase}/debug/login`, data)
  },
  currentUser(): Promise<UserInfo> {
    return http.get(`${authBase}/me`)
  },
  listDrafts(): Promise<ListRuleSetsResponse> {
    return http.get(`${adminBase}/rulesets`)
  },
  listProtocols(): Promise<ListProtocolsResponse> {
    return http.get(`${adminBase}/protocols`)
  },
  listNamespaces(): Promise<ListNamespacesResponse> {
    return http.get(`${adminBase}/namespaces`)
  },
  getNamespace(id: string): Promise<NamespaceConfig> {
    return http.get(`${adminBase}/namespaces/${encodeURIComponent(id)}`)
  },
  saveNamespace(data: NamespaceConfig, update = false): Promise<NamespaceConfig> {
    if (update) {
      return http.put(`${adminBase}/namespaces/${encodeURIComponent(data.id)}`, data)
    }
    return http.post(`${adminBase}/namespaces`, data)
  },
  getDraft(id: string): Promise<RuleSet> {
    return http.get(`${adminBase}/rulesets/${encodeURIComponent(id)}`)
  },
  upsertDraft(data: RuleSet): Promise<RuleSet> {
    return http.post(`${adminBase}/rulesets`, data)
  },
  validateDraft(id: string): Promise<ValidateRuleSetResponse> {
    return http.post(`${adminBase}/rulesets/${encodeURIComponent(id)}/validate`)
  },
  publish(id: string, reason?: string): Promise<PublishRuleSetResponse> {
    return http.post(`${adminBase}/rulesets/${encodeURIComponent(id)}/publish`, { reason })
  },
  simulateDraft(id: string, data: SimulateRuleSetRequest): Promise<SimulateRuleSetResponse> {
    return http.post(`${adminBase}/rulesets/${encodeURIComponent(id)}/simulate`, data)
  },
  addRule(id: string, rule: Rule): Promise<RuleSet> {
    return http.post(`${adminBase}/rulesets/${encodeURIComponent(id)}/rules`, { rule })
  },
  updateRule(id: string, ruleId: string, rule: Rule): Promise<RuleSet> {
    return http.put(
      `${adminBase}/rulesets/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}`,
      { rule }
    )
  },
  deleteRule(id: string, ruleId: string): Promise<RuleSet> {
    return http.delete(
      `${adminBase}/rulesets/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}`
    )
  },
  setRuleEnabled(id: string, ruleId: string, enabled: boolean): Promise<RuleSet> {
    const action = enabled ? 'enable' : 'disable'
    return http.post(
      `${adminBase}/rulesets/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}/${action}`
    )
  },
  setRulePriority(id: string, ruleId: string, priority: number): Promise<RuleSet> {
    return http.post(
      `${adminBase}/rulesets/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}/priority`,
      { priority }
    )
  },
  listPublished(): Promise<ListPublishedRuleSetsResponse> {
    return http.get(`${adminBase}/published/rulesets`)
  },
  getPublished(id: string): Promise<PublishedRuleSetResponse> {
    return http.get(`${adminBase}/published/rulesets/${encodeURIComponent(id)}`)
  },
  listSnapshots(id: string): Promise<ListPublishedRuleSetsResponse> {
    return http.get(`${adminBase}/published/rulesets/${encodeURIComponent(id)}/snapshots`)
  },
  rollback(id: string, snapshotId: string, reason?: string): Promise<RollbackRuleSetResponse> {
    return http.post(`${adminBase}/published/rulesets/${encodeURIComponent(id)}/rollback`, {
      snapshot_id: snapshotId,
      reason,
    })
  },
  rollbackPreview(id: string, data: RollbackPreviewRequest): Promise<RollbackPreviewResponse> {
    return http.post(
      `${adminBase}/published/rulesets/${encodeURIComponent(id)}/rollback/preview`,
      data
    )
  },
  simulatePublished(data: SimulateRuleSetRequest): Promise<SimulateRuleSetResponse> {
    return http.post(`${adminBase}/published/simulate`, data)
  },
  runtimeMetrics(): Promise<RuntimeMetrics> {
    return http.get(`${adminBase}/metrics/runtime`)
  },
  listTrafficEvents(params: TrafficQueryParams = {}): Promise<ListTrafficEventsResponse> {
    return http.get(`${adminBase}/traffic/events`, { params })
  },
  getTrafficEvent(id: number): Promise<TrafficEvent> {
    return http.get(`${adminBase}/traffic/events/${encodeURIComponent(String(id))}`)
  },
  listScenarios(params: {
    status?: string
    include_expired?: boolean
    limit?: number
    offset?: number
  } = {}): Promise<ListScenariosResponse> {
    return http.get(`${agentBase}/scenarios`, { params })
  },
  createScenario(data: CreateScenarioRequest = {}): Promise<Scenario> {
    return http.post(`${agentBase}/scenarios`, data)
  },
  getScenario(id: string): Promise<Scenario> {
    return http.get(`${agentBase}/scenarios/${encodeURIComponent(id)}`)
  },
  updateScenario(id: string, data: UpdateScenarioRequest): Promise<Scenario> {
    return http.patch(`${agentBase}/scenarios/${encodeURIComponent(id)}`, data)
  },
  deleteScenario(id: string): Promise<{ deleted: boolean }> {
    return http.delete(`${agentBase}/scenarios/${encodeURIComponent(id)}`)
  },
  listScenarioRules(id: string): Promise<ListScenarioRulesResponse> {
    return http.get(`${agentBase}/scenarios/${encodeURIComponent(id)}/rules`)
  },
  upsertHTTPQuickRule(id: string, data: HTTPQuickRuleRequest): Promise<ScenarioRule> {
    return http.post(`${agentBase}/scenarios/${encodeURIComponent(id)}/rules/quick`, data)
  },
  upsertScenarioRule(
    id: string,
    ruleId: string,
    data: UpsertScenarioRuleRequest
  ): Promise<ScenarioRule> {
    return http.put(
      `${agentBase}/scenarios/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}`,
      data
    )
  },
  deleteScenarioRule(id: string, ruleId: string): Promise<{ deleted: boolean }> {
    return http.delete(
      `${agentBase}/scenarios/${encodeURIComponent(id)}/rules/${encodeURIComponent(ruleId)}`
    )
  },
  simulateScenario(id: string, data: SimulateScenarioRequest): Promise<SimulateScenarioResponse> {
    return http.post(`${agentBase}/scenarios/${encodeURIComponent(id)}/simulate`, data)
  },
  listScenarioTraffic(id: string, params: TrafficQueryParams = {}): Promise<ListTrafficEventsResponse> {
    return http.get(`${agentBase}/scenarios/${encodeURIComponent(id)}/traffic`, { params })
  },
}
