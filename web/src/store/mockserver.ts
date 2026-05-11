import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import { mockserverApi } from '@/api'
import type {
  ListPublishedRuleSetsResponse,
  ListNamespacesResponse,
  ListProtocolsResponse,
  ListRuleSetsResponse,
  NamespaceConfig,
  ProtocolSpec,
  PublishedRuleSetSnapshot,
  Rule,
  RuleSet,
  RuntimeMetrics,
  SimulateRuleSetRequest,
  SimulationResult,
  ValidateRuleSetResponse,
} from '@/types'

export const useMockserverStore = defineStore('mockserver', () => {
  const drafts = ref<RuleSet[]>([])
  const namespaces = ref<NamespaceConfig[]>([])
  const protocols = ref<ProtocolSpec[]>([])
  const published = ref<PublishedRuleSetSnapshot[]>([])
  const snapshots = ref<PublishedRuleSetSnapshot[]>([])
  const currentDraft = ref<RuleSet | null>(null)
  const validation = ref<ValidateRuleSetResponse | null>(null)
  const simulation = ref<SimulationResult | null>(null)
  const metrics = ref<RuntimeMetrics | null>(null)
  const loading = ref(false)
  const saving = ref(false)

  const draftMap = computed(() => {
    const result = new Map<string, RuleSet>()
    drafts.value.forEach((item) => result.set(item.id, item))
    return result
  })

  const namespaceMap = computed(() => {
    const result = new Map<string, NamespaceConfig>()
    namespaces.value.forEach((item) => result.set(item.id, item))
    return result
  })

  const protocolMap = computed(() => {
    const result = new Map<string, ProtocolSpec>()
    protocols.value.forEach((item) => result.set(item.name, item))
    return result
  })

  async function fetchDrafts() {
    loading.value = true
    try {
      const response: ListRuleSetsResponse = await mockserverApi.listDrafts()
      drafts.value = response.items
      return response
    } finally {
      loading.value = false
    }
  }

  async function fetchNamespaces() {
    loading.value = true
    try {
      const response: ListNamespacesResponse = await mockserverApi.listNamespaces()
      namespaces.value = response.items
      return response
    } finally {
      loading.value = false
    }
  }

  async function fetchProtocols() {
    const response: ListProtocolsResponse = await mockserverApi.listProtocols()
    protocols.value = response.items
    return response
  }

  async function saveNamespace(data: NamespaceConfig) {
    saving.value = true
    try {
      const item = await mockserverApi.saveNamespace(data)
      const index = namespaces.value.findIndex((namespace) => namespace.id === item.id)
      if (index >= 0) {
        namespaces.value[index] = item
      } else {
        namespaces.value.unshift(item)
      }
      ElMessage.success('Namespace saved')
      return item
    } finally {
      saving.value = false
    }
  }

  async function fetchPublished() {
    loading.value = true
    try {
      const response: ListPublishedRuleSetsResponse = await mockserverApi.listPublished()
      published.value = response.items
      return response
    } finally {
      loading.value = false
    }
  }

  async function fetchDraft(id: string) {
    const item = await mockserverApi.getDraft(id)
    currentDraft.value = item
    const index = drafts.value.findIndex((draft) => draft.id === id)
    if (index >= 0) {
      drafts.value[index] = item
    } else {
      drafts.value.unshift(item)
    }
    return item
  }

  async function saveDraft(data: RuleSet) {
    saving.value = true
    try {
      const item = await mockserverApi.upsertDraft(data)
      currentDraft.value = item
      const index = drafts.value.findIndex((draft) => draft.id === item.id)
      if (index >= 0) {
        drafts.value[index] = item
      } else {
        drafts.value.unshift(item)
      }
      ElMessage.success('Ruleset saved')
      return item
    } finally {
      saving.value = false
    }
  }

  function updateDraftCache(item: RuleSet) {
    currentDraft.value = item
    const index = drafts.value.findIndex((draft) => draft.id === item.id)
    if (index >= 0) {
      drafts.value[index] = item
    } else {
      drafts.value.unshift(item)
    }
  }

  async function addRule(id: string, rule: Rule) {
    const item = await mockserverApi.addRule(id, rule)
    updateDraftCache(item)
    ElMessage.success('Rule added')
    return item
  }

  async function updateRule(id: string, ruleId: string, rule: Rule) {
    const item = await mockserverApi.updateRule(id, ruleId, rule)
    updateDraftCache(item)
    ElMessage.success('Rule updated')
    return item
  }

  async function deleteRule(id: string, ruleId: string) {
    const item = await mockserverApi.deleteRule(id, ruleId)
    updateDraftCache(item)
    ElMessage.success('Rule deleted')
    return item
  }

  async function setRuleEnabled(id: string, ruleId: string, enabled: boolean) {
    const item = await mockserverApi.setRuleEnabled(id, ruleId, enabled)
    updateDraftCache(item)
    ElMessage.success(enabled ? 'Rule enabled' : 'Rule disabled')
    return item
  }

  async function setRulePriority(id: string, ruleId: string, priority: number) {
    const item = await mockserverApi.setRulePriority(id, ruleId, priority)
    updateDraftCache(item)
    ElMessage.success('Rule priority updated')
    return item
  }

  async function validateDraft(id: string) {
    validation.value = await mockserverApi.validateDraft(id)
    if (validation.value.result.valid) {
      ElMessage.success('Ruleset validation passed')
    } else {
      ElMessage.warning('Ruleset validation failed')
    }
    return validation.value
  }

  async function publishDraft(id: string, reason?: string) {
    const result = await mockserverApi.publish(id, reason)
    await Promise.all([fetchDrafts(), fetchPublished(), fetchSnapshots(id)])
    ElMessage.success('Ruleset published')
    return result
  }

  async function fetchSnapshots(id: string) {
    const response = await mockserverApi.listSnapshots(id)
    snapshots.value = response.items
    return response
  }

  async function rollback(id: string, snapshotId: string, reason?: string) {
    const result = await mockserverApi.rollback(id, snapshotId, reason)
    await Promise.all([fetchPublished(), fetchSnapshots(id)])
    ElMessage.success('Rolled back to target snapshot')
    return result
  }

  async function simulateDraft(id: string, request: SimulateRuleSetRequest) {
    const response = await mockserverApi.simulateDraft(id, request)
    simulation.value = response.result
    return response.result
  }

  async function simulatePublished(request: SimulateRuleSetRequest) {
    const response = await mockserverApi.simulatePublished(request)
    simulation.value = response.result
    return response.result
  }

  async function fetchMetrics() {
    metrics.value = await mockserverApi.runtimeMetrics()
    return metrics.value
  }

  function setCurrentDraft(ruleSet: RuleSet | null) {
    currentDraft.value = ruleSet
    validation.value = null
    simulation.value = null
  }

  return {
    drafts,
    namespaces,
    protocols,
    published,
    snapshots,
    currentDraft,
    validation,
    simulation,
    metrics,
    loading,
    saving,
    draftMap,
    namespaceMap,
    protocolMap,
    fetchDrafts,
    fetchNamespaces,
    fetchProtocols,
    saveNamespace,
    fetchPublished,
    fetchDraft,
    saveDraft,
    addRule,
    updateRule,
    deleteRule,
    setRuleEnabled,
    setRulePriority,
    validateDraft,
    publishDraft,
    fetchSnapshots,
    rollback,
    simulateDraft,
    simulatePublished,
    fetchMetrics,
    setCurrentDraft,
  }
})
