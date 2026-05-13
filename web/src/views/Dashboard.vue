<template>
  <PageContainer title="Traffic Inspector" eyebrow="sdk traffic">
    <template #meta>
      <span class="meta-pill">
        <span class="status-dot is-on" />
        {{ formatNumber(total) }} events
      </span>
      <span class="meta-pill">match {{ matchRate }}%</span>
      <span class="meta-pill">{{ formatNumber(fallbackTotal) }} fallback</span>
    </template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">Refresh</el-button>
    </template>

    <div class="traffic-page">
      <section class="traffic-main panel-surface">
        <header class="panel-heading">
          <div>
            <span>SDK decisions</span>
            <strong>{{ formatNumber(filteredEvents.length) }}</strong>
          </div>
          <small>{{ rangeLabel }}</small>
        </header>

        <div class="traffic-toolbar">
          <el-input
            v-model="filters.query"
            clearable
            :prefix-icon="Search"
            placeholder="Search trace, event, ruleset, rule, namespace"
          />
          <el-select v-model="filters.timeRange" placeholder="Range" @change="loadData">
            <el-option label="Last hour" value="1h" />
            <el-option label="Last 24h" value="24h" />
            <el-option label="Last 7d" value="7d" />
            <el-option label="All" value="all" />
          </el-select>
          <el-select v-model="filters.protocol" clearable filterable placeholder="Protocol" @change="loadData">
            <el-option v-for="protocol in protocolOptions" :key="protocol" :label="protocol" :value="protocol" />
          </el-select>
          <el-select v-model="filters.namespace" clearable filterable placeholder="Namespace" @change="loadData">
            <el-option
              v-for="namespace in namespaceOptions"
              :key="namespace"
              :label="namespace"
              :value="namespace"
            />
          </el-select>
          <FilterSegment v-model="filters.outcome" :options="outcomeOptions" @update:model-value="loadData" />
          <div class="field-filter">
            <el-select v-model="filters.fieldPath" clearable filterable placeholder="Indexed field">
              <el-option
                v-for="option in fieldOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-input v-model="filters.fieldValue" clearable placeholder="Exact value" @keyup.enter="loadData" />
            <el-button :icon="Search" @click="loadData">Apply</el-button>
          </div>
        </div>

        <el-table
          v-loading="loading"
          class="traffic-table"
          :data="filteredEvents"
          height="100%"
          row-key="id"
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
          <el-table-column label="Protocol" min-width="120">
            <template #default="{ row }">
              <StateChip :label="row.protocol_name || '-'" tone="primary" />
            </template>
          </el-table-column>
          <el-table-column label="Namespace" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <code>{{ row.namespace_id || 'default' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Operation" min-width="130" show-overflow-tooltip>
            <template #default="{ row }">
              <code>{{ row.operation_name || '-' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Outcome" min-width="130">
            <template #default="{ row }">
              <StateChip :label="row.outcome" :tone="outcomeTone(row.outcome)" />
            </template>
          </el-table-column>
          <el-table-column label="Decision" min-width="110">
            <template #default="{ row }">
              <code>{{ row.decision_kind || '-' }}</code>
            </template>
          </el-table-column>
          <el-table-column label="Rule" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="rule-cell">
                <code>{{ row.ruleset_id || '-' }}</code>
                <small>{{ row.rule_id || row.fallback_reason || '-' }}</small>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Trace" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <button v-if="row.trace_id" class="trace-button" type="button" @click.stop="copyTrace(row.trace_id)">
                {{ shortText(row.trace_id, 18) }}
              </button>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <aside class="traffic-side">
        <section class="side-panel panel-surface">
          <header class="panel-heading compact">
            <span>outcomes</span>
            <strong>{{ outcomeEntries.length }}</strong>
          </header>
          <MetricCard
            v-for="entry in outcomeEntries"
            :key="entry.key"
            :label="entry.key"
            :value="formatNumber(entry.value)"
            :caption="`${entry.percent}% of filtered traffic`"
            :tone="metricTone(entry.key)"
          />
          <el-empty v-if="!outcomeEntries.length" description="No traffic events" />
        </section>

        <section class="side-panel panel-surface">
          <header class="panel-heading compact">
            <span>protocols</span>
            <strong>{{ protocolEntries.length }}</strong>
          </header>
          <div v-if="protocolEntries.length" class="compact-list">
            <button
              v-for="entry in protocolEntries"
              :key="entry.key"
              class="compact-row as-button"
              type="button"
              @click="setProtocol(entry.key)"
            >
              <code>{{ entry.key }}</code>
              <div class="mini-bar"><span :style="{ width: entry.width }" /></div>
              <strong>{{ entry.value }}</strong>
            </button>
          </div>
          <el-empty v-else description="No protocol data" />
        </section>

        <section class="side-panel panel-surface">
          <header class="panel-heading compact">
            <span>namespaces</span>
            <strong>{{ namespaceEntries.length }}</strong>
          </header>
          <div v-if="namespaceEntries.length" class="compact-list">
            <button
              v-for="entry in namespaceEntries"
              :key="entry.key"
              class="compact-row as-button"
              type="button"
              @click="setNamespace(entry.key)"
            >
              <code>{{ entry.key }}</code>
              <div class="mini-bar is-accent"><span :style="{ width: entry.width }" /></div>
              <strong>{{ entry.value }}</strong>
            </button>
          </div>
          <el-empty v-else description="No namespace data" />
        </section>

        <section class="side-panel panel-surface">
          <header class="panel-heading compact">
            <span>replay result</span>
            <strong>{{ replaySourceLabel }}</strong>
          </header>
          <ResultInspector :raw-json="replayJson" />
        </section>
      </aside>
    </div>

    <el-drawer
      v-model="detailVisible"
      :title="selectedEvent ? `Traffic #${selectedEvent.id}` : 'Traffic detail'"
      size="min(640px, calc(100vw - 24px))"
      append-to-body
    >
      <div v-if="selectedEvent" class="traffic-drawer">
        <header class="drawer-heading">
          <div>
            <StateChip :label="selectedEvent.protocol_name || '-'" tone="primary" />
            <strong>{{ selectedEvent.namespace_id || 'default' }}</strong>
            <small>{{ formatTime(selectedEvent.event_time) }} / {{ formatDuration(selectedEvent.duration_ms) }}</small>
          </div>
          <StateChip :label="selectedEvent.outcome" :tone="outcomeTone(selectedEvent.outcome)" />
        </header>

        <KeyValueGrid :items="detailFacts(selectedEvent)" />

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

        <el-tabs model-value="event" class="json-tabs">
          <el-tab-pane label="Event" name="event">
            <ResultInspector :raw-json="jsonString(selectedEvent.event)" :show-raw="false" />
          </el-tab-pane>
          <el-tab-pane label="Decision" name="decision">
            <ResultInspector :raw-json="jsonString(selectedEvent.decision)" :show-raw="false" />
          </el-tab-pane>
          <el-tab-pane label="Explain" name="explain">
            <ResultInspector :raw-json="jsonString(selectedEvent.explain || {})" :show-raw="false" />
          </el-tab-pane>
        </el-tabs>

        <footer class="drawer-actions">
          <el-button
            type="primary"
            :icon="VideoPlay"
            :loading="replayLoadingId === selectedEvent.id"
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

type TimeRange = '1h' | '24h' | '7d' | 'all'
type ChipTone = 'neutral' | 'ok' | 'warn' | 'danger' | 'accent' | 'primary'

interface Entry {
  key: string
  value: number
  width: string
  percent: number
}

const router = useRouter()
const store = useMockserverStore()
const loading = ref(false)
const selectedEventId = ref<number | null>(null)
const detailVisible = ref(false)
const replayLoadingId = ref<number | null>(null)
const replaySource = ref<TrafficEvent | null>(null)
const replayJson = ref('')

const filters = reactive({
  query: '',
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
const filteredEvents = computed(() => {
  const query = filters.query.trim().toLowerCase()
  if (!query) return events.value
  return events.value.filter((event) => {
    return [
      event.event_id,
      event.trace_id,
      event.namespace_id,
      event.operation_name,
      event.ruleset_id,
      event.rule_id,
      event.fallback_reason,
      event.protocol_name,
      event.outcome,
    ].some((value) => String(value || '').toLowerCase().includes(query))
  })
})
const selectedEvent = computed(() => {
  return events.value.find((event) => event.id === selectedEventId.value) || null
})
const protocolOptions = computed(() => Object.keys(stats.value?.by_protocol || {}).sort())
const namespaceOptions = computed(() => Object.keys(stats.value?.by_namespace || {}).sort())
const outcomeEntries = computed(() => rankedEntries(stats.value?.by_outcome || {}, total.value))
const protocolEntries = computed(() => rankedEntries(stats.value?.by_protocol || {}, total.value))
const namespaceEntries = computed(() => rankedEntries(stats.value?.by_namespace || {}, total.value))
const fallbackTotal = computed(() => stats.value?.by_outcome?.fallback || 0)
const matchRate = computed(() => {
  if (!total.value) return 0
  return Math.round(((stats.value?.by_outcome?.matched || 0) / total.value) * 100)
})
const fieldOptions = computed(() => indexedFieldOptions(filters.protocol))
const rangeLabel = computed(() => {
  if (filters.timeRange === 'all') return 'all retained traffic'
  if (filters.timeRange === '1h') return 'last hour'
  if (filters.timeRange === '24h') return 'last 24h'
  return 'last 7 days'
})
const replaySourceLabel = computed(() => replaySource.value ? `#${replaySource.value.id}` : 'idle')

async function loadData() {
  loading.value = true
  try {
    const range = timeRangeSeconds(filters.timeRange)
    await store.fetchTrafficEvents({
      limit: 100,
      start_time: range.start,
      end_time: range.end,
      protocol_name: filters.protocol || undefined,
      namespace_id: filters.namespace || undefined,
      outcome: filters.outcome || undefined,
      field_path: filters.fieldPath || undefined,
      field_value: filters.fieldPath ? filters.fieldValue : undefined,
      include_indexes: true,
    })
  } finally {
    loading.value = false
  }
}

function openEvent(row: TrafficEvent) {
  selectedEventId.value = row.id
  detailVisible.value = true
}

async function replay(event: TrafficEvent) {
  replayLoadingId.value = event.id
  replaySource.value = event
  selectedEventId.value = event.id
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
    replayLoadingId.value = null
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
  loadData()
}

function setProtocol(protocol: string) {
  filters.protocol = protocol
  loadData()
}

function setNamespace(namespace: string) {
  filters.namespace = namespace
  loadData()
}

function detailFacts(event: TrafficEvent) {
  return [
    { label: 'Event ID', value: event.event_id, code: true },
    { label: 'Trace', value: event.trace_id || '-', code: true },
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

function rankedEntries(values: Record<string, number>, totalValue: number): Entry[] {
  const max = Math.max(...Object.values(values), 1)
  return Object.entries(values)
    .filter(([, value]) => value > 0)
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .slice(0, 8)
    .map(([key, value]) => ({
      key: key || '-',
      value,
      width: `${Math.max((value / max) * 100, 6)}%`,
      percent: totalValue ? Math.round((value / totalValue) * 100) : 0,
    }))
}

function indexedFieldOptions(protocol: string) {
  const base = [
    { label: 'event.operation', value: 'event.operation' },
    { label: 'decision.response.status', value: 'decision.response.status' },
    { label: 'decision.forward.timeout_ms', value: 'decision.forward.timeout_ms' },
  ]
  if (protocol === 'cache') {
    return [
      { label: 'event.request.operation', value: 'event.request.operation' },
      { label: 'event.request.key', value: 'event.request.key' },
      ...base,
    ]
  }
  if (protocol === 'http' || protocol === '') {
    return [
      { label: 'event.request.method', value: 'event.request.method' },
      { label: 'event.request.host', value: 'event.request.host' },
      { label: 'event.request.original_host', value: 'event.request.original_host' },
      { label: 'event.request.path', value: 'event.request.path' },
      ...base,
    ]
  }
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

function outcomeTone(outcome: string): ChipTone {
  if (outcome === 'matched') return 'ok'
  if (outcome === 'fallback') return 'warn'
  if (outcome === 'error') return 'danger'
  if (outcome === 'unmatched') return 'neutral'
  return 'accent'
}

function metricTone(outcome: string): ChipTone {
  if (outcome === 'matched') return 'ok'
  if (outcome === 'fallback') return 'warn'
  if (outcome === 'error') return 'danger'
  return 'accent'
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
.traffic-page {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(310px, 0.3fr);
  gap: var(--ms-space-3);
}

.traffic-main,
.traffic-side {
  min-height: 0;
}

.traffic-main {
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  overflow: hidden;
}

.traffic-side {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  overflow: auto;
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

.side-panel {
  overflow: hidden;
}

.side-panel :deep(.metric-card) {
  margin: var(--ms-space-3);
}

.compact-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
}

.compact-row {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(80px, 110px) auto;
  align-items: center;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg);

  &.as-button {
    width: 100%;
    text-align: left;
    cursor: pointer;
  }

  code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
  }
}

.mini-bar {
  height: 7px;
  overflow: hidden;
  border-radius: var(--ms-radius-pill);
  background: var(--ms-panel-bg-soft);

  span {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--ms-green-500);
  }

  &.is-accent span {
    background: var(--ms-blue-500);
  }
}

.traffic-drawer {
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

.json-tabs {
  min-width: 0;
}

.drawer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

@media (max-width: 1180px) {
  .traffic-page {
    grid-template-columns: 1fr;
  }

  .traffic-side {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .traffic-page {
    height: auto;
    min-height: 100%;
  }

  .traffic-main {
    grid-template-rows: auto auto minmax(260px, 1fr);
    overflow: visible;
  }

  .traffic-toolbar,
  .field-filter,
  .traffic-side {
    grid-template-columns: 1fr;
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
}
</style>
