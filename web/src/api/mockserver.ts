import { http } from '@/utils/request'
import type {
  ListPublishedRuleSetsResponse,
  ListNamespacesResponse,
  ListProtocolsResponse,
  ListRuleSetsResponse,
  NamespaceConfig,
  PublishedRuleSetResponse,
  PublishRuleSetResponse,
  RollbackPreviewRequest,
  RollbackPreviewResponse,
  RollbackRuleSetResponse,
  Rule,
  RuleSet,
  RuntimeMetrics,
  SimulateRuleSetRequest,
  SimulateRuleSetResponse,
  ValidateRuleSetResponse,
} from '@/types'

const adminBase = '/mockserver/api/v1/admin'

export const mockserverApi = {
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
  saveNamespace(data: NamespaceConfig): Promise<NamespaceConfig> {
    if (data.id) {
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
}
