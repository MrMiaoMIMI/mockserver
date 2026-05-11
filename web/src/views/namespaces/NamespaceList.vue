<template>
  <PageContainer title="Namespaces" eyebrow="fallback">
    <template #meta>
      <span class="meta-pill">{{ metrics.shown }}/{{ metrics.total }} shown</span>
      <span class="meta-pill">{{ metrics.used }} used</span>
      <span class="meta-pill">{{ metrics.forward }} forward</span>
    </template>
    <template #actions>
      <el-button :icon="Refresh" :loading="store.loading" @click="loadAll">Refresh</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreateDialog">Create namespace</el-button>
    </template>

    <div class="namespace-entry-page">
      <section class="entry-toolbar panel-surface">
        <el-input
          v-model="filters.query"
          clearable
          :prefix-icon="Search"
          placeholder="Search id, name, fallback, ruleset"
        />
        <el-select v-model="filters.sort" placeholder="Sort">
          <el-option v-for="option in sortOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>

        <div class="filter-row">
          <div class="filter-block">
            <span>Usage</span>
            <div class="filter-group" aria-label="usage filter">
              <button
                v-for="option in usageOptions"
                :key="option.value"
                type="button"
                :class="{ active: filters.usage === option.value }"
                @click="filters.usage = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
          <div class="filter-block">
            <span>Ruleset miss</span>
            <div class="filter-group" aria-label="ruleset miss filter">
              <button
                v-for="option in fallbackOptions"
                :key="`ruleset-${option.value}`"
                type="button"
                :class="{ active: filters.rulesetMissType === option.value }"
                @click="filters.rulesetMissType = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
          <div class="filter-block">
            <span>Rule miss</span>
            <div class="filter-group" aria-label="rule miss filter">
              <button
                v-for="option in fallbackOptions"
                :key="`rule-${option.value}`"
                type="button"
                :class="{ active: filters.ruleMissType === option.value }"
                @click="filters.ruleMissType = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="summary-rail" aria-label="namespace summary">
        <div v-for="card in summaryCards" :key="card.label" :class="['summary-pill', `is-${card.tone}`]">
          <strong>{{ card.value }}</strong>
          <span>{{ card.label }}</span>
        </div>
      </section>

      <el-empty v-if="!filteredRows.length && !store.loading" description="No namespaces match the current filters" />
      <section v-else v-loading="store.loading" class="namespace-list panel-surface" aria-label="Namespace policies">
        <div class="namespace-list-header" aria-hidden="true">
          <span>Namespace</span>
          <span>Usage</span>
          <span>Ruleset miss</span>
          <span>Rule miss</span>
          <span>Linked rulesets</span>
          <span>Actions</span>
        </div>
        <article v-for="row in filteredRows" :key="row.namespace.id" class="namespace-row">
          <div class="namespace-cell identity-cell">
            <div class="namespace-title">
              <strong :title="row.displayName">{{ row.displayName }}</strong>
              <code :title="row.namespace.id">{{ row.namespace.id }}</code>
            </div>
            <p v-if="row.namespace.description" class="namespace-description">
              {{ row.namespace.description }}
            </p>
          </div>

          <div class="namespace-cell usage-cell">
            <span :class="['state-chip', row.usage.rulesetCount ? 'is-ok' : '']">
              {{ row.usage.rulesetCount ? 'used' : 'unused' }}
            </span>
            <div class="usage-metrics">
              <strong>{{ row.usage.rulesetCount }}</strong>
              <span>rulesets</span>
              <strong>{{ row.usage.ruleCount }}</strong>
              <span>rules</span>
            </div>
          </div>

          <div class="namespace-cell fallback-cell">
            <FallbackSummary :action="row.namespace.ruleset_miss_action" />
          </div>

          <div class="namespace-cell fallback-cell">
            <FallbackSummary :action="row.namespace.rule_miss_action" />
          </div>

          <div class="namespace-cell linked-cell">
            <div v-if="row.usage.rulesets.length" class="linked-rulesets">
              <button
                v-for="ruleSet in visibleUsage(row)"
                :key="ruleSet.id"
                type="button"
                :title="ruleSet.id"
                @click="goRuleset(ruleSet.id)"
              >
                {{ ruleSet.name }}
              </button>
              <el-popover
                v-if="hiddenUsageCount(row)"
                trigger="click"
                placement="bottom-start"
                width="320"
                popper-class="namespace-ruleset-popover"
              >
                <template #reference>
                  <button type="button" class="more-rulesets">+{{ hiddenUsageCount(row) }}</button>
                </template>
                <div class="ruleset-popover-list">
                  <button
                    v-for="ruleSet in hiddenUsage(row)"
                    :key="ruleSet.id"
                    type="button"
                    :title="ruleSet.id"
                    @click="goRuleset(ruleSet.id)"
                  >
                    <strong>{{ ruleSet.name }}</strong>
                    <code>{{ ruleSet.id }}</code>
                  </button>
                </div>
              </el-popover>
            </div>
            <code v-else class="empty-value">none</code>
          </div>

          <div class="namespace-cell action-cell">
            <el-button :icon="EditPen" @click="openEditDialog(row.namespace)">Edit policy</el-button>
          </div>
        </article>
      </section>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editingNamespace ? 'Edit namespace' : 'Create namespace'"
      width="min(1040px, calc(100vw - 32px))"
      append-to-body
      destroy-on-close
    >
      <el-form label-position="top" class="namespace-form">
        <div class="dialog-grid">
          <div class="dialog-main">
            <section class="dialog-section">
              <div v-if="editingNamespace" class="generated-id-row">
                <span>Namespace ID</span>
                <code>{{ form.id }}</code>
              </div>
              <div class="form-grid">
                <el-form-item label="Name">
                  <el-input v-model="form.name" placeholder="default" />
                </el-form-item>
                <el-form-item label="Description">
                  <el-input v-model="form.description" placeholder="optional" />
                </el-form-item>
              </div>
            </section>

            <el-alert v-if="validationIssues.length" type="warning" :closable="false">
              <ul>
                <li v-for="issue in validationIssues" :key="issue">{{ issue }}</li>
              </ul>
            </el-alert>

            <div class="fallback-grid">
              <FallbackEditor v-model="form.rulesetMiss" title="Ruleset Miss" />
              <FallbackEditor v-model="form.ruleMiss" title="Rule Miss" />
            </div>
          </div>

          <aside class="policy-preview">
            <span>Fallback preview</span>
            <div class="preview-row">
              <small>ruleset miss</small>
              <FallbackSummary :action="previewActions.rulesetMiss" />
            </div>
            <div class="preview-row">
              <small>rule miss</small>
              <FallbackSummary :action="previewActions.ruleMiss" />
            </div>
            <p>
              Forward keeps the original request flowing to the upstream service. Response returns the
              configured mock response when no match is found.
            </p>
          </aside>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="store.saving" @click="submit">Save</el-button>
      </template>
    </el-dialog>
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { EditPen, Plus, Refresh, Search } from '@element-plus/icons-vue'
import PageContainer from '@/components/common/PageContainer.vue'
import FallbackEditor from '@/components/namespaces/FallbackEditor.vue'
import FallbackSummary from '@/components/namespaces/FallbackSummary.vue'
import { useMockserverStore } from '@/store'
import type {
  NamespaceConfig,
  NamespaceFallbackAction,
  NamespaceFallbackForm,
} from '@/types'
import {
  buildNamespaceEntryMetrics,
  buildNamespaceEntryRows,
  defaultNamespaceEntryFilters,
  type EntryTone,
  type NamespaceEntryRow,
  type NamespaceFallbackFilter,
  type NamespaceSortKey,
  type NamespaceUsageFilter,
} from '@/utils/entryLists'

interface NamespaceForm {
  id: string
  name: string
  description: string
  rulesetMiss: NamespaceFallbackForm
  ruleMiss: NamespaceFallbackForm
}

const router = useRouter()
const store = useMockserverStore()

const fallbackOptions: Array<{ label: string; value: NamespaceFallbackFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Forward', value: 'forward' },
  { label: 'Response', value: 'response' },
]
const usageOptions: Array<{ label: string; value: NamespaceUsageFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Used', value: 'used' },
  { label: 'Unused', value: 'unused' },
]
const sortOptions: Array<{ label: string; value: NamespaceSortKey }> = [
  { label: 'ID', value: 'id' },
  { label: 'Usage', value: 'usage' },
  { label: 'Ruleset miss', value: 'ruleset_miss' },
  { label: 'Rule miss', value: 'rule_miss' },
]
const filters = reactive(defaultNamespaceEntryFilters())
const dialogVisible = ref(false)
const editingNamespace = ref<NamespaceConfig | null>(null)

const form = reactive<NamespaceForm>({
  id: '',
  name: '',
  description: '',
  rulesetMiss: createFallbackForm(),
  ruleMiss: createFallbackForm(),
})

const allRows = computed(() => buildNamespaceEntryRows(store.namespaces, store.drafts, defaultNamespaceEntryFilters()))
const filteredRows = computed(() => buildNamespaceEntryRows(store.namespaces, store.drafts, filters))
const metrics = computed(() => buildNamespaceEntryMetrics(allRows.value, filteredRows.value))
const summaryCards = computed<Array<{ label: string; value: number; tone: EntryTone }>>(() => [
  { label: 'Namespaces', value: metrics.value.total, tone: 'neutral' },
  { label: 'Used', value: metrics.value.used, tone: 'ok' },
  { label: 'Unused', value: metrics.value.unused, tone: 'neutral' },
  { label: 'Forward fallback', value: metrics.value.forward, tone: 'warn' },
  { label: 'Response fallback', value: metrics.value.response, tone: 'accent' },
])
const validationIssues = computed(() => validateForm())
const previewActions = computed(() => ({
  rulesetMiss: fallbackFormToAction(form.rulesetMiss),
  ruleMiss: fallbackFormToAction(form.ruleMiss),
}))

async function loadAll() {
  await Promise.all([store.fetchNamespaces(), store.fetchDrafts()])
}

function openCreateDialog() {
  editingNamespace.value = null
  resetForm(null)
  dialogVisible.value = true
}

function openEditDialog(namespace: NamespaceConfig) {
  editingNamespace.value = namespace
  resetForm(namespace)
  dialogVisible.value = true
}

async function submit() {
  if (validationIssues.value.length) {
    ElMessage.error(validationIssues.value[0])
    return
  }
  const payload: NamespaceConfig = {
    id: editingNamespace.value ? form.id.trim().toLowerCase() : '',
    name: form.name.trim(),
    description: form.description.trim(),
    ruleset_miss_action: fallbackFormToAction(form.rulesetMiss),
    rule_miss_action: fallbackFormToAction(form.ruleMiss),
  }
  await store.saveNamespace(payload)
  dialogVisible.value = false
  await loadAll()
}

function resetForm(namespace: NamespaceConfig | null) {
  form.id = namespace?.id || ''
  form.name = namespace?.name || ''
  form.description = namespace?.description || ''
  form.rulesetMiss = actionToFallbackForm(namespace?.ruleset_miss_action)
  form.ruleMiss = actionToFallbackForm(namespace?.rule_miss_action)
}

function visibleUsage(row: NamespaceEntryRow) {
  return row.usage.rulesets.slice(0, 2)
}

function hiddenUsage(row: NamespaceEntryRow) {
  return row.usage.rulesets.slice(2)
}

function hiddenUsageCount(row: NamespaceEntryRow) {
  return Math.max(row.usage.rulesets.length - 2, 0)
}

function goRuleset(rulesetId: string) {
  router.push(`/rulesets/${encodeURIComponent(rulesetId)}/rules`)
}

function validateForm() {
  const issues: string[] = []
  if (editingNamespace.value && !form.id.trim()) {
    issues.push('Namespace ID is required')
  }
  if (editingNamespace.value && !/^[a-zA-Z0-9_-]+$/.test(form.id.trim())) {
    issues.push('Namespace ID can only contain letters, numbers, underscores, and hyphens')
  }
  validateFallbackForm(form.rulesetMiss, 'Ruleset miss', issues)
  validateFallbackForm(form.ruleMiss, 'Rule miss', issues)
  return issues
}

function validateFallbackForm(fallback: NamespaceFallbackForm, label: string, issues: string[]) {
  if (fallback.type === 'response') {
    if (fallback.responseStatus < 100 || fallback.responseStatus > 599) {
      issues.push(`${label} response status must be between 100 and 599`)
    }
    if (!parseHeadersJSON(fallback.responseHeaders, `${label} Response headers`, issues)) return
    parseBody(fallback.responseBody)
    return
  }
  if (fallback.forwardTimeoutMs < 0 || fallback.forwardTimeoutMs > 30000) {
    issues.push(`${label} forward timeout must be between 0 and 30000`)
  }
}

function fallbackFormToAction(fallback: NamespaceFallbackForm): NamespaceFallbackAction {
  if (fallback.type === 'forward') {
    return {
      type: 'forward',
      forward: {
        timeout_ms: fallback.forwardTimeoutMs || undefined,
      },
    }
  }
  return {
    type: 'response',
    response: {
      status: fallback.responseStatus,
      headers: parseHeadersJSON(fallback.responseHeaders, 'headers', []) || undefined,
      body: parseBody(fallback.responseBody),
    },
  }
}

function actionToFallbackForm(action?: NamespaceFallbackAction): NamespaceFallbackForm {
  const fallback = createFallbackForm()
  if (!action) return fallback
  fallback.type = action.type
  if (action.type === 'forward') {
    fallback.forwardTimeoutMs = action.forward?.timeout_ms || 0
    return fallback
  }
  fallback.responseStatus = action.response?.status || 404
  fallback.responseHeaders = prettyJSON(action.response?.headers || {})
  fallback.responseBody = prettyJSON(action.response?.body ?? { message: 'no mock matched' })
  return fallback
}

function createFallbackForm(): NamespaceFallbackForm {
  return {
    type: 'forward',
    responseStatus: 404,
    responseHeaders: '{}',
    responseBody: '{\n  "message": "no mock matched"\n}',
    forwardTimeoutMs: 5000,
  }
}

function parseHeadersJSON(value: string, label: string, issues: string[]) {
  const trimmed = value.trim()
  if (!trimmed) return {}
  try {
    const parsed = JSON.parse(trimmed)
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      issues.push(`${label} must be a JSON object`)
      return null
    }
    const headers: Record<string, string[]> = {}
    for (const [key, rawValue] of Object.entries(parsed)) {
      if (typeof rawValue === 'string') {
        headers[key] = [rawValue]
        continue
      }
      if (Array.isArray(rawValue) && rawValue.every((item) => typeof item === 'string')) {
        headers[key] = rawValue
        continue
      }
      issues.push(`${label}.${key} must be string or string[]`)
      return null
    }
    return headers
  } catch {
    issues.push(`${label} is not valid JSON`)
    return null
  }
}

function parseBody(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return ''
  try {
    return JSON.parse(trimmed)
  } catch {
    return value
  }
}

function prettyJSON(value: unknown) {
  if (value === undefined || value === null) return ''
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

onMounted(loadAll)
</script>

<style lang="scss" scoped>
.namespace-entry-page {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  gap: var(--ms-space-3);
}

.entry-toolbar {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) minmax(160px, 210px);
  gap: var(--ms-space-3);
  align-items: center;
  padding: var(--ms-space-3);
}

.filter-row {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-3);
}

.filter-block {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);

  > span {
    flex: 0 0 auto;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
  }
}

.filter-group {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  button {
    height: 32px;
    padding: 0 var(--ms-space-3);
    border: none;
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-tertiary);
    background: transparent;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    cursor: pointer;
    transition:
      color var(--ms-transition-fast),
      background var(--ms-transition-fast);

    &.active {
      color: var(--ms-text-inverse);
      background: var(--ms-teal-600);
      box-shadow: var(--ms-shadow-xs);
    }

    &:hover:not(.active) {
      color: var(--ms-teal-700);
      background: var(--ms-teal-50);
    }
  }
}

.summary-rail {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

.summary-pill {
  min-width: 0;
  display: inline-flex;
  align-items: baseline;
  gap: var(--ms-space-1);
  padding: 6px 10px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  background: var(--ms-panel-bg-soft);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);

  strong {
    color: var(--ms-text-primary);
  }

  &.is-ok strong {
    color: var(--ms-green-600);
  }

  &.is-warn strong {
    color: var(--ms-amber-600);
  }

  &.is-danger strong {
    color: var(--ms-red-600);
  }

  &.is-accent strong {
    color: var(--ms-blue-500);
  }
}

.namespace-list {
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: auto;
  padding: 0;
}

.namespace-list-header,
.namespace-row {
  min-width: 1080px;
  display: grid;
  grid-template-columns:
    minmax(220px, 1.35fr)
    minmax(148px, 0.78fr)
    minmax(180px, 0.95fr)
    minmax(180px, 0.95fr)
    minmax(240px, 1.2fr)
    116px;
  gap: var(--ms-space-3);
  align-items: center;
}

.namespace-list-header {
  position: sticky;
  top: 0;
  z-index: 1;
  padding: 11px var(--ms-space-4);
  border-bottom: 1px solid var(--ms-border-light);
  background: var(--ms-panel-bg);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-xs);
  font-weight: var(--ms-font-bold);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.namespace-row {
  min-height: 92px;
  padding: var(--ms-space-3) var(--ms-space-4);
  border-bottom: 1px solid var(--ms-border-light);
  transition:
    background var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast);

  &:hover {
    background: var(--ms-panel-bg-hover);
    box-shadow: inset 3px 0 0 var(--ms-teal-600);
  }
}

.namespace-cell {
  min-width: 0;
  display: flex;
  align-items: center;
}

.identity-cell,
.usage-cell,
.fallback-cell,
.linked-cell {
  align-items: flex-start;
}

.identity-cell,
.usage-cell {
  flex-direction: column;
  gap: var(--ms-space-1);
}

.namespace-title {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;

  strong {
    overflow: hidden;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    font-weight: var(--ms-font-bold);
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.state-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 9px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-tertiary);
  background: var(--ms-control-bg);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  line-height: 1.35;

  &.is-ok {
    color: var(--ms-green-600);
    background: var(--ms-green-50);
  }

  &.is-warn {
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
  }

  &.is-accent {
    color: var(--ms-blue-500);
    background: var(--ms-blue-50);
  }
}

.namespace-description {
  display: -webkit-box;
  overflow: hidden;
  max-width: 100%;
  margin: 0;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
}

.usage-metrics {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 3px 5px;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
    font-weight: var(--ms-font-bold);
  }
}

.fallback-cell {
  min-width: 0;
}

.fallback-cell :deep(.fallback-summary) {
  max-width: 100%;
}

.linked-rulesets {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  button {
    max-width: 132px;
    height: 28px;
    padding: 0 9px;
    overflow: hidden;
    border: 1px solid transparent;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
    transition:
      border-color var(--ms-transition-fast),
      background var(--ms-transition-fast);

    &:hover {
      border-color: rgba(13, 148, 136, 0.32);
      background: var(--ms-teal-100);
    }
  }
}

.linked-rulesets .more-rulesets {
  color: var(--ms-text-secondary);
  background: var(--ms-control-bg);
}

.empty-value {
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
}

.action-cell {
  justify-content: flex-end;
}

:global(.namespace-ruleset-popover) .ruleset-popover-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

:global(.namespace-ruleset-popover) .ruleset-popover-list button {
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: var(--ms-space-2);
  border: none;
  border-radius: var(--ms-radius-md);
  background: transparent;
  text-align: left;
  cursor: pointer;

  &:hover {
    background: var(--ms-teal-50);
  }

  strong,
  code {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
  }

  code {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-xs);
  }
}

.namespace-form {
  min-width: 0;

  ul {
    margin: 0;
    padding-left: var(--ms-space-4);
  }
}

.dialog-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 260px;
  gap: var(--ms-space-4);
}

.dialog-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-4);
}

.dialog-section,
.policy-preview {
  padding: var(--ms-space-4);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
}

.form-grid,
.fallback-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ms-space-4);
}

.generated-id-row {
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);
  margin-bottom: var(--ms-space-3);
  padding: var(--ms-space-2) var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg);
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);

  code {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
  }
}

.policy-preview {
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);

  > span {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    font-weight: var(--ms-font-bold);
  }

  p {
    margin: 0;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.55;
  }
}

.preview-row {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-transform: uppercase;
  }
}

@media (max-width: 900px) {
  .entry-toolbar,
  .dialog-grid,
  .form-grid,
  .fallback-grid {
    grid-template-columns: 1fr;
  }

  .filter-block {
    align-items: flex-start;
    flex-direction: column;
  }

  .filter-group {
    width: 100%;
  }

  .filter-group {
    overflow-x: auto;
  }

  .filter-group button {
    flex: 1 0 auto;
  }
}
</style>
