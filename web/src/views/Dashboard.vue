<template>
  <PageContainer title="Traffic Log" eyebrow="sdk traffic">
    <template #meta>
      <span class="meta-pill">
        <span class="status-dot is-on" />
        {{ formatNumber(total) }} events
      </span>
      <span class="meta-pill">match {{ matchRate }}%</span>
      <span class="meta-pill">{{ pageRangeLabel }}</span>
    </template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">Refresh</el-button>
    </template>

    <div class="traffic-log-page">
      <section class="summary-strip">
        <MetricCard
          v-for="metric in summaryMetrics"
          :key="metric.label"
          :label="metric.label"
          :value="metric.value"
          :caption="metric.caption"
          :tone="metric.tone"
        />
      </section>

      <section class="traffic-log-panel panel-surface">
        <header class="panel-heading">
          <div>
            <span>events</span>
            <strong>{{ pageRangeLabel }}</strong>
          </div>
          <el-button @click="clearFilters">Reset</el-button>
        </header>

        <div class="traffic-toolbar">
          <el-input
            v-model="filters.lookup"
            clearable
            :prefix-icon="Search"
            placeholder="Trace ID or event ID"
            @keyup.enter="resetAndLoad"
            @clear="resetAndLoad"
          />
          <el-select v-model="filters.timeRange" placeholder="Range" @change="resetAndLoad">
            <el-option label="Last hour" value="1h" />
            <el-option label="Last 24h" value="24h" />
            <el-option label="Last 7d" value="7d" />
            <el-option label="All" value="all" />
          </el-select>
          <el-select v-model="filters.protocol" clearable filterable placeholder="Protocol" @change="resetAndLoad">
            <el-option v-for="protocol in protocolOptions" :key="protocol" :label="protocol" :value="protocol" />
          </el-select>
          <el-select v-model="filters.namespace" clearable filterable placeholder="Namespace" @change="resetAndLoad">
            <el-option
              v-for="namespace in namespaceOptions"
              :key="namespace"
              :label="namespace"
              :value="namespace"
            />
          </el-select>
          <FilterSegment v-model="filters.outcome" :options="outcomeOptions" @update:model-value="resetAndLoad" />
          <div class="field-filter">
            <el-select v-model="filters.fieldPath" clearable filterable placeholder="Indexed field" @change="resetAndLoad">
              <el-option
                v-for="option in fieldOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-input v-model="filters.fieldValue" clearable placeholder="Exact value" @keyup.enter="resetAndLoad" />
            <el-button :icon="Search" @click="resetAndLoad">Apply</el-button>
          </div>
        </div>

        <el-table
          v-loading="loading"
          class="traffic-table"
          :data="events"
          height="100%"
          row-key="id"
          :row-class-name="rowClassName"
          @row-click="openEvent"
        >
          <el-table-column label="Time" min-width="150">
            <template #default="{ row }">
              <div class="time-cell">
                <strong>{{ formatTime(row.event_time) }}</strong>
                <small>{{ formatDuration(row.duration_ms) }}</small>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Protocol" width="112">
            <template #default="{ row }">
              <StateChip :label="row.protocol_name || '-'" tone="primary" />
            </template>
          </el-table-column>
          <el-table-column label="Namespace" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <code>{{ row.namespace_id || 'default' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Operation" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <code>{{ row.operation_name || '-' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Outcome" width="132">
            <template #default="{ row }">
              <StateChip :label="row.outcome" :tone="outcomeTone(row.outcome)" />
            </template>
          </el-table-column>
          <el-table-column label="Decision" width="120">
            <template #default="{ row }">
              <code>{{ row.decision_kind || '-' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Rule" min-width="230" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="rule-cell">
                <code>{{ row.ruleset_id || '-' }}</code>
                <small>{{ row.rule_id || row.fallback_reason || '-' }}</small>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Trace" min-width="190" show-overflow-tooltip>
            <template #default="{ row }">
              <button v-if="row.trace_id" class="trace-button" type="button" @click.stop="copyTrace(row.trace_id)">
                {{ shortText(row.trace_id, 22) }}
              </button>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>

        <footer class="pagination-bar">
          <span>{{ formatNumber(events.length) }} loaded</span>
          <el-pagination
            background
            layout="sizes, prev, pager, next, jumper"
            :total="total"
            :page-size="page.size"
            :current-page="page.current"
            :page-sizes="[25, 50, 100, 200]"
            @size-change="onPageSizeChange"
            @current-change="onPageChange"
          />
        </footer>
      </section>
    </div>

    <el-drawer
      v-model="detailVisible"
      :title="selectedEvent ? `Traffic #${selectedEvent.id}` : 'Traffic detail'"
      size="min(760px, calc(100vw - 24px))"
      append-to-body
    >
      <div v-loading="detailLoading" class="traffic-drawer">
        <template v-if="selectedEvent">
          <header class="drawer-heading">
            <div>
              <StateChip :label="selectedEvent.protocol_name || '-'" tone="primary" />
              <strong>{{ selectedEvent.namespace_id || 'default' }}</strong>
              <small>{{ formatTime(selectedEvent.event_time) }} / {{ formatDuration(selectedEvent.duration_ms) }}</small>
            </div>
            <StateChip :label="selectedEvent.outcome" :tone="outcomeTone(selectedEvent.outcome)" />
          </header>

          <KeyValueGrid :items="detailFacts(selectedEvent)" />

          <section v-if="selectedSelectionDiagnostics.hasDiagnostics" class="drawer-section selection-diagnostics">
            <header>
              <span>selection diagnostics</span>
              <strong>{{ selectedSelectionDiagnostics.rulesetCandidates.length }} candidates</strong>
            </header>
            <div class="selection-summary">
              <div>
                <span>Winner ruleset</span>
                <code>{{ selectedSelectionDiagnostics.winnerRuleSetId || '-' }}</code>
              </div>
              <div>
                <span>Winner snapshot</span>
                <code>{{ selectedSelectionDiagnostics.winnerSnapshotId || '-' }}</code>
              </div>
              <div>
                <span>Winner rule</span>
                <code>{{ selectedSelectionDiagnostics.winnerRuleId || '-' }}</code>
              </div>
            </div>
            <div v-if="selectedSelectionDiagnostics.rulesetCandidates.length" class="candidate-list">
              <div
                v-for="candidate in selectedSelectionDiagnostics.rulesetCandidates"
                :key="`${candidate.rulesetId}:${candidate.snapshotId}`"
                class="candidate-row"
                :class="{ 'is-selected': candidate.selected }"
              >
                <StateChip :label="candidate.selected ? 'winner' : 'candidate'" :tone="candidate.selected ? 'ok' : 'neutral'" />
                <code>{{ candidate.rulesetId }}</code>
                <small>specificity {{ candidate.selectorSpecificity }}</small>
                <span>{{ candidate.message || '-' }}</span>
              </div>
            </div>
            <div v-if="selectedSelectionDiagnostics.candidateRuleIds.length" class="rule-candidates">
              <span>Candidate rules</span>
              <code>{{ selectedSelectionDiagnostics.candidateRuleIds.join(', ') }}</code>
            </div>
          </section>

          <section class="drawer-section">
            <header>
              <span>indexed fields</span>
              <strong>{{ selectedEvent.indexes?.length || 0 }}</strong>
            </header>
            <div v-if="selectedEvent.indexes?.length" class="index-list">
              <button
                v-for="index in selectedEvent.indexes"
                :key="`${index.field_path}:${index.field_value_hash}`"
                type="button"
                @click="applyIndexFilter(index.field_path, index.field_value_text || index.field_value_preview)"
              >
                <span>{{ index.field_path }}</span>
                <code>{{ index.field_value_preview }}</code>
              </button>
            </div>
            <el-empty v-else description="No indexed fields" />
          </section>

          <el-tabs v-model="detailTab" class="json-tabs">
            <el-tab-pane label="Event" name="event">
              <ResultInspector :raw-json="jsonString(selectedEvent.event || {})" :show-raw="false" />
            </el-tab-pane>
            <el-tab-pane label="Decision" name="decision">
              <ResultInspector :raw-json="jsonString(selectedEvent.decision || {})" :show-raw="false" />
            </el-tab-pane>
            <el-tab-pane label="Explain" name="explain">
              <ResultInspector :raw-json="jsonString(selectedEvent.explain || {})" :show-raw="false" />
            </el-tab-pane>
            <el-tab-pane label="Replay" name="replay">
              <ResultInspector :raw-json="replayJson" />
            </el-tab-pane>
          </el-tabs>

          <footer class="drawer-actions">
            <el-button
              type="primary"
              :icon="VideoPlay"
              :disabled="!selectedEvent.event"
              :loading="replayLoading"
              @click="replay(selectedEvent)"
            >
              Replay
            </el-button>
            <el-button :disabled="!selectedEvent.ruleset_id" :icon="Right" @click="openRuleset(selectedEvent)">
              Open rules
            </el-button>
            <el-button v-if="selectedEvent.trace_id" :icon="CopyDocument" @click="copyTrace(selectedEvent.trace_id)">
              Copy trace
            </el-button>
          </footer>
        </template>
      </div>
    </el-drawer>
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CopyDocument, Refresh, Right, Search, VideoPlay } from '@element-plus/icons-vue'
import FilterSegment, { type FilterSegmentOption } from '@/components/common/FilterSegment.vue'
import KeyValueGrid from '@/components/common/KeyValueGrid.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageContainer from '@/components/common/PageContainer.vue'
import StateChip from '@/components/common/StateChip.vue'
import ResultInspector from '@/components/rulesets/ResultInspector.vue'
import { useMockserverStore } from '@/store'
import type { TrafficEvent } from '@/types'
import { buildTrafficSelectionDiagnostics } from '@/utils/trafficSelectionDiagnostics'

type TimeRange = '1h' | '24h' | '7d' | 'all'
type ChipTone = 'neutral' | 'ok' | 'warn' | 'danger' | 'accent' | 'primary'

interface Entry {
  key: string
  value: number
}

const router = useRouter()
const store = useMockserverStore()
const loading = ref(false)
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailTab = ref('event')
const selectedEvent = ref<TrafficEvent | null>(null)
const replayLoading = ref(false)
const replayJson = ref('')

const page = reactive({
  current: 1,
  size: 50,
})

const filters = reactive({
  lookup: '',
  timeRange: '24h' as TimeRange,
  protocol: '',
  namespace: '',
  outcome: '',
  fieldPath: '',
  fieldValue: '',
})

const outcomeOptions: FilterSegmentOption[] = [
  { label: 'All', value: '' },
  { label: 'Matched', value: 'matched' },
  { label: 'Fallback', value: 'fallback' },
  { label: 'Unmatched', value: 'unmatched' },
  { label: 'Error', value: 'error' },
]

const total = computed(() => store.trafficTotal)
const stats = computed(() => store.trafficStats)
const events = computed(() => store.trafficEvents)
const selectedSelectionDiagnostics = computed(() => buildTrafficSelectionDiagnostics(selectedEvent.value))
const protocolOptions = computed(() => Object.keys(stats.value?.by_protocol || {}).sort())
const namespaceOptions = computed(() => Object.keys(stats.value?.by_namespace || {}).sort())
const protocolEntries = computed(() => rankedEntries(stats.value?.by_protocol || {}))
const namespaceEntries = computed(() => rankedEntries(stats.value?.by_namespace || {}))
const protocolCount = computed(() => Object.keys(stats.value?.by_protocol || {}).length)
const namespaceCount = computed(() => Object.keys(stats.value?.by_namespace || {}).length)
const fallbackTotal = computed(() => stats.value?.by_outcome?.fallback || 0)
const matchRate = computed(() => {
  if (!total.value) return 0
  return Math.round(((stats.value?.by_outcome?.matched || 0) / total.value) * 100)
})
const pageRangeLabel = computed(() => {
  if (!total.value || !events.value.length) return '0 of 0'
  const start = (page.current - 1) * page.size + 1
  const end = Math.min(start + events.value.length - 1, total.value)
  return `${formatNumber(start)}-${formatNumber(end)} of ${formatNumber(total.value)}`
})
const summaryMetrics = computed(() => [
  {
    label: 'Matched',
    value: formatNumber(stats.value?.by_outcome?.matched || 0),
    caption: `${matchRate.value}% of filtered traffic`,
    tone: 'ok' as ChipTone,
  },
  {
    label: 'Fallback',
    value: formatNumber(fallbackTotal.value),
    caption: `${outcomePercent('fallback')}% of filtered traffic`,
    tone: 'warn' as ChipTone,
  },
  {
    label: 'Errors',
    value: formatNumber(stats.value?.by_outcome?.error || 0),
    caption: `${outcomePercent('error')}% of filtered traffic`,
    tone: 'danger' as ChipTone,
  },
  {
    label: 'Protocols',
    value: formatNumber(protocolCount.value),
    caption: topEntryLabel(protocolEntries.value),
    tone: 'accent' as ChipTone,
  },
  {
    label: 'Namespaces',
    value: formatNumber(namespaceCount.value),
    caption: topEntryLabel(namespaceEntries.value),
    tone: 'primary' as ChipTone,
  },
])
const fieldOptions = computed(() => indexedFieldOptions(filters.protocol))

async function loadData() {
  loading.value = true
  try {
    const range = timeRangeSeconds(filters.timeRange)
    const lookup = filters.lookup.trim()
    const eventID = lookup.startsWith('te_') ? lookup : ''
    const traceID = lookup && !eventID ? lookup : ''
    await store.fetchTrafficEvents({
      limit: page.size,
      offset: (page.current - 1) * page.size,
      start_time: range.start,
      end_time: range.end,
      event_id: eventID || undefined,
      trace_id: traceID || undefined,
      protocol_name: filters.protocol || undefined,
      namespace_id: filters.namespace || undefined,
      outcome: filters.outcome || undefined,
      field_path: filters.fieldPath || undefined,
      field_value: filters.fieldPath ? filters.fieldValue : undefined,
    })
  } finally {
    loading.value = false
  }
}

function resetAndLoad() {
  page.current = 1
  loadData()
}

function clearFilters() {
  filters.lookup = ''
  filters.timeRange = '24h'
  filters.protocol = ''
  filters.namespace = ''
  filters.outcome = ''
  filters.fieldPath = ''
  filters.fieldValue = ''
  resetAndLoad()
}

function onPageSizeChange(size: number) {
  page.size = size
  page.current = 1
  loadData()
}

function onPageChange(current: number) {
  page.current = current
  loadData()
}

async function openEvent(row: TrafficEvent) {
  detailVisible.value = true
  detailLoading.value = true
  detailTab.value = 'event'
  replayJson.value = ''
  selectedEvent.value = row
  try {
    selectedEvent.value = await store.fetchTrafficEvent(row.id)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Load traffic detail failed')
  } finally {
    detailLoading.value = false
  }
}

async function replay(event: TrafficEvent) {
  if (!event.event) return
  replayLoading.value = true
  detailTab.value = 'replay'
  try {
    const result = await store.simulatePublished({
      event: event.event,
      explain_only: false,
      explain_summary: true,
      explain_max_depth: 3,
    })
    replayJson.value = JSON.stringify({ result }, null, 2)
  } catch (error) {
    replayJson.value = ''
    ElMessage.error(error instanceof Error ? error.message : 'Replay failed')
  } finally {
    replayLoading.value = false
  }
}

function openRuleset(event: TrafficEvent) {
  if (!event.ruleset_id) return
  router.push({
    path: `/rulesets/${encodeURIComponent(event.ruleset_id)}/rules`,
    query: {
      workbench: 'inspect',
      ...(event.rule_id ? { rule: event.rule_id } : {}),
    },
  })
}

function applyIndexFilter(path: string, value: string) {
  filters.fieldPath = path
  filters.fieldValue = value
  detailVisible.value = false
  resetAndLoad()
}

function rowClassName({ row }: { row: TrafficEvent }) {
  return selectedEvent.value?.id === row.id ? 'is-selected-traffic' : ''
}

function detailFacts(event: TrafficEvent) {
  return [
    { label: 'Event ID', value: event.event_id, code: true },
    { label: 'Trace', value: event.trace_id || '-', code: true },
    { label: 'Protocol', value: event.protocol_name || '-' },
    { label: 'Namespace', value: event.namespace_id || 'default', code: true },
    { label: 'Operation', value: event.operation_name || '-' },
    { label: 'Outcome', value: event.outcome, tone: outcomeTone(event.outcome) },
    { label: 'Decision', value: event.decision_kind || '-' },
    { label: 'Ruleset', value: event.ruleset_id || '-', code: true },
    { label: 'Rule', value: event.rule_id || '-', code: true },
    { label: 'Fallback', value: event.fallback_reason || '-' },
    { label: 'Error', value: event.error_message || '-' },
  ]
}

async function copyTrace(traceId: string) {
  try {
    await navigator.clipboard.writeText(traceId)
    ElMessage.success('Trace ID copied')
  } catch {
    ElMessage.error('Copy failed')
  }
}

function timeRangeSeconds(range: TimeRange) {
  const end = Math.floor(Date.now() / 1000)
  if (range === 'all') {
    return { start: undefined, end: undefined }
  }
  const seconds = range === '1h' ? 3600 : range === '24h' ? 86400 : 7 * 86400
  return { start: end - seconds, end }
}

function rankedEntries(values: Record<string, number>): Entry[] {
  return Object.entries(values)
    .filter(([, value]) => value > 0)
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .slice(0, 8)
    .map(([key, value]) => ({
      key: key || '-',
      value,
    }))
}

function indexedFieldOptions(protocol: string) {
  const base = [
    { label: 'decision.response.status', value: 'decision.response.status' },
  ]
  const http = [
    { label: 'event.request.method', value: 'event.request.method' },
    { label: 'event.request.host', value: 'event.request.host' },
    { label: 'event.request.path', value: 'event.request.path' },
  ]
  const cache = [
    { label: 'event.request.operation', value: 'event.request.operation' },
    { label: 'event.request.key', value: 'event.request.key' },
  ]
  const spex = [
    { label: 'event.request.cmd', value: 'event.request.cmd' },
  ]
  if (protocol === 'cache') {
    return [...cache, ...base]
  }
  if (protocol === 'spex') {
    return [...spex, ...base]
  }
  if (protocol === 'http') {
    return [...http, ...base]
  }
  if (protocol) {
    return [
      { label: 'event.request.operation', value: 'event.request.operation' },
      { label: 'event.request.service', value: 'event.request.service' },
      { label: 'event.request.method', value: 'event.request.method' },
      { label: 'event.request.topic', value: 'event.request.topic' },
      { label: 'event.request.group', value: 'event.request.group' },
      { label: 'event.request.key', value: 'event.request.key' },
      ...base,
    ]
  }
  return uniqueFieldOptions([
    ...http,
    ...spex,
    ...cache,
    { label: 'event.request.operation', value: 'event.request.operation' },
    { label: 'event.request.service', value: 'event.request.service' },
    { label: 'event.request.topic', value: 'event.request.topic' },
    { label: 'event.request.group', value: 'event.request.group' },
    ...base,
  ])
}

function uniqueFieldOptions(options: Array<{ label: string; value: string }>) {
  const seen = new Set<string>()
  return options.filter((option) => {
    if (seen.has(option.value)) return false
    seen.add(option.value)
    return true
  })
}

function outcomeTone(outcome: string): ChipTone {
  if (outcome === 'matched') return 'ok'
  if (outcome === 'fallback') return 'warn'
  if (outcome === 'error') return 'danger'
  if (outcome === 'unmatched') return 'neutral'
  return 'accent'
}

function outcomePercent(outcome: string) {
  if (!total.value) return 0
  return Math.round(((stats.value?.by_outcome?.[outcome] || 0) / total.value) * 100)
}

function topEntryLabel(entries: Entry[]) {
  if (!entries.length) return 'no data'
  return `${entries[0].key} ${formatNumber(entries[0].value)}`
}

function formatTime(seconds: number) {
  if (!seconds) return '-'
  return new Intl.DateTimeFormat('en-US', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(seconds * 1000))
}

function formatDuration(value: number) {
  return `${value || 0} ms`
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US').format(value || 0)
}

function shortText(value: string, size: number) {
  if (value.length <= size) return value
  return `${value.slice(0, size)}...`
}

function jsonString(value: unknown) {
  return JSON.stringify(value || {}, null, 2)
}

onMounted(loadData)
</script>

<style lang="scss" scoped>
.traffic-log-page {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: var(--ms-space-3);
}

.summary-strip {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: var(--ms-space-3);
}

.traffic-log-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
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

  strong {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }
}

.traffic-toolbar {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) 120px 140px 160px;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);
}

.traffic-toolbar :deep(.filter-segment),
.field-filter {
  grid-column: 1 / -1;
}

.field-filter {
  display: grid;
  grid-template-columns: minmax(200px, 260px) minmax(180px, 1fr) auto;
  gap: var(--ms-space-2);
}

.traffic-table {
  min-height: 0;
}

.traffic-table :deep(.el-table__row) {
  cursor: pointer;
}

.traffic-table :deep(.is-selected-traffic td.el-table__cell) {
  background: #eff6ff;
}

.time-cell,
.rule-cell {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;

  strong,
  code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-xs);
  }
}

code {
  color: var(--ms-text-secondary);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
}

.trace-button {
  max-width: 100%;
  padding: 0;
  border: 0;
  color: var(--ms-blue-500);
  background: transparent;
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
  cursor: pointer;
}

.pagination-bar {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding: 0 var(--ms-space-4);
  border-top: 1px solid var(--ms-border-light);

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }
}

.traffic-drawer {
  min-height: 240px;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.drawer-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-1);
  }

  strong {
    min-width: 0;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-xl);
    overflow-wrap: anywhere;
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.drawer-section {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-2);

    span {
      color: var(--ms-text-primary);
      font-size: var(--ms-text-sm);
      font-weight: var(--ms-font-bold);
      text-transform: uppercase;
    }

    strong {
      color: var(--ms-text-tertiary);
      font-size: var(--ms-text-sm);
    }
  }
}

.index-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--ms-space-2);

  button {
    min-width: 0;
    display: grid;
    grid-template-columns: minmax(170px, 0.45fr) minmax(0, 1fr);
    gap: var(--ms-space-2);
    padding: var(--ms-space-2);
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-md);
    background: var(--ms-panel-bg);
    text-align: left;
    cursor: pointer;
  }

  span,
  code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.selection-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--ms-space-2);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--ms-space-2);
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-sm);
    background: var(--ms-panel-muted);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-xs);
    font-weight: var(--ms-font-semibold);
    text-transform: uppercase;
  }
}

.candidate-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.candidate-row {
  min-width: 0;
  display: grid;
  grid-template-columns: auto minmax(120px, 0.7fr) auto minmax(0, 1fr);
  align-items: center;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-sm);
  background: var(--ms-panel-bg);

  &.is-selected {
    border-color: var(--ms-green-400);
    background: #f0fdf4;
  }

  code,
  span,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span,
  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.rule-candidates {
  min-width: 0;
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  gap: var(--ms-space-2);
  align-items: start;

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }

  code {
    min-width: 0;
    white-space: normal;
    overflow-wrap: anywhere;
  }
}

.json-tabs {
  min-width: 0;
  min-height: 360px;
}

.drawer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

@media (max-width: 1180px) {
  .summary-strip {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .traffic-log-page {
    height: auto;
    min-height: 100%;
  }

  .summary-strip,
  .traffic-toolbar,
  .field-filter {
    grid-template-columns: 1fr;
  }

  .traffic-log-panel {
    grid-template-rows: auto auto minmax(320px, 1fr) auto;
  }

  .traffic-toolbar :deep(.filter-segment) {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    width: 100%;
    overflow: visible;
  }

  .traffic-toolbar :deep(.filter-segment button) {
    width: 100%;
  }

  .selection-summary,
  .candidate-row,
  .rule-candidates {
    grid-template-columns: 1fr;
  }

  .pagination-bar {
    align-items: flex-start;
    flex-direction: column;
    padding: var(--ms-space-3);
  }
}
</style>
