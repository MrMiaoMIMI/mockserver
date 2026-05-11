<template>
  <PageContainer title="Runtime Diagnostics" eyebrow="telemetry">
    <template #meta>
      <span class="meta-pill">
        <span class="status-dot is-on" />
        match {{ matchRate }}%
      </span>
      <span class="meta-pill">{{ recentRows.length }} recent</span>
      <span class="meta-pill">{{ fallbackTotal }} fallback</span>
    </template>
    <template #actions>
      <el-button :type="insightsExpanded ? 'primary' : 'default'" :icon="View" @click="insightsExpanded = !insightsExpanded">
        {{ insightsExpanded ? 'Hide insights' : 'Insights' }}
      </el-button>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">Refresh</el-button>
    </template>

    <div class="diagnostics-dashboard">
      <section class="diagnostic-grid" :class="{ 'insights-open': insightsExpanded }">
        <section class="request-stream panel-surface">
          <header class="panel-heading">
            <div>
              <span>recent requests</span>
              <strong>{{ recentRows.length }}</strong>
            </div>
            <small>live runtime requests only</small>
          </header>

          <div class="request-toolbar">
            <el-input
              v-model="filters.query"
              clearable
              :prefix-icon="Search"
              placeholder="Search path, namespace, trace, ruleset, rule"
            />
            <el-select v-model="filters.namespace" clearable filterable placeholder="Namespace">
              <el-option
                v-for="namespace in namespaceOptions"
                :key="namespace"
                :label="namespace"
                :value="namespace"
              />
            </el-select>
            <el-select v-model="filters.sort" placeholder="Sort">
              <el-option v-for="option in sortOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
            <FilterSegment
              v-model="filters.outcome"
              :options="outcomeOptions"
              aria-label="Runtime outcome filter"
            />
          </div>

          <el-empty v-if="!recentRows.length && !loading" description="No runtime request diagnostics yet" />
          <div v-else v-loading="loading" class="request-list">
            <article
              v-for="row in recentRows"
              :key="row.record.id"
              :class="['request-row', `is-${row.tone}`, { active: selectedRequest?.record.id === row.record.id }]"
              tabindex="0"
              @click="openRequestDetail(row)"
              @keydown.enter.prevent="openRequestDetail(row)"
              @keydown.space.prevent="openRequestDetail(row)"
            >
              <div class="request-main">
                <div class="request-line">
                  <span class="method-chip">{{ row.record.method || '-' }}</span>
                  <code :title="row.pathLabel">{{ row.pathLabel }}</code>
                </div>
                <div class="request-meta">
                  <span>{{ row.observedAtLabel }}</span>
                  <span>{{ row.record.namespace || 'default' }}</span>
                  <span v-if="row.record.host">{{ row.record.host }}</span>
                  <button
                    v-if="row.record.trace_id"
                    type="button"
                    class="trace-button"
                    :title="row.record.trace_id"
                    @click.stop="copyText(row.record.trace_id)"
                  >
                    trace {{ shortText(row.record.trace_id, 10) }}
                  </button>
                </div>
              </div>

              <div class="request-state">
                <span :class="['outcome-chip', `is-${row.tone}`]">{{ row.outcomeLabel }}</span>
                <strong :class="`is-${statusTone(row.record.status)}`">{{ row.statusLabel }}</strong>
                <small>{{ row.durationLabel }}</small>
              </div>

              <div class="request-links">
                <div>
                  <span>ruleset</span>
                  <code :title="row.record.ruleset_id || ''">{{ row.record.ruleset_id || '-' }}</code>
                </div>
                <div>
                  <span>rule</span>
                  <code :title="row.record.rule_id || ''">{{ row.record.rule_id || '-' }}</code>
                </div>
                <div>
                  <span>fallback</span>
                  <code>{{ row.record.fallback_reason || '-' }}</code>
                </div>
                <div v-if="row.record.message">
                  <span>message</span>
                  <code :title="row.record.message">{{ row.record.message }}</code>
                </div>
              </div>
            </article>
          </div>
        </section>

        <aside v-if="insightsExpanded" class="diagnostic-side">
          <section class="side-panel panel-surface">
            <header class="panel-heading compact">
              <span>traffic summary</span>
              <strong>{{ formatNumber(metrics?.total_requests ?? 0) }}</strong>
            </header>
            <div class="insight-metrics">
              <MetricCard
                v-for="card in metricCards"
                :key="card.label"
                :label="card.label"
                :value="card.value"
                :caption="card.caption"
                :tone="card.tone"
              />
            </div>
          </section>

          <section class="side-panel panel-surface">
            <header class="panel-heading compact">
              <span>status classes</span>
              <strong>{{ statusEntries.length }}</strong>
            </header>
            <div v-if="statusEntries.length" class="compact-list">
              <div v-for="entry in statusEntries" :key="entry.key" class="compact-row">
                <code>{{ entry.label }}</code>
                <div class="mini-bar"><span :style="{ width: hitWidth(entry.value, statusEntries) }" /></div>
                <strong>{{ entry.value }}</strong>
              </div>
            </div>
            <el-empty v-else description="No status data yet" />
          </section>

          <section class="side-panel panel-surface">
            <header class="panel-heading compact">
              <span>fallback reasons</span>
              <strong>{{ fallbackEntries.length }}</strong>
            </header>
            <div v-if="fallbackEntries.length" class="compact-list">
              <div v-for="entry in fallbackEntries" :key="entry.key" class="compact-row">
                <code>{{ entry.label }}</code>
                <div class="mini-bar is-warn"><span :style="{ width: hitWidth(entry.value, fallbackEntries) }" /></div>
                <strong>{{ entry.value }}</strong>
              </div>
            </div>
            <el-empty v-else description="No fallback data yet" />
          </section>

          <section class="side-panel panel-surface">
            <header class="panel-heading compact">
              <span>ruleset hits</span>
              <strong>{{ rulesetHits.length }}</strong>
            </header>
            <div v-if="rulesetHits.length" class="compact-list">
              <button
                v-for="entry in rulesetHits"
                :key="entry.key"
                class="compact-row as-button"
                type="button"
                @click="openRulesetById(entry.key)"
              >
                <code :title="entry.key">{{ entry.key }}</code>
                <div class="mini-bar"><span :style="{ width: hitWidth(entry.value, rulesetHits) }" /></div>
                <strong>{{ entry.value }}</strong>
              </button>
            </div>
            <el-empty v-else description="No ruleset hits yet" />
          </section>

          <section class="side-panel panel-surface">
            <header class="panel-heading compact">
              <span>rule hits</span>
              <strong>{{ ruleHits.length }}</strong>
            </header>
            <div v-if="ruleHits.length" class="compact-list">
              <div v-for="entry in ruleHits" :key="entry.key" class="compact-row">
                <code :title="entry.key">{{ entry.key }}</code>
                <div class="mini-bar is-accent"><span :style="{ width: hitWidth(entry.value, ruleHits) }" /></div>
                <strong>{{ entry.value }}</strong>
              </div>
            </div>
            <el-empty v-else description="No rule hits yet" />
          </section>

          <section class="replay-panel panel-surface">
            <header class="panel-heading compact">
              <span>replay result</span>
              <strong>{{ replaySourceLabel }}</strong>
            </header>
            <ResultInspector :raw-json="replayJson" />
          </section>
        </aside>
      </section>
    </div>

    <el-drawer
      v-model="requestDetailVisible"
      :title="selectedRequest ? `Request #${selectedRequest.record.id}` : 'Request detail'"
      size="min(560px, calc(100vw - 24px))"
      append-to-body
    >
      <div v-if="selectedRequest" class="request-drawer">
        <header class="request-drawer-heading">
          <div>
            <StateChip :label="selectedRequest.record.method || '-'" tone="primary" />
            <strong :title="selectedRequest.pathLabel">{{ selectedRequest.pathLabel }}</strong>
            <small>{{ selectedRequest.observedAtLabel }} / {{ selectedRequest.durationLabel }}</small>
          </div>
          <StateChip :label="selectedRequest.outcomeLabel" :tone="selectedRequest.tone" />
        </header>

        <KeyValueGrid :items="requestFacts(selectedRequest)" />

        <section class="drawer-section">
          <header>
            <span>Route context</span>
            <strong>{{ selectedRequest.record.namespace || 'default' }}</strong>
          </header>
          <KeyValueGrid :items="routeFacts(selectedRequest)" />
        </section>

        <section v-if="selectedDiagnosis" class="drawer-section diagnosis-section">
          <header>
            <span>Diagnosis</span>
            <strong>{{ selectedDiagnosis.title }}</strong>
          </header>
          <p>{{ selectedDiagnosis.detail }}</p>
          <small>{{ selectedDiagnosis.nextAction }}</small>
          <code>
            target: {{ debugTargetLabel(selectedRequest) }}
          </code>
          <p v-if="selectedRequest.record.message">{{ selectedRequest.record.message }}</p>
          <code v-if="selectedRequest.record.fallback_reason">
            fallback: {{ selectedRequest.record.fallback_reason }}
          </code>
        </section>

        <section v-if="selectedRequest.record.event" class="drawer-section">
          <header>
            <span>Replay event</span>
            <strong>captured</strong>
          </header>
          <ResultInspector :raw-json="requestEventJson(selectedRequest)" :show-raw="false" />
        </section>

        <footer class="drawer-actions">
          <el-button
            type="primary"
            :disabled="!selectedRequest.canReplay || !selectedDebugTarget"
            :icon="VideoPlay"
            @click="sendToWorkspace(selectedRequest, 'simulate')"
          >
            Use as simulation
          </el-button>
          <el-button
            :disabled="!selectedRequest.canReplay || !selectedDebugTarget"
            :icon="EditPen"
            @click="sendToWorkspace(selectedRequest, 'create_rule')"
          >
            Create rule
          </el-button>
          <el-button
            :disabled="!selectedRequest.canOpenRuleset"
            :icon="Right"
            @click="openRuleset(selectedRequest)"
          >
            Open rules
          </el-button>
          <el-button
            :disabled="!selectedRequest.canReplay"
            :loading="replayLoadingId === selectedRequest.record.id"
            @click="replay(selectedRequest)"
          >
            Replay
          </el-button>
          <el-button
            v-if="selectedRequest.record.trace_id"
            @click="copyText(selectedRequest.record.trace_id)"
          >
            Copy trace
          </el-button>
        </footer>
      </div>
    </el-drawer>
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { EditPen, Refresh, Right, Search, VideoPlay, View } from '@element-plus/icons-vue'
import FilterSegment, { type FilterSegmentOption } from '@/components/common/FilterSegment.vue'
import KeyValueGrid from '@/components/common/KeyValueGrid.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageContainer from '@/components/common/PageContainer.vue'
import StateChip from '@/components/common/StateChip.vue'
import ResultInspector from '@/components/rulesets/ResultInspector.vue'
import { useMockserverStore } from '@/store'
import type { RuntimeDiagnosticFilters, RuntimeOutcomeFilter, RuntimeSortKey } from '@/utils/runtimeDiagnostics'
import {
  buildDebugToFixRoute,
  buildRuntimeDiagnosis,
  createDebugToFixPayload,
  inferDebugToFixTarget,
  storeDebugToFixPayload,
  type DebugToFixAction,
} from '@/utils/debugToFix'
import {
  buildRuntimeDiagnosticRows,
  buildRuntimeMetricCards,
  defaultRuntimeDiagnosticFilters,
  runtimeMatchRate,
  runtimeNamespaceOptions,
  sumMetricValues,
  toSortedMetricEntries,
  type RuntimeDiagnosticRow,
  type RuntimeMetricEntry,
  type RuntimeTone,
} from '@/utils/runtimeDiagnostics'

const router = useRouter()
const store = useMockserverStore()
const loading = ref(false)
const insightsExpanded = ref(false)
const replayLoadingId = ref<number | null>(null)
const replaySource = ref<RuntimeDiagnosticRow | null>(null)
const selectedRequestId = ref<number | null>(null)
const requestDetailVisible = ref(false)
const replayJson = ref('')

const outcomeOptions: Array<FilterSegmentOption & { value: RuntimeOutcomeFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Matched', value: 'matched' },
  { label: 'Fallback', value: 'fallback' },
  { label: 'Unmatched', value: 'unmatched' },
  { label: 'Error', value: 'error' },
]
const sortOptions: Array<{ label: string; value: RuntimeSortKey }> = [
  { label: 'Newest', value: 'newest' },
  { label: 'Slowest', value: 'slowest' },
  { label: 'Status', value: 'status' },
  { label: 'Outcome', value: 'outcome' },
]

const filters = reactive<RuntimeDiagnosticFilters>(defaultRuntimeDiagnosticFilters())
const metrics = computed(() => store.metrics)
const matchRate = computed(() => runtimeMatchRate(metrics.value))
const fallbackTotal = computed(() => sumMetricValues(metrics.value?.fallback_reasons))
const metricCards = computed(() => buildRuntimeMetricCards(metrics.value))
const recentRows = computed(() => buildRuntimeDiagnosticRows(metrics.value, filters))
const selectedRequest = computed(() => {
  return recentRows.value.find((row) => row.record.id === selectedRequestId.value) || null
})
const selectedDiagnosis = computed(() => selectedRequest.value ? buildRuntimeDiagnosis(selectedRequest.value) : null)
const selectedDebugTarget = computed(() => {
  return selectedRequest.value ? inferDebugToFixTarget(selectedRequest.value, store.drafts) : null
})
const namespaceOptions = computed(() => runtimeNamespaceOptions(metrics.value))
const statusEntries = computed(() => toSortedMetricEntries(metrics.value?.status_codes))
const fallbackEntries = computed(() => toSortedMetricEntries(metrics.value?.fallback_reasons))
const rulesetHits = computed(() => toSortedMetricEntries(metrics.value?.ruleset_matches).slice(0, 8))
const ruleHits = computed(() => toSortedMetricEntries(metrics.value?.rule_matches).slice(0, 8))
const replaySourceLabel = computed(() => {
  if (!replaySource.value) return 'idle'
  return `#${replaySource.value.record.id}`
})

async function loadData() {
  loading.value = true
  try {
    await Promise.all([store.fetchMetrics(), store.fetchDrafts()])
  } finally {
    loading.value = false
  }
}

async function replay(row: RuntimeDiagnosticRow) {
  if (!row.record.event) {
    ElMessage.warning('This request has no replayable event')
    return
  }
  replayLoadingId.value = row.record.id
  replaySource.value = row
  selectedRequestId.value = row.record.id
  try {
    const result = await store.simulatePublished({
      event: row.record.event,
      explain_only: false,
      explain_summary: true,
      explain_max_depth: 3,
    })
    replayJson.value = JSON.stringify({ result }, null, 2)
  } catch (error) {
    replayJson.value = ''
    ElMessage.error(error instanceof Error ? error.message : 'Replay failed')
  } finally {
    replayLoadingId.value = null
  }
}

function openRequestDetail(row: RuntimeDiagnosticRow) {
  selectedRequestId.value = row.record.id
  requestDetailVisible.value = true
}

function openRuleset(row: RuntimeDiagnosticRow) {
  if (!row.record.ruleset_id) return
  router.push({
    path: `/rulesets/${encodeURIComponent(row.record.ruleset_id)}/rules`,
    query: {
      workbench: 'inspect',
      ...(row.record.rule_id ? { rule: row.record.rule_id } : {}),
    },
  })
}

function sendToWorkspace(row: RuntimeDiagnosticRow, action: DebugToFixAction) {
  const target = inferDebugToFixTarget(row, store.drafts)
  if (!target) {
    ElMessage.warning('No target ruleset found for this runtime request')
    return
  }
  const payload = createDebugToFixPayload(row, action, target)
  if (!payload) {
    ElMessage.warning('This request has no captured event')
    return
  }
  storeDebugToFixPayload(payload)
  router.push(buildDebugToFixRoute(payload))
}

function requestFacts(row: RuntimeDiagnosticRow) {
  return [
    { label: 'Outcome', value: row.outcomeLabel, tone: row.tone },
    { label: 'Status', value: row.statusLabel, tone: statusTone(row.record.status) },
    { label: 'Duration', value: row.durationLabel },
    { label: 'Observed', value: row.observedAtLabel },
    { label: 'Trace', value: row.record.trace_id || '-', code: true },
    { label: 'Replay', value: row.canReplay ? 'available' : 'none', tone: row.canReplay ? 'ok' : 'neutral' },
  ]
}

function routeFacts(row: RuntimeDiagnosticRow) {
  return [
    { label: 'Namespace', value: row.record.namespace || 'default', code: true },
    { label: 'Host', value: row.record.host || '-', code: true },
    { label: 'Path', value: row.pathLabel, code: true },
    { label: 'Ruleset', value: row.record.ruleset_id || '-', code: true },
    { label: 'Rule', value: row.record.rule_id || '-', code: true },
    { label: 'Fallback', value: row.record.fallback_reason || '-' },
  ]
}

function debugTargetLabel(row: RuntimeDiagnosticRow) {
  return inferDebugToFixTarget(row, store.drafts)?.label || 'No draft target found'
}

function requestEventJson(row: RuntimeDiagnosticRow) {
  return JSON.stringify({ event: row.record.event }, null, 2)
}

function openRulesetById(rulesetId: string) {
  router.push(`/rulesets/${encodeURIComponent(rulesetId)}/rules`)
}

async function copyText(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    ElMessage.success('Trace ID copied')
  } catch {
    ElMessage.error('Copy failed')
  }
}

function hitWidth(value: number, rows: RuntimeMetricEntry[]) {
  const max = Math.max(...rows.map((row) => row.value), 1)
  return `${Math.max((value / max) * 100, 6)}%`
}

function statusTone(status?: number): RuntimeTone {
  if (!status) return 'neutral'
  if (status >= 500) return 'danger'
  if (status >= 400) return 'warn'
  if (status >= 200 && status < 400) return 'ok'
  return 'neutral'
}

function shortText(value: string, size: number) {
  if (value.length <= size) return value
  return `${value.slice(0, size)}...`
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US').format(value)
}

onMounted(loadData)
</script>

<style lang="scss" scoped>
.diagnostics-dashboard {
  min-height: 0;
  height: 100%;
  display: grid;
  grid-template-rows: minmax(0, 1fr);
  gap: var(--ms-space-3);
}

.insight-metrics {
  min-width: 0;
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
}

.diagnostic-grid {
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--ms-space-3);

  &.insights-open {
    grid-template-columns: minmax(0, 1fr) minmax(330px, 0.34fr);
  }
}

.request-stream,
.diagnostic-side {
  min-height: 0;
}

.request-stream {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-heading {
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: 0 var(--ms-space-4);
  border-bottom: 1px solid var(--ms-border-light);

  div {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: var(--ms-space-2);
  }

  span {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong,
  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }

  &.compact {
    min-height: 42px;
  }
}

.request-toolbar {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) minmax(150px, 190px) minmax(130px, 160px);
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);
}

.request-toolbar :deep(.filter-segment) {
  grid-column: 1 / -1;
  width: max-content;
  max-width: 100%;
}

.request-list {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  overflow: auto;
  padding: var(--ms-space-3);
}

.request-row {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) 126px minmax(240px, 0.9fr);
  align-items: center;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-left-width: 3px;
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
  cursor: pointer;
  transition:
    background var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast),
    transform var(--ms-transition-fast);

  &:hover,
  &:focus-visible,
  &.active {
    background: var(--ms-panel-bg-soft);
    box-shadow:
      var(--ms-shadow-xs),
      0 0 0 1px rgba(15, 118, 110, 0.16);
    outline: none;
  }

  &.active {
    transform: translateY(-1px);
  }

  &.is-ok {
    border-left-color: var(--ms-green-500);
  }

  &.is-warn {
    border-left-color: var(--ms-amber-500);
  }

  &.is-danger {
    border-left-color: var(--ms-red-500);
  }
}

.request-main,
.request-links {
  min-width: 0;
}

.request-line {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);

  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.method-chip,
.outcome-chip {
  flex: 0 0 auto;
  padding: 4px 9px;
  border-radius: var(--ms-radius-pill);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  line-height: 1.2;
}

.method-chip {
  color: var(--ms-blue-500);
  background: var(--ms-blue-50);
}

.request-meta {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--ms-space-1) var(--ms-space-2);
  margin-top: var(--ms-space-2);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
}

.trace-button {
  max-width: 150px;
  padding: 0;
  overflow: hidden;
  border: none;
  color: var(--ms-teal-700);
  background: transparent;
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.request-state {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--ms-space-1);

  strong {
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    line-height: 1.1;

    &.is-ok {
      color: var(--ms-green-600);
    }

    &.is-warn {
      color: var(--ms-amber-600);
    }

    &.is-danger {
      color: var(--ms-red-600);
    }
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.outcome-chip {
  color: var(--ms-text-tertiary);
  background: var(--ms-panel-bg);

  &.is-ok {
    color: var(--ms-green-600);
    background: var(--ms-green-50);
  }

  &.is-warn {
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
  }

  &.is-danger {
    color: var(--ms-red-600);
    background: var(--ms-red-50);
  }
}

.request-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--ms-space-2);

  div {
    min-width: 0;
  }

  span {
    display: block;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    display: block;
    min-width: 0;
    margin-top: 2px;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.diagnostic-side {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  overflow: auto;
}

.side-panel,
.replay-panel {
  min-height: 0;
  overflow: hidden;
}

.compact-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
}

.compact-row {
  min-width: 0;
  width: 100%;
  display: grid;
  grid-template-columns: minmax(88px, 0.9fr) minmax(90px, 1fr) 42px;
  align-items: center;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border: none;
  border-radius: var(--ms-radius-md);
  color: inherit;
  background: var(--ms-control-bg);
  text-align: left;

  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    justify-self: end;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
  }

  &.as-button {
    cursor: pointer;
  }
}

.mini-bar {
  height: 7px;
  overflow: hidden;
  border-radius: var(--ms-radius-pill);
  background: rgba(148, 163, 184, 0.16);

  span {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--ms-teal-600);
  }

  &.is-warn span {
    background: var(--ms-amber-500);
  }

  &.is-accent span {
    background: var(--ms-violet-500);
  }
}

.replay-panel {
  flex: 1 1 360px;
  display: flex;
  flex-direction: column;
}

.replay-panel :deep(.result-inspector) {
  padding: var(--ms-space-3);
}

.request-drawer {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.request-drawer-heading {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-2);
  }

  strong {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-xl);
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.drawer-section {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding-top: var(--ms-space-3);
  border-top: 1px solid var(--ms-border-light);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-2);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  strong {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  p,
  > code {
    margin: 0;
    padding: var(--ms-space-3);
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: var(--ms-panel-bg-soft);
    font-size: var(--ms-text-sm);
    line-height: 1.55;
  }

  > code {
    display: block;
    overflow: auto;
    font-family: var(--ms-font-mono);
  }
}

.diagnosis-section {
  small {
    display: block;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }

  p + p {
    margin-top: calc(var(--ms-space-2) * -1);
  }
}

.drawer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
  padding-top: var(--ms-space-3);
  border-top: 1px solid var(--ms-border-light);
}

@media (max-width: 1280px) {
  .diagnostic-grid {
    grid-template-columns: 1fr;
  }

  .diagnostic-side {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .replay-panel {
    grid-column: 1 / -1;
  }
}

@media (max-width: 900px) {
  .request-toolbar,
  .request-row,
  .request-links,
  .diagnostic-side {
    grid-template-columns: 1fr;
  }

}

@media (max-width: 640px) {
  .request-toolbar :deep(.filter-segment) {
    width: 100%;
  }
}
</style>
