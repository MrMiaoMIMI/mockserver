import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import { mockserverApi } from '@/api'
import type {
  CreateScenarioRequest,
  HTTPQuickRuleRequest,
  ListPublishedRuleSetsResponse,
  ListNamespacesResponse,
  ListProtocolsResponse,
  ListRuleSetsResponse,
  ListScenarioRulesResponse,
  ListScenariosResponse,
  ListTrafficEventsResponse,
  NamespaceConfig,
  ProtocolSpec,
  PublishedRuleSetSnapshot,
  Rule,
  Scenario,
  ScenarioRule,
  RuleSet,
  RuntimeMetrics,
  SimulateScenarioRequest,
  SimulateRuleSetRequest,
  SimulationResult,
  TrafficEvent,
  TrafficQueryParams,
  TrafficStats,
  UpdateScenarioRequest,
  UpsertScenarioRuleRequest,
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
  const trafficEvents = ref<TrafficEvent[]>([])
  const trafficStats = ref<TrafficStats | null>(null)
  const trafficTotal = ref(0)
  const scenarios = ref<Scenario[]>([])
  const currentScenario = ref<Scenario | null>(null)
  const scenarioRules = ref<ScenarioRule[]>([])
  const scenarioTrafficEvents = ref<TrafficEvent[]>([])
  const scenarioTrafficStats = ref<TrafficStats | null>(null)
  const scenarioTrafficTotal = ref(0)
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

  const scenarioMap = computed(() => {
    const result = new Map<string, Scenario>()
    scenarios.value.forEach((item) => result.set(item.id, item))
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

  async function saveNamespace(data: NamespaceConfig, update = false) {
    saving.value = true
    try {
      const item = await mockserverApi.saveNamespace(data, update)
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

  async function fetchTrafficEvents(query: TrafficQueryParams = {}) {
    loading.value = true
    try {
      const response: ListTrafficEventsResponse = await mockserverApi.listTrafficEvents(query)
      trafficEvents.value = response.items
      trafficStats.value = response.stats
      trafficTotal.value = response.total
      return response
    } finally {
      loading.value = false
    }
  }

  async function fetchTrafficSummary(query: TrafficQueryParams = {}) {
    const response: ListTrafficEventsResponse = await mockserverApi.listTrafficEvents({
      limit: 1,
      ...query,
    })
    trafficStats.value = response.stats
    trafficTotal.value = response.total
    return response
  }

  async function fetchTrafficEvent(id: number) {
    return mockserverApi.getTrafficEvent(id)
  }

  async function fetchScenarios(params: Parameters<typeof mockserverApi.listScenarios>[0] = {}) {
    loading.value = true
    try {
      const response: ListScenariosResponse = await mockserverApi.listScenarios(params)
      scenarios.value = response.items
      return response
    } finally {
      loading.value = false
    }
  }

  function updateScenarioCache(item: Scenario) {
    currentScenario.value = item
    const index = scenarios.value.findIndex((scenario) => scenario.id === item.id)
    if (index >= 0) {
      scenarios.value[index] = item
    } else {
      scenarios.value.unshift(item)
    }
  }

  async function createScenario(data: CreateScenarioRequest) {
    saving.value = true
    try {
      const item = await mockserverApi.createScenario(data)
      updateScenarioCache(item)
      ElMessage.success('Scenario created')
      return item
    } finally {
      saving.value = false
    }
  }

  async function fetchScenario(id: string) {
    const item = await mockserverApi.getScenario(id)
    updateScenarioCache(item)
    return item
  }

  async function updateScenario(id: string, data: UpdateScenarioRequest) {
    saving.value = true
    try {
      const item = await mockserverApi.updateScenario(id, data)
      updateScenarioCache(item)
      ElMessage.success('Scenario updated')
      return item
    } finally {
      saving.value = false
    }
  }

  async function deleteScenario(id: string) {
    await mockserverApi.deleteScenario(id)
    scenarios.value = scenarios.value.filter((scenario) => scenario.id !== id)
    if (currentScenario.value?.id === id) {
      currentScenario.value = null
      scenarioRules.value = []
      scenarioTrafficEvents.value = []
      scenarioTrafficStats.value = null
      scenarioTrafficTotal.value = 0
    }
    ElMessage.success('Scenario deleted')
  }

  async function fetchScenarioRules(id: string) {
    const response: ListScenarioRulesResponse = await mockserverApi.listScenarioRules(id)
    scenarioRules.value = response.items
    const cached = scenarioMap.value.get(id)
    if (cached) {
      updateScenarioCache({ ...cached, rule_count: response.total })
    }
    return response
  }

  async function upsertHTTPQuickRule(id: string, data: HTTPQuickRuleRequest) {
    const item = await mockserverApi.upsertHTTPQuickRule(id, data)
    await fetchScenarioRules(id)
    ElMessage.success('Scenario rule saved')
    return item
  }

  async function upsertScenarioRule(id: string, ruleId: string, data: UpsertScenarioRuleRequest) {
    const item = await mockserverApi.upsertScenarioRule(id, ruleId, data)
    await fetchScenarioRules(id)
    ElMessage.success('Scenario rule saved')
    return item
  }

  async function deleteScenarioRule(id: string, ruleId: string) {
    await mockserverApi.deleteScenarioRule(id, ruleId)
    await fetchScenarioRules(id)
    ElMessage.success('Scenario rule deleted')
  }

  async function simulateScenario(id: string, request: SimulateScenarioRequest) {
    const response = await mockserverApi.simulateScenario(id, request)
    simulation.value = response.result
    return response.result
  }

  async function fetchScenarioTraffic(id: string, query: TrafficQueryParams = {}) {
    const response = await mockserverApi.listScenarioTraffic(id, {
      limit: 25,
      ...query,
    })
    scenarioTrafficEvents.value = response.items
    scenarioTrafficStats.value = response.stats
    scenarioTrafficTotal.value = response.total
    return response
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
    trafficEvents,
    trafficStats,
    trafficTotal,
    scenarios,
    currentScenario,
    scenarioRules,
    scenarioTrafficEvents,
    scenarioTrafficStats,
    scenarioTrafficTotal,
    loading,
    saving,
    draftMap,
    namespaceMap,
    protocolMap,
    scenarioMap,
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
    fetchTrafficEvents,
    fetchTrafficSummary,
    fetchTrafficEvent,
    fetchScenarios,
    createScenario,
    fetchScenario,
    updateScenario,
    deleteScenario,
    fetchScenarioRules,
    upsertHTTPQuickRule,
    upsertScenarioRule,
    deleteScenarioRule,
    simulateScenario,
    fetchScenarioTraffic,
    setCurrentDraft,
  }
})
