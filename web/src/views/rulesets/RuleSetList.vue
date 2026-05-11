<template>
  <PageContainer title="Rulesets" eyebrow="entry">
    <template #actions>
      <el-button :icon="Refresh" :loading="store.loading" @click="loadAll">Refresh</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreateDialog">Create ruleset</el-button>
    </template>

    <div class="ruleset-entry-page">
      <section class="entry-toolbar panel-surface">
        <div class="toolbar-grid">
          <el-input
            v-model="filters.query"
            class="keyword-input"
            clearable
            :prefix-icon="Search"
            placeholder="Search name, id, host, path, rule"
          />
          <el-select v-model="filters.sort" placeholder="Sort">
            <el-option
              v-for="option in sortOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
          <el-button :icon="Filter" @click="filtersExpanded = !filtersExpanded">
            {{ filtersExpanded ? 'Hide filters' : activeFilterLabel }}
          </el-button>
        </div>

        <Transition name="filter-panel">
          <div v-if="filtersExpanded" class="advanced-filter-panel">
            <el-select v-model="filters.namespace" clearable filterable placeholder="Namespace">
              <el-option
                v-for="namespace in namespaceOptions"
                :key="namespace"
                :label="namespace"
                :value="namespace"
              />
            </el-select>
            <div class="filter-row">
              <FilterSegment
                v-model="filters.enabled"
                :options="enabledOptions"
                aria-label="Enabled filter"
              />
              <FilterSegment
                v-model="filters.publishState"
                :options="publishOptions"
                aria-label="Publish filter"
              />
              <FilterSegment
                v-model="filters.rulePresence"
                :options="rulePresenceOptions"
                aria-label="Rule presence filter"
              />
            </div>
          </div>
        </Transition>

        <div class="entry-summary" aria-label="ruleset state summary">
          <span
            v-for="item in summaryItems"
            :key="item.label"
            class="summary-pill"
            :class="`is-${item.tone}`"
          >
            <strong>{{ item.value }}</strong>
            {{ item.label }}
          </span>
        </div>
      </section>

      <section v-loading="store.loading" class="ruleset-list-shell panel-surface">
        <header class="list-header">
          <div>
            <span>ruleset inventory</span>
            <strong>{{ filteredRows.length }} visible</strong>
          </div>
          <small>Click a row to inspect details without leaving the list.</small>
        </header>

        <el-empty
          v-if="!filteredRows.length && !store.loading"
          description="No rulesets match the current filters"
        />
        <div v-else class="ruleset-list">
          <article
            v-for="row in filteredRows"
            :key="row.ruleSet.id"
            :class="['ruleset-row', `is-${row.publishTone}`, { active: selectedRow?.ruleSet.id === row.ruleSet.id }]"
            tabindex="0"
            @click="openDetails(row)"
            @keydown.enter.prevent="openDetails(row)"
            @keydown.space.prevent="openDetails(row)"
          >
            <div class="ruleset-main">
              <div class="ruleset-title">
                <span class="status-dot" :class="{ 'is-on': row.ruleSet.enabled }" />
                <div>
                  <strong :title="row.ruleSet.name || row.ruleSet.id">
                    {{ row.ruleSet.name || row.ruleSet.id }}
                  </strong>
                  <code :title="row.ruleSet.id">{{ row.ruleSet.id || 'unsaved' }}</code>
                </div>
              </div>
              <div class="ruleset-chips">
                <StateChip :label="row.enabledLabel" :tone="row.enabledTone" />
                <StateChip :label="row.publishLabel" :tone="row.publishTone" />
                <StateChip :label="`${row.ruleCount} rules`" tone="neutral" />
              </div>
            </div>

            <div class="ruleset-scope">
              <div>
                <span>Namespace</span>
                <code :title="row.ruleSet.namespace">{{ row.ruleSet.namespace }}</code>
              </div>
              <div>
                <span>Selectors</span>
                <SelectorValueGroup :values="row.selectorValues" :max-visible="2" />
              </div>
            </div>

            <div class="row-actions" @click.stop>
              <el-button type="primary" :icon="Right" @click="goRules(row.ruleSet)">
                Manage rules
              </el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleRowCommand(command, row)">
                <button class="icon-button" type="button" title="More actions">
                  <el-icon><MoreFilled /></el-icon>
                </button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="details">Details</el-dropdown-item>
                    <el-dropdown-item command="settings">Settings</el-dropdown-item>
                    <el-dropdown-item command="publish">Publish</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </article>
        </div>
      </section>
    </div>

    <el-drawer
      v-model="detailsVisible"
      :title="selectedRow?.ruleSet.name || selectedRow?.ruleSet.id || 'Ruleset details'"
      size="min(520px, calc(100vw - 24px))"
      append-to-body
    >
      <div v-if="selectedRow" class="detail-drawer">
        <div class="drawer-heading">
          <div>
            <span class="status-dot" :class="{ 'is-on': selectedRow.ruleSet.enabled }" />
            <strong>{{ selectedRow.ruleSet.name || selectedRow.ruleSet.id }}</strong>
            <code>{{ selectedRow.ruleSet.id }}</code>
          </div>
          <div class="drawer-chips">
            <StateChip :label="selectedRow.enabledLabel" :tone="selectedRow.enabledTone" />
            <StateChip :label="selectedRow.publishLabel" :tone="selectedRow.publishTone" />
          </div>
        </div>

        <KeyValueGrid :items="drawerFacts(selectedRow)" />

        <section class="drawer-section">
          <header>
            <span>Selector summary</span>
            <strong>{{ selectedRow.selectorTotal }}</strong>
          </header>
          <SelectorValueGroup :values="selectedRow.selectorValues" :max-visible="8" />
        </section>

        <section class="drawer-section">
          <header>
            <span>Rule inventory</span>
            <strong>{{ selectedRow.ruleCount }}</strong>
          </header>
          <div class="drawer-rule-list">
            <code v-for="rule in selectedRow.ruleSet.rules" :key="rule.id">
              {{ ruleDisplayName(rule) }} / {{ ruleTechnicalLabel(rule) }} / {{ rule.action.type }}
            </code>
            <small v-if="!selectedRow.ruleSet.rules.length">No rules yet.</small>
          </div>
        </section>

        <footer class="drawer-actions">
          <el-button :icon="Setting" @click="openEditDialog(selectedRow.ruleSet)">Settings</el-button>
          <el-button :icon="Upload" @click="publish(selectedRow.ruleSet)">Publish</el-button>
          <el-button type="primary" :icon="Right" @click="goRules(selectedRow.ruleSet)">
            Manage rules
          </el-button>
        </footer>
      </div>
    </el-drawer>

    <RuleSetSettingsDialog v-model="settingsVisible" :rule-set="editingRuleSet" @saved="onSaved" />
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Filter,
  MoreFilled,
  Plus,
  Refresh,
  Right,
  Search,
  Setting,
  Upload,
} from '@element-plus/icons-vue'
import FilterSegment, { type FilterSegmentOption } from '@/components/common/FilterSegment.vue'
import KeyValueGrid from '@/components/common/KeyValueGrid.vue'
import PageContainer from '@/components/common/PageContainer.vue'
import StateChip from '@/components/common/StateChip.vue'
import RuleSetSettingsDialog from '@/components/rulesets/RuleSetSettingsDialog.vue'
import SelectorValueGroup from '@/components/rulesets/SelectorValueGroup.vue'
import { useMockserverStore } from '@/store'
import type { RuleSet } from '@/types'
import {
  buildRulesetEntryMetrics,
  buildRulesetEntryRows,
  defaultRulesetEntryFilters,
  type EntryTone,
  type RulesetEnabledFilter,
  type RulesetEntryRow,
  type RulesetPublishFilter,
  type RulesetRulePresenceFilter,
  type RulesetSortKey,
} from '@/utils/entryLists'
import { ruleDisplayName, ruleTechnicalLabel } from '@/utils/ruleSummaries'

const router = useRouter()
const store = useMockserverStore()

const enabledOptions: Array<FilterSegmentOption & { value: RulesetEnabledFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Enabled', value: 'enabled' },
  { label: 'Disabled', value: 'disabled' },
]
const publishOptions: Array<FilterSegmentOption & { value: RulesetPublishFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Current', value: 'published_current' },
  { label: 'Changed', value: 'changed' },
  { label: 'Draft', value: 'draft_only' },
]
const rulePresenceOptions: Array<FilterSegmentOption & { value: RulesetRulePresenceFilter }> = [
  { label: 'All rules', value: 'all' },
  { label: 'Has rules', value: 'with_rules' },
  { label: 'Empty', value: 'empty' },
]
const sortOptions: Array<{ label: string; value: RulesetSortKey }> = [
  { label: 'Name', value: 'name' },
  { label: 'Namespace', value: 'namespace' },
  { label: 'Rule count', value: 'rule_count' },
  { label: 'Publish state', value: 'publish_state' },
  { label: 'Version', value: 'version' },
]

const filters = reactive(defaultRulesetEntryFilters())
const settingsVisible = ref(false)
const detailsVisible = ref(false)
const filtersExpanded = ref(false)
const selectedRuleSetId = ref('')
const editingRuleSet = ref<RuleSet | null>(null)

const allRows = computed(() => {
  return buildRulesetEntryRows(store.drafts, store.published, defaultRulesetEntryFilters())
})
const filteredRows = computed(() => buildRulesetEntryRows(store.drafts, store.published, filters))
const selectedRow = computed(() => {
  return filteredRows.value.find((row) => row.ruleSet.id === selectedRuleSetId.value) || null
})
const metrics = computed(() => buildRulesetEntryMetrics(allRows.value, filteredRows.value))
const namespaceOptions = computed(() => {
  return Array.from(new Set(store.drafts.map((item) => item.namespace).filter(Boolean))).sort()
})
const summaryItems = computed<Array<{ label: string; value: number; tone: EntryTone }>>(() => [
  { label: 'total', value: metrics.value.total, tone: 'neutral' },
  { label: 'published', value: metrics.value.published, tone: 'ok' },
  { label: 'changed', value: metrics.value.changed, tone: 'warn' },
  { label: 'draft only', value: metrics.value.draftOnly, tone: 'accent' },
  { label: 'empty', value: metrics.value.empty, tone: 'danger' },
])
const activeFilterLabel = computed(() => {
  const count = [
    filters.namespace,
    filters.enabled !== 'all',
    filters.publishState !== 'all',
    filters.rulePresence !== 'all',
  ].filter(Boolean).length
  return count ? `Filters ${count}` : 'Filters'
})

async function loadAll() {
  await Promise.all([store.fetchDrafts(), store.fetchPublished(), store.fetchProtocols()])
}

function drawerFacts(row: RulesetEntryRow) {
  return [
    { label: 'Ruleset ID', value: row.ruleSet.id, code: true },
    { label: 'Namespace', value: row.ruleSet.namespace, code: true },
    { label: 'Protocol', value: row.ruleSet.protocol },
    { label: 'Version', value: row.versionLabel },
    { label: 'Rules', value: row.ruleCount },
    { label: 'Selectors', value: row.selectorTotal },
  ]
}

function openCreateDialog() {
  editingRuleSet.value = null
  settingsVisible.value = true
}

function openEditDialog(ruleSet: RuleSet) {
  editingRuleSet.value = ruleSet
  settingsVisible.value = true
}

function openDetails(row: RulesetEntryRow) {
  selectedRuleSetId.value = row.ruleSet.id
  detailsVisible.value = true
}

function handleRowCommand(command: string, row: RulesetEntryRow) {
  if (command === 'details') {
    openDetails(row)
    return
  }
  if (command === 'settings') {
    openEditDialog(row.ruleSet)
    return
  }
  if (command === 'publish') {
    void publish(row.ruleSet)
  }
}

function onSaved(ruleSet: RuleSet) {
  editingRuleSet.value = ruleSet
  void loadAll()
}

function goRules(ruleSet: RuleSet) {
  router.push(`/rulesets/${encodeURIComponent(ruleSet.id)}/rules`)
}

async function publish(ruleSet: RuleSet) {
  try {
    const { value } = await ElMessageBox.prompt('Publish reason', 'Publish ruleset', {
      inputPlaceholder: 'release order status mock',
      confirmButtonText: 'Publish',
      cancelButtonText: 'Cancel',
    })
    await store.publishDraft(ruleSet.id, value)
    await loadAll()
  } catch (error) {
    if (typeof error === 'string' && error === 'cancel') return
    if (error instanceof Error && error.message === 'cancel') return
    ElMessage.error('Publish failed')
  }
}

onMounted(loadAll)
</script>

<style lang="scss" scoped>
.ruleset-entry-page {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: var(--ms-space-3);
}

.entry-toolbar {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
}

.toolbar-grid {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(150px, 190px) auto;
  gap: var(--ms-space-3);
  align-items: center;
}

.advanced-filter-panel {
  display: grid;
  grid-template-columns: minmax(180px, 240px) minmax(0, 1fr);
  gap: var(--ms-space-3);
  align-items: center;
  padding-top: var(--ms-space-3);
  border-top: 1px solid var(--ms-border-light);
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

.entry-summary {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

.summary-pill {
  display: inline-flex;
  align-items: baseline;
  gap: var(--ms-space-1);
  min-height: 28px;
  padding: 4px 9px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-secondary);
  background: var(--ms-control-bg);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
  }

  &.is-ok strong {
    color: var(--ms-green-600);
  }

  &.is-warn strong {
    color: var(--ms-amber-600);
  }

  &.is-accent strong {
    color: var(--ms-blue-500);
  }

  &.is-danger strong {
    color: var(--ms-red-600);
  }
}

.filter-panel-enter-active,
.filter-panel-leave-active {
  transition:
    opacity var(--ms-transition-normal),
    transform var(--ms-transition-normal);
}

.filter-panel-enter-from,
.filter-panel-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.ruleset-list-shell {
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  min-height: 52px;
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
    color: var(--ms-teal-700);
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
}

.ruleset-list {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  overflow: auto;
  padding: var(--ms-space-3);
}

.ruleset-row {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(260px, 0.85fr) minmax(360px, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-left-width: 3px;
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);
  cursor: pointer;
  transition:
    border-color var(--ms-transition-fast),
    background var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast);

  &.is-ok {
    border-left-color: var(--ms-green-500);
  }

  &.is-warn {
    border-left-color: var(--ms-amber-500);
  }

  &.is-accent {
    border-left-color: var(--ms-blue-500);
  }

  &.is-danger {
    border-left-color: var(--ms-red-500);
  }

  &:hover,
  &:focus-visible,
  &.active {
    border-color: rgba(37, 99, 235, 0.32);
    outline: none;
    background: var(--ms-panel-bg-hover);
    box-shadow: var(--ms-shadow-sm);
  }

  &.active {
    box-shadow: inset 3px 0 0 var(--ms-teal-600), var(--ms-shadow-sm);
  }
}

.ruleset-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.ruleset-title {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: var(--ms-space-2);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-1);
  }

  strong,
  code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    font-weight: var(--ms-font-bold);
  }

  code {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.ruleset-chips,
.drawer-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);
}

.ruleset-scope {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(120px, 0.35fr) minmax(0, 0.65fr);
  align-items: center;
  gap: var(--ms-space-3);

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-1);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--ms-space-2);
}

.detail-drawer {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.drawer-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);
  padding-bottom: var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);

  > div:first-child {
    min-width: 0;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: var(--ms-space-1) var(--ms-space-2);
    align-items: center;
  }

  strong,
  code {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    grid-column: 2;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.drawer-section {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg-soft);

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
}

.drawer-rule-list {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);

  code,
  small {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.drawer-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--ms-space-2);
  padding-top: var(--ms-space-3);
  border-top: 1px solid var(--ms-border-light);
}

@media (max-width: 1180px) {
  .ruleset-row {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .row-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}

@media (max-width: 760px) {
  .toolbar-grid,
  .advanced-filter-panel,
  .ruleset-scope {
    grid-template-columns: 1fr;
  }

  .list-header,
  .drawer-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .row-actions,
  .row-actions .el-button,
  .drawer-actions,
  .drawer-actions .el-button {
    width: 100%;
  }
}
</style>
