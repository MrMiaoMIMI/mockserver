<template>
  <PageContainer :title="currentDraft?.name || rulesetId || 'Ruleset'">
    <template #meta>
      <span class="meta-pill">
        <span class="status-dot" :class="{ 'is-on': currentDraft?.enabled }" />
        {{ draftStatusLabel }}
      </span>
      <span class="meta-pill">{{ currentDraft?.namespace || '-' }}</span>
      <span class="meta-pill">{{ rulesCountLabel }}</span>
      <span class="meta-pill">{{ publishedStateLabel }}</span>
    </template>

    <div class="workspace-page">
      <el-alert
        v-if="visibleError"
        class="workspace-alert"
        type="error"
        show-icon
        :closable="false"
        :title="visibleError"
      />

      <section v-if="isInitialLoading" class="workspace-state">
        <el-skeleton animated :rows="8" />
      </section>

      <section v-else-if="isNotFound" class="workspace-state">
        <el-empty :description="`Ruleset ${rulesetId} does not exist`">
          <el-button :icon="Back" @click="router.push('/rulesets')">Back to list</el-button>
        </el-empty>
      </section>

      <section v-else-if="!currentDraft" class="workspace-state">
        <el-empty description="Select a draft ruleset">
          <el-button :icon="Back" @click="router.push('/rulesets')">Back to list</el-button>
        </el-empty>
      </section>

      <template v-else>
        <section v-loading="pageLoading" class="workspace-context">
          <div class="context-identity">
            <span class="status-dot" :class="{ 'is-on': currentDraft.enabled }" />
            <div>
              <span>draft id</span>
              <strong :title="currentDraft.id">{{ currentDraft.id }}</strong>
            </div>
          </div>

          <div class="context-summary">
            <div class="context-cell is-selector is-compact">
              <span>selectors</span>
              <SelectorValueGroup :values="selectorValues" :max-visible="2" />
            </div>
          </div>

          <div class="context-actions">
            <el-button :icon="Setting" @click="settingsVisible = true">Settings</el-button>
            <el-button :icon="InfoFilled" @click="contextExpanded = !contextExpanded">
              {{ contextExpanded ? 'Hide details' : 'Details' }}
            </el-button>
          </div>

          <Transition name="context-details">
            <div v-if="contextExpanded" class="context-grid">
              <div class="context-cell is-wide">
                <span>ruleset id</span>
                <code :title="currentDraft.id">{{ currentDraft.id }}</code>
              </div>
              <div class="context-cell">
                <span>namespace</span>
                <strong>{{ currentDraft.namespace }}</strong>
              </div>
              <div class="context-cell">
                <span>version</span>
                <strong>{{ draftVersionLabel }}</strong>
                <small>{{ publishedVersionLabel }}</small>
              </div>
              <div class="context-cell">
                <span>published</span>
                <strong>{{ publishedStateLabel }}</strong>
                <small>{{ publishedDetailLabel }}</small>
              </div>
              <div class="context-cell">
                <span>snapshots</span>
                <strong>{{ snapshotStateLabel }}</strong>
                <small>{{ snapshotDetailLabel }}</small>
              </div>
              <div class="context-cell is-selector">
                <span>selectors</span>
                <SelectorValueGroup :values="selectorValues" :max-visible="3" />
              </div>
            </div>
          </Transition>
        </section>

        <section v-if="debugPayload" class="debug-source-banner">
          <div>
            <span>runtime source</span>
            <strong>{{ debugSourceSummary }}</strong>
            <small>{{ debugSourceDetail }}</small>
          </div>
          <div class="debug-source-actions">
            <el-button size="small" :icon="VideoPlay" @click="applyDebugSimulationPayload">
              Use in simulation
            </el-button>
            <el-button size="small" :icon="Setting" @click="applyDebugCreatePayload">
              Create seeded rule
            </el-button>
            <el-button size="small" @click="clearDebugPayload">Clear</el-button>
          </div>
        </section>

        <div class="rules-workspace">
          <main class="rules-main">
            <RuleManager
              :rule-set="currentDraft"
              :active-rule-id="activeRuleId"
              :diagnostics="ruleDiagnostics"
              @select="selectRule"
              @create="openCreateRuleEditor"
              @edit="openEditRuleEditor"
              @updated="applyDraftUpdate"
            />
          </main>

          <aside class="tool-dock">
            <header class="workbench-header">
              <div>
                <span>Rule Workspace</span>
                <strong>{{ workbenchTitle }}</strong>
                <small>{{ activeTaskContextLabel }}</small>
              </div>
              <TaskRail
                :model-value="activeWorkbenchIntent"
                :tasks="workbenchModes"
                aria-label="Ruleset workbench tasks"
                @update:model-value="updateWorkbenchIntent"
              />
            </header>

            <div v-if="activeWorkbenchIntent === 'test'" class="intent-subnav" aria-label="Test workflow">
              <button
                type="button"
                :class="{ active: activeWorkbenchMode === 'simulate' }"
                @click="showWorkbenchMode('simulate')"
              >
                Simulation
              </button>
              <button
                type="button"
                :class="{ active: activeWorkbenchMode === 'result' }"
                @click="showWorkbenchMode('result')"
              >
                Result
              </button>
            </div>

            <div v-if="activeWorkbenchIntent === 'release'" class="intent-subnav" aria-label="Release workflow">
              <button
                type="button"
                :class="{ active: activeWorkbenchMode === 'readiness' }"
                @click="showWorkbenchMode('readiness')"
              >
                Readiness
              </button>
              <button
                type="button"
                :class="{ active: activeWorkbenchMode === 'snapshots' }"
                @click="showWorkbenchMode('snapshots')"
              >
                Snapshots
              </button>
            </div>

            <section v-show="activeWorkbenchMode === 'inspect'" class="workbench-panel">
              <RuleWorkbenchOverview
                :rule-set="currentDraft"
                :rule="activeRule"
                :diagnostic="activeRuleDiagnostic"
                @create="openCreateRuleEditor"
                @edit="openEditRuleEditor"
                @simulate="showWorkbenchMode('simulate')"
                @validate="openReadinessAndValidate"
              />
            </section>

            <section v-show="activeWorkbenchMode === 'editor'" class="workbench-panel">
              <RuleEditorWorkbench
                :rule-set="currentDraft"
                :rule="editorRule"
                :mode="editorMode"
                :seed-rule="debugSeedRule"
                :seed-label="debugSeedLabel"
                @updated="applyRuleEditorUpdate"
                @cancel="showWorkbenchMode('inspect')"
              />
            </section>

            <section v-show="activeWorkbenchMode === 'readiness'" class="workbench-panel">
              <PublishReadinessWorkbench
                :rule-set="currentDraft"
                :current-published="currentPublished"
                :validation="currentValidation"
                :publish-result="lastPublishResult"
                :validate-error="validationError"
                :publish-error="publishError"
                :validating="operationLoading.validate"
                :publishing="operationLoading.publish"
                @validate="validateDraft"
                @publish="requestPublishDraft"
              />
            </section>

            <section v-show="activeWorkbenchMode === 'simulate'" class="workbench-panel">
              <div class="tool-stack">
                <div class="simulation-context">
                  <div>
                    <span>simulation context</span>
                    <strong>{{ currentDraft.namespace }} / {{ activeRuleId || 'no selected rule' }}</strong>
                    <small>{{ simulationEventStateLabel }}</small>
                  </div>
                  <el-button size="small" :icon="Refresh" @click="refreshSimulationEventFromContext(true)">
                    Refresh event
                  </el-button>
                </div>
                <div v-if="simulationWarnings.length" class="simulation-warning-list">
                  <header>
                    <span>runtime warnings</span>
                    <strong>{{ simulationWarnings.length }}</strong>
                  </header>
                  <small
                    v-for="warning in simulationWarnings"
                    :key="`${warning.path}-${warning.message}`"
                    :title="warning.message"
                  >
                    {{ warning.message }}
                  </small>
                </div>
                <div class="simulate-toolbar">
                  <div class="simulate-control">
                    <span>
                      Target
                      <el-tooltip
                        content="Draft uses the current editable ruleset. Published uses the latest published snapshot."
                        placement="top"
                      >
                        <el-icon class="tip-icon"><InfoFilled /></el-icon>
                      </el-tooltip>
                    </span>
                    <el-radio-group v-model="simulateTarget" size="small">
                      <el-radio-button value="draft">Draft</el-radio-button>
                      <el-radio-button value="published">Published</el-radio-button>
                    </el-radio-group>
                  </div>
                  <div class="simulate-toggle">
                    <el-checkbox v-model="simulateExplainOnly">Explain only</el-checkbox>
                    <el-tooltip
                      content="Evaluate selectors and rules without executing the matched action. The result explains what would happen."
                      placement="top"
                    >
                      <el-icon class="tip-icon"><InfoFilled /></el-icon>
                    </el-tooltip>
                  </div>
                  <div class="simulate-toggle">
                    <el-checkbox v-model="simulateSummary">Summary</el-checkbox>
                    <el-tooltip
                      content="Return a compact explanation. Turn it off when you need full selector, condition, and action details."
                      placement="top"
                    >
                      <el-icon class="tip-icon"><InfoFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </div>
                <small class="simulate-mode-label">{{ simulateModeLabel }}</small>
                <label class="field-label">
                  <span>
                    Depth
                    <el-tooltip
                      content="Maximum nested condition depth in the explanation tree. 0 keeps only top-level results."
                      placement="top"
                    >
                      <el-icon class="tip-icon"><InfoFilled /></el-icon>
                    </el-tooltip>
                  </span>
                  <el-input-number
                    v-model="simulateMaxDepth"
                    :min="0"
                    :max="10"
                    controls-position="right"
                    size="small"
                  />
                </label>
                <JsonEditor v-model="simulationJsonModel" :min-height="280" title="Event JSON" />
                <el-button
                  class="wide-action"
                  type="primary"
                  :icon="VideoPlay"
                  :loading="operationLoading.simulate"
                  :disabled="simulateDisabled"
                  :title="simulateDisabledReason"
                  @click="runSimulate"
                >
                  Simulate
                </el-button>
                <small v-if="simulateDisabledReason" class="simulate-hint">
                  {{ simulateDisabledReason }}
                </small>
                <section v-if="lastSimulationResultJson" class="simulate-result-panel">
                  <header>
                    <div>
                      <span>latest simulation result</span>
                      <strong>{{ simulateTarget }} / {{ simulateSummary ? 'summary' : 'full explanation' }}</strong>
                    </div>
                    <div>
                      <el-button size="small" :icon="VideoPlay" @click="runSimulate">Rerun</el-button>
                      <el-button size="small" @click="showWorkbenchMode('result')">Open raw</el-button>
                    </div>
                  </header>
                  <ResultInspector :raw-json="lastSimulationResultJson" :show-raw="false" />
                </section>
              </div>
            </section>

            <section v-show="activeWorkbenchMode === 'snapshots'" class="workbench-panel">
              <SnapshotRollbackWorkbench
                v-model:rollback-reason="rollbackReason"
                :rule-set-id="rulesetId"
                :snapshots="store.snapshots"
                :selected-snapshot-id="selectedSnapshotId"
                :preview-result="rollbackPreviewResult"
                :rollback-result="rollbackResult"
                :preview-error="rollbackPreviewError"
                :rollback-error="rollbackError"
                :previewing="operationLoading.rollbackPreview"
                :rolling-back="operationLoading.rollback"
                :has-simulation-event="hasSimulationEvent"
                @select="selectSnapshot"
                @preview="previewRollback"
                @rollback="confirmRollback"
              />
            </section>

            <section v-show="activeWorkbenchMode === 'result'" class="workbench-panel">
              <ResultInspector :raw-json="resultJson" />
            </section>
          </aside>
        </div>
      </template>
    </div>

    <RuleSetSettingsDialog
      v-model="settingsVisible"
      :rule-set="currentDraft"
      @saved="applyDraftUpdate"
    />
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Back,
  InfoFilled,
  Refresh,
  Setting,
  VideoPlay,
} from '@element-plus/icons-vue'
import PageContainer from '@/components/common/PageContainer.vue'
import JsonEditor from '@/components/common/JsonEditor.vue'
import TaskRail from '@/components/common/TaskRail.vue'
import PublishReadinessWorkbench from '@/components/rulesets/PublishReadinessWorkbench.vue'
import RuleEditorWorkbench from '@/components/rulesets/RuleEditorWorkbench.vue'
import RuleManager from '@/components/rulesets/RuleManager.vue'
import RuleSetSettingsDialog from '@/components/rulesets/RuleSetSettingsDialog.vue'
import RuleWorkbenchOverview from '@/components/rulesets/RuleWorkbenchOverview.vue'
import ResultInspector from '@/components/rulesets/ResultInspector.vue'
import SelectorValueGroup from '@/components/rulesets/SelectorValueGroup.vue'
import SnapshotRollbackWorkbench from '@/components/rulesets/SnapshotRollbackWorkbench.vue'
import { mockserverApi } from '@/api'
import { useMockserverStore } from '@/store'
import type {
  MockEvent,
  PublishRuleSetResponse,
  RollbackPreviewRequest,
  RollbackPreviewResponse,
  RollbackRuleSetResponse,
  Rule,
  RuleSet,
  SimulateRuleSetRequest,
} from '@/types'
import {
  buildRuleSeedFromDebugPayload,
  clearDebugToFixPayload,
  debugPayloadSummary,
  loadDebugToFixPayload,
  type DebugToFixPayload,
} from '@/utils/debugToFix'
import { mergeRuleDiagnostics, type RuleDiagnosticState } from '@/utils/ruleCollection'
import type { RuleEditorMode } from '@/utils/ruleFormAdapter'
import { buildValidationRuleDiagnostics } from '@/utils/publishReadiness'
import {
  buildDefaultSimulationEvent,
  buildSimulationEventWarnings,
  buildSimulationRuleDiagnostics,
} from '@/utils/simulationDiagnostics'
import { selectorValueSummaries } from '@/utils/entryLists'
import { ruleDisplayName } from '@/utils/ruleSummaries'
import {
  defaultWorkbenchModeForIntent,
  isWorkbenchMode,
  isWorkbenchIntent,
  workbenchIntentForMode,
  workbenchIntentTasks,
  workbenchTitle as buildWorkbenchTitle,
  type WorkbenchMode,
} from '@/utils/workbenchTasks'

const route = useRoute()
const router = useRouter()
const store = useMockserverStore()

const settingsVisible = ref(false)
const contextExpanded = ref(false)
const selectedSnapshotId = ref('')
const activeRuleId = ref('')
const simulateJson = ref('')
const lastSuggestedSimulationJson = ref('')
const simulateJsonDirty = ref(false)
const resultJson = ref('')
const lastSimulationResultJson = ref('')
const debugPayload = ref<DebugToFixPayload | null>(null)
const debugPayloadId = ref('')
const validationError = ref('')
const publishError = ref('')
const lastPublishResult = ref<PublishRuleSetResponse | null>(null)
const rollbackPreviewResult = ref<RollbackPreviewResponse | null>(null)
const rollbackResult = ref<RollbackRuleSetResponse | null>(null)
const rollbackPreviewError = ref('')
const rollbackError = ref('')
const activeWorkbenchMode = ref<WorkbenchMode>(initialWorkbenchMode())
const editorMode = ref<RuleEditorMode>('edit')
const simulateTarget = ref<'draft' | 'published'>('draft')
const simulateExplainOnly = ref(false)
const simulateSummary = ref(false)
const simulateMaxDepth = ref(3)
const rollbackReason = ref('')
const pageLoading = ref(false)
const pageError = ref('')
const operationError = ref('')
const operationLoading = reactive({
  refresh: false,
  validate: false,
  publish: false,
  simulate: false,
  rollbackPreview: false,
  rollback: false,
})

const workbenchModes = workbenchIntentTasks

const currentDraft = computed(() => store.currentDraft)
const rulesetId = computed(() => String(route.params.id || ''))
const currentPublished = computed(() => {
  return store.published.find((item) => item.ruleset.id === rulesetId.value) || null
})
const currentValidation = computed(() => store.validation)
const activeRule = computed(() => {
  return currentDraft.value?.rules.find((rule) => rule.id === activeRuleId.value) || null
})
const activeRuleDiagnostic = computed(() => {
  return activeRuleId.value ? ruleDiagnostics.value[activeRuleId.value] : undefined
})
const editorRule = computed(() => {
  return editorMode.value === 'edit' ? activeRule.value : null
})
const debugSeedRule = computed<Rule | null>(() => {
  if (!debugPayload.value || editorMode.value !== 'create' || !currentDraft.value) return null
  return buildRuleSeedFromDebugPayload(debugPayload.value, currentDraft.value.rules)
})
const debugSeedLabel = computed(() => {
  if (!debugPayload.value || editorMode.value !== 'create') return ''
  return `Seeded from runtime request #${debugPayload.value.request.recordId}`
})
const ruleDiagnostics = computed<Record<string, RuleDiagnosticState>>(() => {
  return mergeRuleDiagnostics(
    buildValidationRuleDiagnostics(currentValidation.value, currentDraft.value),
    buildSimulationRuleDiagnostics(store.simulation, currentDraft.value?.rules || [])
  )
})
const simulationJsonModel = computed({
  get: () => simulateJson.value,
  set: (value: string) => {
    simulateJson.value = value
    simulateJsonDirty.value = value !== lastSuggestedSimulationJson.value
  },
})
const isPublished = computed(() => Boolean(currentPublished.value))
const draftStatusLabel = computed(() => {
  if (!currentDraft.value) return pageLoading.value ? 'loading' : 'unloaded'
  return currentDraft.value.enabled ? 'enabled' : 'disabled'
})
const rulesCountLabel = computed(() => {
  if (!currentDraft.value) return '-'
  return `${currentDraft.value.rules.length} rules`
})
const isInitialLoading = computed(() => pageLoading.value && !currentDraft.value && !pageError.value)
const isNotFound = computed(() => isNotFoundError(pageError.value))
const visibleError = computed(() => {
  if (isNotFound.value) return ''
  return operationError.value || pageError.value
})
const draftVersionLabel = computed(() => {
  return currentDraft.value?.version ? `v${currentDraft.value.version}` : 'unversioned'
})
const publishedVersionLabel = computed(() => {
  if (!currentPublished.value) return 'no published snapshot'
  return `published v${currentPublished.value.ruleset.version || 0}`
})
const publishedStateLabel = computed(() => {
  if (!currentDraft.value) return '-'
  return currentPublished.value ? 'published' : 'draft only'
})
const publishedDetailLabel = computed(() => {
  if (!currentPublished.value) return 'not released yet'
  return currentPublished.value.snapshot_id
})
const snapshotStateLabel = computed(() => {
  if (!currentPublished.value) return '0'
  return String(store.snapshots.length)
})
const snapshotDetailLabel = computed(() => {
  if (!currentPublished.value) return 'publish first'
  return store.snapshots.length === 1 ? '1 snapshot' : `${store.snapshots.length} snapshots`
})
const workbenchTitle = computed(() => {
  return buildWorkbenchTitle(activeWorkbenchMode.value, activeRule.value, editorMode.value)
})
const activeWorkbenchIntent = computed(() => workbenchIntentForMode(activeWorkbenchMode.value))
const activeTaskContextLabel = computed(() => {
  if (debugPayload.value) return `debug request #${debugPayload.value.request.recordId}`
  if (activeWorkbenchMode.value === 'editor' && editorMode.value === 'create') return 'creating a new rule'
  if (activeRule.value) return `selected ${ruleDisplayName(activeRule.value)}`
  if (currentDraft.value?.rules.length) return 'select a rule to focus the task'
  return 'no rules in this ruleset yet'
})
const simulationEventStateLabel = computed(() => {
  return simulateJsonDirty.value ? 'custom event payload' : 'derived from current ruleset and selected rule'
})
const simulateModeLabel = computed(() => {
  return simulateSummary.value
    ? 'Summary mode: compact result for quick checks.'
    : 'Full explanation mode: selector checks, candidate rules, and condition traces are returned.'
})
const hasSimulationEvent = computed(() => Boolean(simulateJson.value.trim()))
const simulationWarnings = computed(() => {
  return buildSimulationEventWarnings(currentDraft.value, activeRule.value)
})
const selectorValues = computed(() => selectorValueSummaries(currentDraft.value?.selector))
const debugSourceSummary = computed(() => debugPayload.value ? debugPayloadSummary(debugPayload.value) : '')
const debugSourceDetail = computed(() => {
  if (!debugPayload.value) return ''
  const request = debugPayload.value.request
  const target = debugPayload.value.targetReason === 'runtime_match' ? 'runtime matched target' : 'same namespace/protocol target'
  return `${request.outcome} / ${request.status} / ${target}`
})
const simulateDisabledReason = computed(() => {
  if (simulateTarget.value === 'draft' && !currentDraft.value) {
    return 'Select a draft ruleset before simulation.'
  }
  if (simulateTarget.value === 'published' && !currentPublished.value) {
    return 'Publish this ruleset before simulating the published snapshot.'
  }
  if (!hasSimulationEvent.value) {
    return 'Event JSON is required.'
  }
  return ''
})
const simulateDisabled = computed(() => Boolean(simulateDisabledReason.value))

async function loadRuleSet() {
  if (!rulesetId.value) return
  pageLoading.value = true
  pageError.value = ''
  operationError.value = ''
  operationLoading.refresh = true
  if (currentDraft.value?.id !== rulesetId.value) {
    store.setCurrentDraft(null)
  }
  try {
    const [, draft] = await Promise.all([
      store.fetchPublished(),
      store.fetchDraft(rulesetId.value),
      store.fetchProtocols(),
    ])
    syncActiveRuleAfterLoad(draft)
    await refreshSnapshotsIfPublished()
    consumeDebugPayload()
  } catch (error) {
    pageError.value = toErrorMessage(error)
    if (currentDraft.value?.id !== rulesetId.value) {
      store.setCurrentDraft(null)
    }
  } finally {
    pageLoading.value = false
    operationLoading.refresh = false
  }
}

function applyDraftUpdate(ruleSet: RuleSet) {
  operationError.value = ''
  invalidateReadiness()
  store.setCurrentDraft(ruleSet)
  if (activeRuleId.value && ruleSet.rules.some((rule) => rule.id === activeRuleId.value)) return
  activeRuleId.value = firstRuleId(ruleSet)
}

function applyRuleEditorUpdate(ruleSet: RuleSet, ruleId: string) {
  applyDraftUpdate(ruleSet)
  activeRuleId.value = ruleId
  editorMode.value = 'edit'
}

function selectRule(ruleId: string) {
  activeRuleId.value = ruleId
  if (activeWorkbenchMode.value === 'result') {
    activeWorkbenchMode.value = 'inspect'
  }
  if (activeWorkbenchMode.value === 'editor' && editorMode.value === 'create') {
    editorMode.value = 'edit'
  }
}

function selectSnapshot(snapshotId: string) {
  if (selectedSnapshotId.value === snapshotId) return
  selectedSnapshotId.value = snapshotId
  clearRollbackWorkflow()
}

function openCreateRuleEditor() {
  editorMode.value = 'create'
  showWorkbenchMode('editor')
}

function openEditRuleEditor(rule: Rule) {
  activeRuleId.value = rule.id
  editorMode.value = 'edit'
  showWorkbenchMode('editor')
}

function refreshSimulationEventFromContext(force = false) {
  if (!currentDraft.value) return
  if (!force && debugPayload.value && simulateJsonDirty.value) return
  const nextJson = formatJSON(buildDefaultSimulationEvent(currentDraft.value, activeRule.value))
  const suggestedChanged = nextJson !== lastSuggestedSimulationJson.value
  lastSuggestedSimulationJson.value = nextJson
  if (suggestedChanged) {
    lastSimulationResultJson.value = ''
  }
  if (force || !simulateJsonDirty.value || !simulateJson.value) {
    simulateJson.value = nextJson
    simulateJsonDirty.value = false
  }
}

function consumeDebugPayload() {
  const id = String(route.query.debug || '')
  if (!id || id === debugPayloadId.value) return
  const payload = loadDebugToFixPayload(id)
  if (!payload || payload.targetRulesetId !== rulesetId.value) return
  debugPayload.value = payload
  debugPayloadId.value = id
  if (payload.action === 'create_rule') {
    applyDebugCreatePayload()
  } else {
    applyDebugSimulationPayload()
  }
}

function applyDebugSimulationPayload() {
  if (!debugPayload.value) return
  const eventJson = formatJSON(debugPayload.value.event)
  simulateTarget.value = 'draft'
  simulateExplainOnly.value = false
  simulateSummary.value = false
  simulateJson.value = eventJson
  lastSuggestedSimulationJson.value = eventJson
  simulateJsonDirty.value = true
  lastSimulationResultJson.value = ''
  showWorkbenchMode('simulate')
}

function applyDebugCreatePayload() {
  if (!debugPayload.value) return
  editorMode.value = 'create'
  showWorkbenchMode('editor')
}

function clearDebugPayload() {
  if (debugPayloadId.value) {
    clearDebugToFixPayload(debugPayloadId.value)
  }
  debugPayload.value = null
  debugPayloadId.value = ''
  const nextQuery = { ...route.query }
  delete nextQuery.debug
  router.replace({ query: nextQuery })
  refreshSimulationEventFromContext(true)
}

async function refreshSnapshotsIfPublished(preserveSelection = false) {
  const previousSnapshotId = selectedSnapshotId.value
  if (!preserveSelection) {
    selectedSnapshotId.value = ''
    clearRollbackWorkflow()
  }
  if (!isPublished.value) {
    store.snapshots = []
    return
  }
  await store.fetchSnapshots(rulesetId.value)
  if (!preserveSelection) return
  selectedSnapshotId.value = store.snapshots.some(
    (snapshot) => snapshot.snapshot_id === previousSnapshotId
  )
    ? previousSnapshotId
    : ''
}

async function validateDraft() {
  if (!currentDraft.value) return
  operationLoading.validate = true
  operationError.value = ''
  validationError.value = ''
  publishError.value = ''
  activeWorkbenchMode.value = 'readiness'
  try {
    const response = await store.validateDraft(currentDraft.value.id)
    resultJson.value = formatJSON(response)
  } catch (error) {
    validationError.value = toErrorMessage(error)
  } finally {
    operationLoading.validate = false
  }
}

async function openReadinessAndValidate() {
  activeWorkbenchMode.value = 'readiness'
  await validateDraft()
}

async function requestPublishDraft() {
  activeWorkbenchMode.value = 'readiness'
  publishError.value = ''
  if (!currentValidation.value?.result.valid) {
    publishError.value = 'Run validation and fix all issues before publishing.'
    return
  }
  await publishDraft()
}

async function publishDraft() {
  if (!currentDraft.value) return
  operationError.value = ''
  publishError.value = ''
  try {
    const { value } = await ElMessageBox.prompt('Publish reason', 'Publish ruleset', {
      inputPlaceholder: 'release order status mock',
      confirmButtonText: 'Publish',
      cancelButtonText: 'Cancel',
    })
    operationLoading.publish = true
    const response = await store.publishDraft(currentDraft.value.id, value)
    lastPublishResult.value = response
    store.setCurrentDraft(response.ruleset)
    resultJson.value = formatJSON(response)
    activeWorkbenchMode.value = 'readiness'
    await refreshSnapshotsIfPublished()
  } catch (error) {
    if (isDialogCancel(error)) return
    publishError.value = toErrorMessage(error)
  } finally {
    operationLoading.publish = false
  }
}

async function runSimulate() {
  operationLoading.simulate = true
  operationError.value = ''
  try {
    const event = parseJSON<MockEvent>(simulateJson.value, 'Event JSON')
    const request: SimulateRuleSetRequest = {
      event,
      explain_only: simulateExplainOnly.value,
      explain_max_depth: simulateMaxDepth.value,
      explain_summary: simulateSummary.value,
    }
    const result =
      simulateTarget.value === 'draft'
        ? await store.simulateDraft(rulesetId.value, request)
        : await store.simulatePublished(request)
    const formattedResult = formatJSON(result)
    resultJson.value = formattedResult
    lastSimulationResultJson.value = formattedResult
  } catch (error) {
    operationError.value = toErrorMessage(error)
  } finally {
    operationLoading.simulate = false
  }
}

async function previewRollback() {
  if (!selectedSnapshotId.value) return
  operationLoading.rollbackPreview = true
  operationError.value = ''
  rollbackPreviewError.value = ''
  rollbackError.value = ''
  rollbackPreviewResult.value = null
  try {
    const event = safeParseEvent()
    const request: RollbackPreviewRequest = {
      snapshot_id: selectedSnapshotId.value,
      explain_only: true,
      explain_max_depth: simulateMaxDepth.value,
      explain_summary: true,
    }
    if (event) {
      request.event = event
    }
    const response = await mockserverApi.rollbackPreview(rulesetId.value, request)
    rollbackPreviewResult.value = response
    rollbackResult.value = null
    resultJson.value = formatJSON(response)
    activeWorkbenchMode.value = 'snapshots'
  } catch (error) {
    rollbackPreviewError.value = toErrorMessage(error)
  } finally {
    operationLoading.rollbackPreview = false
  }
}

async function confirmRollback() {
  if (!selectedSnapshotId.value) return
  operationError.value = ''
  rollbackError.value = ''
  rollbackPreviewError.value = ''
  try {
    await ElMessageBox.confirm(`Rollback ${rulesetId.value} to ${selectedSnapshotId.value}?`, 'Run rollback', {
      type: 'warning',
      confirmButtonText: 'Rollback',
      cancelButtonText: 'Cancel',
    })
    operationLoading.rollback = true
    const response = await store.rollback(rulesetId.value, selectedSnapshotId.value, rollbackReason.value)
    selectedSnapshotId.value = response.snapshot.snapshot_id
    rollbackResult.value = response
    rollbackPreviewResult.value = null
    resultJson.value = formatJSON(response)
    invalidateReadiness()
    activeWorkbenchMode.value = 'snapshots'
    await refreshSnapshotsIfPublished(true)
  } catch (error) {
    if (isDialogCancel(error)) return
    rollbackError.value = toErrorMessage(error)
  } finally {
    operationLoading.rollback = false
  }
}

function safeParseEvent() {
  if (!simulateJson.value.trim()) return undefined
  try {
    return parseJSON<MockEvent>(simulateJson.value, 'Event JSON')
  } catch {
    return undefined
  }
}

function parseJSON<T>(value: string, label: string): T {
  try {
    return JSON.parse(value) as T
  } catch (error) {
    const message = `${label} is not valid JSON: ${toErrorMessage(error)}`
    ElMessage.error(message)
    throw new Error(message)
  }
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}

function isNotFoundError(message: string) {
  const normalized = message.toLowerCase()
  return normalized.includes('not found') || normalized.includes('does not exist')
}

function isDialogCancel(error: unknown) {
  return error === 'cancel' || error === 'close' || toErrorMessage(error).toLowerCase() === 'cancel'
}

function showWorkbenchMode(mode: WorkbenchMode) {
  if (mode === 'editor' && editorMode.value === 'edit' && !activeRule.value) {
    editorMode.value = currentDraft.value?.rules.length ? 'edit' : 'create'
  }
  activeWorkbenchMode.value = mode
  operationError.value = ''
}

function updateWorkbenchIntent(intent: string) {
  if (isWorkbenchIntent(intent)) {
    showWorkbenchMode(defaultWorkbenchModeForIntent(intent))
  }
}

function invalidateReadiness() {
  store.validation = null
  validationError.value = ''
  publishError.value = ''
  lastPublishResult.value = null
}

function clearRollbackWorkflow() {
  rollbackPreviewResult.value = null
  rollbackResult.value = null
  rollbackPreviewError.value = ''
  rollbackError.value = ''
}

function firstRuleId(ruleSet: RuleSet) {
  return [...ruleSet.rules].sort((left, right) => left.priority - right.priority)[0]?.id || ''
}

function syncActiveRuleAfterLoad(ruleSet: RuleSet) {
  const deepLinkedRule = String(route.query.rule || '')
  if (deepLinkedRule && ruleSet.rules.some((rule) => rule.id === deepLinkedRule)) {
    activeRuleId.value = deepLinkedRule
    return
  }
  if (activeRuleId.value && ruleSet.rules.some((rule) => rule.id === activeRuleId.value)) return
  activeRuleId.value = firstRuleId(ruleSet)
}

function initialWorkbenchMode(): WorkbenchMode {
  const mode = String(route.query.workbench || '')
  if (isWorkbenchMode(mode)) {
    return mode
  }
  return 'inspect'
}

watch(rulesetId, loadRuleSet)
watch(
  () => route.query.rule,
  () => {
    if (currentDraft.value) syncActiveRuleAfterLoad(currentDraft.value)
  }
)
watch(
  () => route.query.debug,
  () => {
    if (currentDraft.value) consumeDebugPayload()
  }
)
watch([currentDraft, activeRule], () => refreshSimulationEventFromContext(false), { immediate: true })
onMounted(loadRuleSet)
</script>

<style lang="scss" scoped>
.workspace-page {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.workspace-alert {
  flex-shrink: 0;
}

.workspace-state {
  min-height: 360px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--ms-space-6);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-xl);
  background: var(--ms-panel-bg);
  box-shadow: var(--ms-shadow-xs), var(--ms-shadow-inset);
}

.workspace-context {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: minmax(230px, 0.7fr) minmax(320px, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 6px var(--ms-space-2);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-xl);
  background: var(--ms-panel-bg);
  box-shadow: var(--ms-shadow-xs), var(--ms-shadow-inset);
}

.context-identity {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  span:not(.status-dot) {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong {
    overflow: hidden;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.context-actions {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--ms-space-1);

  :deep(.el-button) {
    min-width: 0;
    height: 30px;
    margin-left: 0;
  }
}

.context-summary {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-2);

  .context-cell.is-selector {
    grid-column: auto;
  }
}

.context-grid {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(5, minmax(130px, 1fr));
  gap: var(--ms-space-2);
}

.context-cell {
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  min-height: 46px;
  padding: 6px 8px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  &.is-compact {
    min-height: 38px;
  }

  &.is-wide,
  &.is-selector {
    grid-column: span 2;
  }

  > span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    font-family: var(--ms-font-mono);
  }

  small {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.debug-source-banner {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--ms-radius-xl);
  background: var(--ms-amber-50);
  box-shadow: var(--ms-shadow-xs);

  > div:first-child {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  span {
    color: var(--ms-amber-600);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
  }

  small {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
  }
}

.debug-source-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--ms-space-2);
}

.rules-workspace {
  flex: 1;
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(420px, 0.42fr) minmax(640px, 0.58fr);
  gap: var(--ms-space-3);
}

.rules-main,
.tool-dock {
  min-height: 0;
  overflow: hidden;
}

.rules-main {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.tool-dock {
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-xl);
  background: var(--ms-panel-bg);
  box-shadow: var(--ms-shadow-xs), var(--ms-shadow-inset);
}

.workbench-header {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: minmax(150px, 0.28fr) minmax(0, 0.72fr);
  align-items: center;
  gap: var(--ms-space-3);
  padding: var(--ms-space-2) var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);

  > div:first-child {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong {
    overflow: hidden;
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    font-weight: var(--ms-font-bold);
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.intent-subnav {
  flex-shrink: 0;
  display: flex;
  gap: var(--ms-space-1);
  padding: var(--ms-space-2) var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);
  background: var(--ms-panel-bg-soft);

  button {
    height: 28px;
    padding: 0 10px;
    border: 1px solid transparent;
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: transparent;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    cursor: pointer;
    transition:
      border-color var(--ms-transition-fast),
      background var(--ms-transition-fast),
      color var(--ms-transition-fast);

    &:hover,
    &.active {
      border-color: rgba(37, 99, 235, 0.26);
      color: var(--ms-teal-700);
      background: var(--ms-control-bg);
    }
  }
}

.workbench-panel {
  min-height: 0;
  flex: 1;
  height: 100%;
  overflow: auto;
  padding: var(--ms-space-3);
}

.tool-stack {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.simulation-context {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid rgba(37, 99, 235, 0.22);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-teal-50);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.simulation-warning-list {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid rgba(245, 158, 11, 0.28);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-amber-50);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-2);
  }

  span,
  strong {
    color: var(--ms-amber-600);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  small {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
  }
}

.simulate-toolbar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) auto auto;
  gap: var(--ms-space-2);
  align-items: center;
}

.simulate-control,
.simulate-toggle {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 8px 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
}

.simulate-control {
  justify-content: space-between;
}

.simulate-control > span,
.field-label > span {
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.simulate-toggle {
  justify-content: center;
}

.tip-icon {
  color: var(--ms-text-tertiary);
  font-size: 14px;
  cursor: help;
}

.field-label {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 8px 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  color: var(--ms-text-tertiary);
  background: var(--ms-control-bg);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.wide-action {
  width: 100%;
}

.simulate-hint {
  display: block;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
}

.simulate-mode-label {
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
}

.simulate-result-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);

  > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-3);

    > div {
      min-width: 0;
    }

    > div:first-child {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    > div:last-child {
      display: flex;
      align-items: center;
      gap: var(--ms-space-2);
    }
  }

  span {
    color: var(--ms-teal-700);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong {
    overflow: hidden;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.context-details-enter-active,
.context-details-leave-active {
  transition:
    opacity var(--ms-transition-fast),
    transform var(--ms-transition-fast);
}

.context-details-enter-from,
.context-details-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

@media (max-width: 1320px) {
  .context-actions {
    justify-content: flex-start;
  }

  .context-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .rules-workspace {
    height: auto;
    grid-template-columns: 1fr;
  }

  .rules-main {
    order: 1;
    height: min(64vh, 620px);
    min-height: 560px;
    overflow: hidden;
  }

  .tool-dock {
    order: 2;
    min-height: 680px;
  }
}

@media (max-width: 1180px) {
  .workspace-context {
    grid-template-columns: 1fr;
    gap: var(--ms-space-2);
  }

  .debug-source-banner {
    align-items: stretch;
    flex-direction: column;
  }

  .debug-source-actions {
    justify-content: flex-start;
  }

  .context-identity {
    order: 1;
  }

  .context-summary {
    order: 2;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .context-actions {
    order: 3;
  }

  .context-grid,
  .context-cell.is-wide,
  .context-cell.is-selector {
    order: 4;
    grid-template-columns: 1fr;
    grid-column: auto;
  }

  .context-actions {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));

    :deep(.el-button) {
      width: 100%;
      min-width: 0;
      margin-left: 0;
    }

    :deep(.el-dropdown) {
      width: 100%;
    }

    :deep(.el-dropdown .el-button) {
      width: 100%;
    }
  }
}

@media (max-width: 840px) {
  .workbench-header {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .rules-main {
    height: min(52vh, 460px);
    min-height: 360px;
  }
}

@media (max-width: 560px) {
  .workspace-page {
    gap: var(--ms-space-3);
  }

  .workspace-context,
  .workbench-panel {
    padding: var(--ms-space-2);
  }

  .context-actions {
    grid-template-columns: 1fr;

    :deep(.el-dropdown) {
      width: 100%;
    }

    :deep(.el-dropdown .el-button) {
      width: 100%;
    }
  }

  .context-summary {
    grid-template-columns: 1fr;
  }

  .simulate-toolbar {
    grid-template-columns: 1fr;
  }

  .simulate-control,
  .simulate-toggle {
    justify-content: space-between;
  }

  .simulate-result-panel > header {
    align-items: stretch;
    flex-direction: column;
  }

  .rules-main {
    min-height: 340px;
  }

  .tool-dock {
    min-height: 520px;
  }

}
</style>
