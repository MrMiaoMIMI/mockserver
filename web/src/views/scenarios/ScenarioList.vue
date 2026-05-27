<template>
  <PageContainer title="Scenarios" eyebrow="agent overlay">
    <template #meta>
      <span class="meta-pill">
        <span class="status-dot is-on" />
        {{ activeCount }} active
      </span>
      <span class="meta-pill">{{ expiredCount }} expired</span>
      <span class="meta-pill">{{ totalRuleCount }} rules</span>
    </template>
    <template #actions>
      <el-button :icon="Refresh" :loading="store.loading" @click="loadScenarios">Refresh</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreateScenario">Create scenario</el-button>
    </template>

    <div class="scenario-workbench">
      <section class="scenario-index panel-surface">
        <header class="panel-heading">
          <div>
            <span>scenario inventory</span>
            <strong>{{ filteredScenarios.length }} visible</strong>
          </div>
          <el-switch
            v-model="filters.includeExpired"
            active-text="All"
            inactive-text="Active only"
            @change="loadScenarios"
          />
        </header>

        <div class="scenario-search">
          <el-input
            v-model="filters.query"
            clearable
            :prefix-icon="Search"
            placeholder="Search id, name, description"
          />
        </div>

        <el-empty
          v-if="!filteredScenarios.length && !store.loading"
          description="No scenarios match the current filters"
        />
        <div v-else class="scenario-list">
          <article
            v-for="scenario in filteredScenarios"
            :key="scenario.id"
            :class="['scenario-row', { active: selectedScenario?.id === scenario.id }]"
            tabindex="0"
            @click="selectScenario(scenario)"
            @keydown.enter.prevent="selectScenario(scenario)"
            @keydown.space.prevent="selectScenario(scenario)"
          >
            <div class="scenario-row-main">
              <div>
                <strong :title="scenario.name || scenario.id">{{ scenario.name || scenario.id }}</strong>
                <code :title="scenario.id">{{ scenario.id }}</code>
              </div>
              <StateChip :label="scenarioStateLabel(scenario)" :tone="scenarioTone(scenario)" />
            </div>
            <div class="scenario-row-facts">
              <span>{{ ttlLabel(scenario) }}</span>
              <span>{{ scenario.rule_count || 0 }} rules</span>
            </div>
          </article>
        </div>
      </section>

      <section class="scenario-detail panel-surface">
        <el-empty
          v-if="!selectedScenario"
          description="Select a scenario to inspect rules, traffic, simulator, and setup"
        />
        <template v-else>
          <header class="detail-heading">
            <div class="detail-title">
              <StateChip
                :label="scenarioStateLabel(selectedScenario)"
                :tone="scenarioTone(selectedScenario)"
                dot
              />
              <div>
                <strong>{{ selectedScenario.name || selectedScenario.id }}</strong>
                <button type="button" class="copy-id" @click="copyText(selectedScenario.id)">
                  <code>{{ selectedScenario.id }}</code>
                  <el-icon><CopyDocument /></el-icon>
                </button>
              </div>
            </div>
            <div class="detail-actions">
              <el-button :icon="Edit" @click="openEditScenario">Edit</el-button>
              <el-button :icon="AlarmClock" @click="extendScenario(3600)">Extend 1h</el-button>
              <el-button :icon="Refresh" @click="refreshSelected">Refresh</el-button>
              <el-button type="danger" :icon="Delete" @click="confirmDeleteScenario">
                Delete
              </el-button>
            </div>
          </header>

          <section class="detail-facts">
            <div>
              <span>Expires</span>
              <strong>{{ formatUnixMilli(selectedScenario.expire_time) }}</strong>
            </div>
            <div>
              <span>TTL</span>
              <strong>{{ ttlLabel(selectedScenario) }}</strong>
            </div>
            <div>
              <span>Rules</span>
              <strong>{{ store.scenarioRules.length }}</strong>
            </div>
            <div>
              <span>Traffic</span>
              <strong>{{ store.scenarioTrafficTotal }}</strong>
            </div>
          </section>

          <el-tabs v-model="activeTab" class="scenario-tabs">
            <el-tab-pane label="Rules" name="rules">
              <section class="tab-panel">
                <header class="tab-heading">
                  <div>
                    <span>scenario rules</span>
                    <strong>{{ store.scenarioRules.length }} configured</strong>
                  </div>
                  <el-button type="primary" :icon="Plus" @click="openRuleEditor()">
                    Create rule
                  </el-button>
                </header>

                <el-table :data="store.scenarioRules" height="330" row-key="rule_id">
                  <el-table-column label="Rule" min-width="190" show-overflow-tooltip>
                    <template #default="{ row }">
                      <div class="rule-name">
                        <strong>{{ row.name || row.rule_id }}</strong>
                        <code>{{ row.rule_id }}</code>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column label="Scope" min-width="150">
                    <template #default="{ row }">
                      <div class="rule-scope">
                        <StateChip :label="row.protocol" tone="primary" />
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column label="Match" min-width="220" show-overflow-tooltip>
                    <template #default="{ row }">
                      <code>{{ ruleMatchSummary(row) }}</code>
                    </template>
                  </el-table-column>
                  <el-table-column label="Response" width="120">
                    <template #default="{ row }">
                      <StateChip :label="ruleResponseSummary(row)" tone="ok" />
                    </template>
                  </el-table-column>
                  <el-table-column label="Action" width="150" fixed="right" align="right">
                    <template #default="{ row }">
                      <el-button size="small" :icon="Edit" @click="openRuleEditor(row)">
                        Edit
                      </el-button>
                      <el-button size="small" :icon="Delete" @click="confirmDeleteRule(row.rule_id)">
                        Delete
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </section>
            </el-tab-pane>

            <el-tab-pane label="Traffic" name="traffic">
              <section class="tab-panel">
                <header class="tab-heading">
                  <div>
                    <span>sdk traffic</span>
                    <strong>{{ store.scenarioTrafficTotal }} events</strong>
                  </div>
                  <el-button :icon="Refresh" @click="refreshTraffic">Refresh traffic</el-button>
                </header>
                <el-table :data="store.scenarioTrafficEvents" height="330" row-key="id">
                  <el-table-column label="Time" min-width="150">
                    <template #default="{ row }">
                      <div class="time-cell">
                        <strong>{{ formatUnixSecond(row.event_time) }}</strong>
                        <small>{{ row.duration_ms }}ms</small>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column label="Outcome" width="110">
                    <template #default="{ row }">
                      <StateChip :label="row.outcome" :tone="trafficTone(row.outcome)" />
                    </template>
                  </el-table-column>
                  <el-table-column label="Protocol" width="100">
                    <template #default="{ row }">
                      <code>{{ row.protocol_name || '-' }}</code>
                    </template>
                  </el-table-column>
                  <el-table-column label="Namespace" min-width="120">
                    <template #default="{ row }">
                      <code>{{ row.namespace_id || '-' }}</code>
                    </template>
                  </el-table-column>
                  <el-table-column label="Rule" min-width="190" show-overflow-tooltip>
                    <template #default="{ row }">
                      <div class="rule-name">
                        <strong>{{ row.rule_id || row.fallback_reason || '-' }}</strong>
                        <code>{{ row.ruleset_id || '-' }}</code>
                      </div>
                    </template>
                  </el-table-column>
                </el-table>
              </section>
            </el-tab-pane>

            <el-tab-pane label="Simulator" name="simulator">
              <section class="simulator-grid">
                <JsonEditor
                  v-model="simulateJSON"
                  title="Event"
                  :min-height="360"
                  :show-expand="false"
                />
                <JsonEditor
                  :model-value="simulationJSON"
                  title="Result"
                  readonly
                  :min-height="360"
                  :show-expand="false"
                  :show-format="false"
                >
                  <template #actions>
                    <el-button type="primary" size="small" :icon="CaretRight" @click="runSimulation">
                      Run
                    </el-button>
                  </template>
                </JsonEditor>
              </section>
            </el-tab-pane>

            <el-tab-pane label="Setup" name="setup">
              <section class="setup-grid">
                <div class="setup-block">
                  <span>Go context</span>
                  <pre><code>{{ goSetupSnippet }}</code></pre>
                </div>
                <div class="setup-block">
                  <span>HTTP header</span>
                  <pre><code>X-Mockserver-Scenario-ID: {{ selectedScenario.id }}</code></pre>
                </div>
                <div class="setup-block">
                  <span>MCP endpoint</span>
                  <pre><code>/mockserver/api/v1/agent/mcp</code></pre>
                </div>
              </section>
            </el-tab-pane>
          </el-tabs>
        </template>
      </section>
    </div>

    <el-dialog v-model="createVisible" title="Create scenario" width="560px" append-to-body>
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="Scenario ID">
          <el-input
            v-model="scenarioForm.scenario_id"
            placeholder="scn_checkout_timeout or empty for generated id"
          />
        </el-form-item>
        <el-form-item label="Name">
          <el-input v-model="scenarioForm.name" placeholder="Checkout timeout" />
        </el-form-item>
        <el-form-item label="Description">
          <el-input v-model="scenarioForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="TTL">
          <el-select v-model="scenarioForm.ttl_seconds">
            <el-option label="30 minutes" :value="1800" />
            <el-option label="1 hour" :value="3600" />
            <el-option label="4 hours" :value="14400" />
            <el-option label="1 day" :value="86400" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="store.saving" @click="submitScenario">
          Save
        </el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="ruleEditorVisible"
      :title="ruleEditorTitle"
      size="min(980px, calc(100vw - 24px))"
      append-to-body
      destroy-on-close
    >
      <div class="scenario-rule-editor">
        <section class="rule-scope-bar">
          <el-form label-position="top" class="rule-scope-form">
            <el-form-item label="Protocol">
              <el-select
                v-model="ruleScope.protocol"
                filterable
                :disabled="ruleEditorMode === 'edit'"
              >
                <el-option
                  v-for="protocol in protocolOptions"
                  :key="protocol.value"
                  :label="protocol.label"
                  :value="protocol.value"
                />
              </el-select>
            </el-form-item>
          </el-form>
        </section>
        <RuleEditorWorkbench
          :rule-set="scenarioRuleSet"
          :rule="editingScenarioRule?.rule || null"
          :mode="ruleEditorMode"
          :seed-rule="scenarioSeedRule"
          :save-rule="saveScenarioRule"
          :enable-preview="false"
          :enable-advanced="false"
          :allowed-action-types="['static']"
          @updated="handleScenarioRuleUpdated"
          @cancel="ruleEditorVisible = false"
        />
      </div>
    </el-drawer>
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  AlarmClock,
  CaretRight,
  CopyDocument,
  Delete,
  Edit,
  Plus,
  Refresh,
  Search,
} from '@element-plus/icons-vue'
import JsonEditor from '@/components/common/JsonEditor.vue'
import PageContainer from '@/components/common/PageContainer.vue'
import StateChip from '@/components/common/StateChip.vue'
import RuleEditorWorkbench from '@/components/rulesets/RuleEditorWorkbench.vue'
import { useMockserverStore } from '@/store'
import { actionSummary, conditionSummary } from '@/utils/ruleSummaries'
import type { Rule, RuleSet, Scenario, ScenarioRule } from '@/types'
import type { RuleEditorMode } from '@/utils/ruleFormAdapter'

type ChipTone = 'neutral' | 'ok' | 'warn' | 'caution' | 'danger' | 'accent' | 'primary'

const store = useMockserverStore()
const now = ref(nowMilliseconds())
const filters = reactive({
  query: '',
  includeExpired: true,
})
const activeTab = ref('rules')
const selectedId = ref('')
const createVisible = ref(false)
const ruleEditorVisible = ref(false)
const ruleEditorMode = ref<RuleEditorMode>('create')
const editingScenarioRule = ref<ScenarioRule | null>(null)
const scenarioForm = reactive({
  scenario_id: '',
  name: '',
  description: '',
  ttl_seconds: 3600,
})
const ruleScope = reactive({
  protocol: 'http',
})
const simulateJSON = ref(defaultSimulationJSON())
const simulationJSON = ref('{}')

const selectedScenario = computed(() => {
  if (store.currentScenario?.id === selectedId.value) return store.currentScenario
  return store.scenarioMap.get(selectedId.value) || null
})
const filteredScenarios = computed(() => {
  const query = filters.query.trim().toLowerCase()
  if (!query) return store.scenarios
  return store.scenarios.filter((scenario) =>
    [scenario.id, scenario.name, scenario.description].some((value) =>
      String(value || '').toLowerCase().includes(query)
    )
  )
})
const activeCount = computed(() => store.scenarios.filter((item) => !isExpired(item)).length)
const expiredCount = computed(() => store.scenarios.filter((item) => isExpired(item)).length)
const totalRuleCount = computed(() =>
  store.scenarios.reduce((total, item) => total + (item.rule_count || 0), 0)
)
const protocolOptions = computed(() => {
  const options = store.protocols.map((protocol) => ({
    label: protocol.name.toUpperCase(),
    value: protocol.name,
  }))
  if (ruleScope.protocol && !options.some((option) => option.value === ruleScope.protocol)) {
    options.unshift({ label: ruleScope.protocol.toUpperCase(), value: ruleScope.protocol })
  }
  return options.length ? options : [{ label: 'HTTP', value: 'http' }]
})
const scenarioRuleSet = computed<RuleSet | null>(() => {
  if (!selectedScenario.value) return null
  return {
    id: selectedScenario.value.id,
    name: selectedScenario.value.name || selectedScenario.value.id,
    enabled: true,
    protocol: ruleScope.protocol || 'http',
    namespace: 'default',
    selector: {},
    rules: store.scenarioRules.map((item) => item.rule),
    version: selectedScenario.value.version,
  }
})
const scenarioSeedRule = computed<Rule | null>(() => {
  if (ruleEditorMode.value !== 'create') return null
  const protocol = ruleScope.protocol || 'http'
  const id = `${protocol}-${Date.now().toString(36)}`
  return {
    id,
    name: `${protocol.toUpperCase()} static response`,
    enabled: true,
    priority: nextScenarioRulePriority(),
    when: defaultRuleCondition(protocol),
    action: {
      type: 'respond',
      renderer: 'static',
      response: {
        protocol,
        payload: defaultResponsePayload(protocol) as Record<string, unknown>,
      },
    },
  }
})
const ruleEditorTitle = computed(() =>
  ruleEditorMode.value === 'edit' ? 'Edit scenario rule' : 'Create scenario rule'
)
const goSetupSnippet = computed(() => {
  const id = selectedScenario.value?.id || 'scn_example'
  return `ctx = mocksdk.WithScenarioID(ctx, "${id}")`
})

async function loadScenarios() {
  now.value = nowMilliseconds()
  await store.fetchScenarios({
    include_expired: filters.includeExpired,
    limit: 200,
  })
  if (!selectedId.value && store.scenarios.length) {
    await selectScenario(store.scenarios[0])
  } else if (selectedId.value && !store.scenarioMap.has(selectedId.value)) {
    selectedId.value = ''
  }
}

async function selectScenario(scenario: Scenario) {
  selectedId.value = scenario.id
  store.currentScenario = scenario
  await refreshSelected()
}

async function refreshSelected() {
  if (!selectedId.value) return
  await Promise.all([
    store.fetchScenario(selectedId.value),
    store.fetchScenarioRules(selectedId.value),
    store.fetchScenarioTraffic(selectedId.value),
  ])
  simulateJSON.value = defaultSimulationJSON()
}

async function refreshTraffic() {
  if (!selectedId.value) return
  await store.fetchScenarioTraffic(selectedId.value)
}

function openEditScenario() {
  if (!selectedScenario.value) return
  scenarioForm.scenario_id = selectedScenario.value.id
  scenarioForm.name = selectedScenario.value.name || ''
  scenarioForm.description = selectedScenario.value.description || ''
  scenarioForm.ttl_seconds = Math.max(
    300,
    Math.ceil((selectedScenario.value.expire_time - nowMilliseconds()) / 1000)
  )
  createVisible.value = true
}

function openCreateScenario() {
  resetScenarioForm()
  createVisible.value = true
}

async function submitScenario() {
  const id = scenarioForm.scenario_id.trim()
  if (id && !isValidScenarioID(id)) {
    ElMessage.error('Scenario ID must start with scn_ and contain only letters, numbers, _ or -')
    return
  }
  if (selectedScenario.value?.id === id) {
    const updated = await store.updateScenario(id, {
      name: scenarioForm.name.trim(),
      description: scenarioForm.description.trim(),
      ttl_seconds: scenarioForm.ttl_seconds,
    })
    selectedId.value = updated.id
  } else {
    const created = await store.createScenario({
      scenario_id: id || undefined,
      name: scenarioForm.name.trim() || undefined,
      description: scenarioForm.description.trim() || undefined,
      ttl_seconds: scenarioForm.ttl_seconds,
    })
    selectedId.value = created.id
  }
  createVisible.value = false
  resetScenarioForm()
  await refreshSelected()
}

async function extendScenario(ttlSeconds: number) {
  if (!selectedScenario.value) return
  await store.updateScenario(selectedScenario.value.id, { ttl_seconds: ttlSeconds })
  await refreshSelected()
}

async function confirmDeleteScenario() {
  if (!selectedScenario.value) return
  await ElMessageBox.confirm(
    `Delete scenario ${selectedScenario.value.id}? Its scenario rules will be removed as well.`,
    'Delete scenario',
    { type: 'warning' }
  )
  await store.deleteScenario(selectedScenario.value.id)
  if (store.scenarios.length) {
    await selectScenario(store.scenarios[0])
  }
}

function openRuleEditor(rule?: ScenarioRule) {
  if (!selectedScenario.value) return
  editingScenarioRule.value = rule || null
  ruleEditorMode.value = rule ? 'edit' : 'create'
  ruleScope.protocol = rule?.protocol || ruleScope.protocol || protocolOptions.value[0]?.value || 'http'
  ruleEditorVisible.value = true
}

async function saveScenarioRule(rule: Rule, mode: RuleEditorMode, lockedRuleId?: string) {
  if (!selectedScenario.value || !scenarioRuleSet.value) {
    throw new Error('Select a scenario before saving a rule')
  }
  const ruleId = mode === 'edit' ? lockedRuleId || editingScenarioRule.value?.rule_id || rule.id : rule.id
  const saved = await store.upsertScenarioRule(selectedScenario.value.id, ruleId, {
    protocol: ruleScope.protocol,
    rule: { ...rule, id: ruleId },
  })
  return {
    ...scenarioRuleSet.value,
    rules: store.scenarioRules.map((item) =>
      item.rule_id === saved.rule_id ? saved.rule : item.rule
    ),
  }
}

async function handleScenarioRuleUpdated() {
  ruleEditorVisible.value = false
  editingScenarioRule.value = null
  await refreshSelected()
}

async function confirmDeleteRule(ruleID: string) {
  if (!selectedScenario.value) return
  await ElMessageBox.confirm(`Delete scenario rule ${ruleID}?`, 'Delete rule', { type: 'warning' })
  await store.deleteScenarioRule(selectedScenario.value.id, ruleID)
}

async function runSimulation() {
  if (!selectedScenario.value) return
  try {
    const event = JSON.parse(simulateJSON.value)
    const result = await store.simulateScenario(selectedScenario.value.id, {
      event,
      explain_summary: true,
    })
    simulationJSON.value = JSON.stringify(result, null, 2)
  } catch (error) {
    ElMessage.error(`Simulation input is not valid: ${toErrorMessage(error)}`)
  }
}

function resetScenarioForm() {
  scenarioForm.scenario_id = ''
  scenarioForm.name = ''
  scenarioForm.description = ''
  scenarioForm.ttl_seconds = 3600
}

function isExpired(scenario: Scenario) {
  return scenario.expire_time <= now.value
}

function scenarioStateLabel(scenario: Scenario) {
  return isExpired(scenario) ? 'expired' : scenario.status
}

function scenarioTone(scenario: Scenario): ChipTone {
  if (isExpired(scenario)) return 'warn'
  return scenario.status === 'active' ? 'ok' : 'neutral'
}

function trafficTone(outcome: string): ChipTone {
  if (outcome === 'matched') return 'ok'
  if (outcome === 'fallback') return 'warn'
  if (outcome === 'error') return 'danger'
  return 'neutral'
}

function ttlLabel(scenario: Scenario) {
  const seconds = Math.ceil((scenario.expire_time - now.value) / 1000)
  if (seconds <= 0) return 'expired'
  if (seconds < 3600) return `${Math.ceil(seconds / 60)}m left`
  if (seconds < 86400) return `${Math.ceil(seconds / 3600)}h left`
  return `${Math.ceil(seconds / 86400)}d left`
}

function formatUnixMilli(value: number) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function formatUnixSecond(value: number) {
  if (!value) return '-'
  return new Date(value * 1000).toLocaleString()
}

function ruleMatchSummary(rule: ScenarioRule) {
  return conditionSummary(rule.rule.when)
}

function ruleResponseSummary(rule: ScenarioRule) {
  return actionSummary(rule.rule.action)
}

function defaultSimulationJSON() {
  const rule = store.scenarioRules[0]
  const protocol = rule?.protocol || ruleScope.protocol || protocolOptions.value[0]?.value || 'http'
  return JSON.stringify(
    {
      protocol,
      namespace: 'default',
      request: defaultSimulationRequest(protocol),
    },
    null,
    2
  )
}

function nextScenarioRulePriority() {
  return Math.max(0, ...store.scenarioRules.map((item) => item.priority || 0)) + 10
}

function defaultRuleCondition(protocol: string) {
  if (protocol === 'cache') {
    return { field: 'request.operation', op: 'exists' }
  }
  if (protocol === 'spex') {
    return { field: 'request.cmd', op: 'exists' }
  }
  return { field: 'request.path', op: 'prefix', value: '/' }
}

function defaultResponsePayload(protocol: string) {
  if (protocol === 'http') {
    return {
      status: 200,
      headers: {},
      body: { ok: true },
    }
  }
  if (protocol === 'cache') {
    return {
      hit: true,
      value: { ok: true },
    }
  }
  if (protocol === 'spex') {
    return {
      code: 0,
      resp: { ok: true },
    }
  }
  return { ok: true }
}

function defaultSimulationRequest(protocol: string) {
  if (protocol === 'cache') {
    return {
      operation: 'get',
      key: 'user:10001',
    }
  }
  if (protocol === 'spex') {
    return {
      cmd: 'service.demo',
      req: {
        id: 'demo',
      },
    }
  }
  return {
    method: 'GET',
    path: '/api/orders',
  }
}

function isValidScenarioID(value: string) {
  return /^scn_[A-Za-z0-9_-]{4,124}$/.test(value) && value.length <= 128
}

function nowMilliseconds() {
  return Date.now()
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}

async function copyText(value: string) {
  await navigator.clipboard.writeText(value)
  ElMessage.success('Copied')
}

onMounted(async () => {
  await Promise.all([store.fetchProtocols(), loadScenarios()])
})
</script>

<style lang="scss" scoped>
.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 5px 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-secondary);
  background: #ffffff;
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
}

.scenario-workbench {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(300px, 360px) minmax(0, 1fr);
  gap: var(--ms-space-4);
}

.panel-surface {
  min-height: 0;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: #ffffff;
  box-shadow: var(--ms-shadow-xs);
}

.panel-heading,
.tab-heading,
.detail-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);

  span {
    display: block;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-transform: uppercase;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
  }
}

.scenario-index,
.scenario-detail {
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.scenario-index {
  padding: var(--ms-space-4);
}

.scenario-search {
  margin: var(--ms-space-4) 0 var(--ms-space-3);
}

.scenario-list {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  overflow: auto;
  padding-right: 2px;
}

.scenario-row {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  cursor: pointer;
  background: #fbfdff;
  transition:
    border-color var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast),
    transform var(--ms-transition-fast);

  &:hover,
  &.active {
    border-color: rgba(37, 99, 235, 0.42);
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
    transform: translateY(-1px);
  }
}

.scenario-row-main {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-2);

  div {
    min-width: 0;
  }

  strong,
  code {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
  }

  code {
    margin-top: 3px;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.scenario-row-facts {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-2);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
}

.scenario-detail {
  padding: var(--ms-space-4);
}

.detail-heading {
  flex-wrap: wrap;
}

.detail-title {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-3);

  strong {
    display: block;
  }
}

.detail-actions {
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);
  flex-wrap: wrap;
}

.copy-id {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 0;
  border: 0;
  color: var(--ms-blue-500);
  background: transparent;
  cursor: pointer;
}

.detail-facts {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--ms-space-3);
  margin: var(--ms-space-4) 0;

  div {
    min-width: 0;
    padding: var(--ms-space-3);
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-md);
    background: var(--ms-panel-bg-soft);
  }

  span,
  strong {
    display: block;
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }

  strong {
    margin-top: var(--ms-space-1);
    overflow-wrap: anywhere;
  }
}

.scenario-tabs {
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tab-panel {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.rule-name {
  min-width: 0;

  strong,
  code {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.rule-scope,
.time-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.simulator-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: var(--ms-space-4);
}

.setup-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--ms-space-3);
}

.setup-block {
  min-width: 0;
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: #fbfdff;

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-transform: uppercase;
  }

  pre {
    margin: var(--ms-space-2) 0 0;
    overflow: auto;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
    white-space: pre-wrap;
  }
}

.scenario-rule-editor {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
  min-height: 100%;
}

.rule-scope-bar {
  padding: var(--ms-space-4);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: #fbfdff;
}

.rule-scope-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(180px, 260px));
  gap: var(--ms-space-3);

  :deep(.el-form-item) {
    margin-bottom: 0;
  }
}

@media (max-width: 1180px) {
  .scenario-workbench {
    grid-template-columns: 1fr;
  }

  .scenario-index {
    max-height: 380px;
  }

  .detail-facts,
  .setup-grid,
  .simulator-grid,
  .rule-scope-form {
    grid-template-columns: 1fr;
  }
}
</style>
