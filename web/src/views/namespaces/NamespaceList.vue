<template>
  <PageContainer title="Namespaces" eyebrow="fallback">
    <template #meta>
      <span class="meta-pill">{{ metrics.shown }}/{{ metrics.total }} namespaces</span>
      <span class="meta-pill">{{ metrics.used }} used</span>
      <span class="meta-pill">{{ metrics.forward }} forward</span>
      <span class="meta-pill">{{ metrics.response }} response</span>
    </template>
    <template #actions>
      <el-button :icon="Refresh" :loading="store.loading" @click="loadAll">Refresh</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreateDialog">Create namespace</el-button>
    </template>

    <div class="namespace-page">
      <section class="namespace-toolbar panel-surface">
        <el-input
          v-model="filters.query"
          clearable
          :prefix-icon="Search"
          placeholder="Search namespace, protocol, fallback, ruleset"
        />
        <el-select v-model="filters.sort" placeholder="Sort">
          <el-option v-for="option in sortOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>

        <div class="filter-row">
          <div class="filter-block">
            <span>Protocol</span>
            <div class="filter-group" aria-label="protocol filter">
              <button
                v-for="option in protocolFilterOptions"
                :key="option.value"
                type="button"
                :class="{ active: filters.protocol === option.value }"
                @click="filters.protocol = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
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
      <section v-else v-loading="store.loading" class="namespace-list panel-surface" aria-label="Namespaces">
        <div class="namespace-list-header" aria-hidden="true">
          <span>Namespace</span>
          <span>Protocols</span>
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

          <div class="namespace-cell protocol-cell">
            <span v-for="protocol in row.protocols" :key="protocol" class="protocol-chip">{{ protocol }}</span>
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

          <div class="namespace-cell policy-cell">
            <div class="policy-stack">
              <div v-for="policy in visiblePolicies(row)" :key="`${policy.protocol}:ruleset`" class="policy-summary">
                <span class="policy-protocol">{{ policy.protocol }}</span>
                <FallbackSummary :action="policy.policy.ruleset_miss_action" />
              </div>
              <button v-if="hiddenPolicyCount(row)" type="button" class="more-policies" @click="openEditDialog(row.namespace)">
                +{{ hiddenPolicyCount(row) }} protocols
              </button>
            </div>
          </div>

          <div class="namespace-cell policy-cell">
            <div class="policy-stack">
              <div v-for="policy in visiblePolicies(row)" :key="`${policy.protocol}:rule`" class="policy-summary">
                <span class="policy-protocol">{{ policy.protocol }}</span>
                <FallbackSummary :action="policy.policy.rule_miss_action" />
              </div>
              <button v-if="hiddenPolicyCount(row)" type="button" class="more-policies" @click="openEditDialog(row.namespace)">
                +{{ hiddenPolicyCount(row) }} protocols
              </button>
            </div>
          </div>

          <div class="namespace-cell linked-cell">
            <div v-if="row.usage.rulesets.length" class="linked-rulesets">
              <button
                v-for="ruleSet in visibleUsage(row)"
                :key="ruleSet.id"
                type="button"
                :title="`${ruleSet.name} · ${ruleSet.protocol}`"
                @click="goRuleset(ruleSet.id)"
              >
                <span>{{ ruleSet.name }}</span>
                <code>{{ ruleSet.protocol }}</code>
              </button>
              <el-popover
                v-if="hiddenUsageCount(row)"
                trigger="click"
                placement="bottom-start"
                width="340"
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
                    <span>{{ ruleSet.protocol.toUpperCase() }}</span>
                    <code>{{ ruleSet.id }}</code>
                  </button>
                </div>
              </el-popover>
            </div>
            <code v-else class="empty-value">none</code>
          </div>

          <div class="namespace-cell action-cell">
            <el-button :icon="EditPen" @click="openEditDialog(row.namespace)">Manage</el-button>
          </div>
        </article>
      </section>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editingNamespace ? 'Manage namespace' : 'Create namespace'"
      width="min(1480px, calc(100vw - 24px))"
      append-to-body
      destroy-on-close
    >
      <el-form label-position="top" class="namespace-form">
        <div class="dialog-grid">
          <section class="profile-section">
            <header class="section-heading">
              <strong>Profile</strong>
              <span>Namespace-level fields shared by every protocol policy.</span>
            </header>
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

          <section class="policy-section">
            <header class="section-heading">
              <strong>Protocol policies</strong>
              <span>Fallback behavior is scoped per protocol under this namespace.</span>
            </header>
            <el-tabs v-model="form.activeProtocol" class="policy-tabs" type="card" @tab-change="onPolicyTabChange">
              <el-tab-pane
                v-for="option in protocolOptions"
                :key="option.value"
                :name="option.value"
                :label="option.label"
              />
            </el-tabs>

            <div class="policy-workspace">
              <div class="policy-context">
                <span class="protocol-chip">{{ form.activeProtocol }}</span>
                <div>
                  <strong>{{ activePolicyUsage.rulesetCount }}</strong>
                  <span>rulesets</span>
                  <strong>{{ activePolicyUsage.ruleCount }}</strong>
                  <span>rules</span>
                </div>
              </div>

              <el-alert v-if="validationIssues.length" type="warning" :closable="false">
                <ul>
                  <li v-for="issue in validationIssues" :key="issue">{{ issue }}</li>
                </ul>
              </el-alert>

              <div class="fallback-grid">
                <FallbackEditor
                  v-model="activePolicy.rulesetMiss"
                  title="Ruleset miss"
                  :protocol="form.activeProtocol"
                  :protocol-spec="activeProtocolSpec"
                />
                <FallbackEditor
                  v-model="activePolicy.ruleMiss"
                  title="Rule miss"
                  :protocol="form.activeProtocol"
                  :protocol-spec="activeProtocolSpec"
                />
              </div>
            </div>
          </section>

          <aside class="policy-preview">
            <header class="section-heading">
              <strong>Preview</strong>
              <span>{{ form.activeProtocol.toUpperCase() }} fallback result</span>
            </header>
            <div class="preview-row">
              <small>ruleset miss</small>
              <FallbackSummary :action="previewActions.rulesetMiss" />
            </div>
            <div class="preview-row">
              <small>rule miss</small>
              <FallbackSummary :action="previewActions.ruleMiss" />
            </div>
            <div class="linked-preview">
              <small>linked rulesets</small>
              <button
                v-for="ruleSet in activePolicyUsage.rulesets"
                :key="ruleSet.id"
                type="button"
                @click="goRuleset(ruleSet.id)"
              >
                {{ ruleSet.name }}
              </button>
              <code v-if="!activePolicyUsage.rulesets.length">none</code>
            </div>
          </aside>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="store.saving" @click="submit">Save namespace</el-button>
      </template>
    </el-dialog>
  </PageContainer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
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
  NamespacePolicy,
} from '@/types'
import {
  buildResponsePayload,
  defaultResponsePayload,
  responseFieldDraftsFromPayload,
} from '@/utils/responseSpec'
import {
  buildNamespaceEntryMetrics,
  buildNamespaceEntryRows,
  defaultNamespaceEntryFilters,
  type EntryTone,
  type NamespaceEntryRow,
  type NamespaceFallbackFilter,
  type NamespacePolicySummary,
  type NamespaceProtocolFilter,
  type NamespaceSortKey,
  type NamespaceUsageFilter,
} from '@/utils/entryLists'

interface NamespaceForm {
  id: string
  name: string
  description: string
  activeProtocol: string
  policies: Record<string, NamespacePolicyForm>
}

interface NamespacePolicyForm {
  rulesetMiss: NamespaceFallbackForm
  ruleMiss: NamespaceFallbackForm
}

const router = useRouter()
const store = useMockserverStore()

const fallbackOptions: Array<{ label: string; value: NamespaceFallbackFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Forward', value: 'forward' },
  { label: 'Response', value: 'respond' },
]
const usageOptions: Array<{ label: string; value: NamespaceUsageFilter }> = [
  { label: 'All', value: 'all' },
  { label: 'Used', value: 'used' },
  { label: 'Unused', value: 'unused' },
]
const sortOptions: Array<{ label: string; value: NamespaceSortKey }> = [
  { label: 'ID', value: 'id' },
  { label: 'Name', value: 'name' },
  { label: 'Protocols', value: 'protocols' },
  { label: 'Usage', value: 'usage' },
]
const filters = reactive(defaultNamespaceEntryFilters())
const dialogVisible = ref(false)
const editingNamespace = ref<NamespaceConfig | null>(null)

const form = reactive<NamespaceForm>({
  id: '',
  name: '',
  description: '',
  activeProtocol: 'http',
  policies: {
    http: createPolicyForm('http'),
  },
})

const allRows = computed(() => buildNamespaceEntryRows(store.namespaces, store.drafts, defaultNamespaceEntryFilters()))
const filteredRows = computed(() => buildNamespaceEntryRows(store.namespaces, store.drafts, filters))
const metrics = computed(() => buildNamespaceEntryMetrics(allRows.value, filteredRows.value))
const summaryCards = computed<Array<{ label: string; value: number; tone: EntryTone }>>(() => [
  { label: 'Namespaces', value: metrics.value.total, tone: 'neutral' },
  { label: 'Shown', value: metrics.value.shown, tone: 'neutral' },
  { label: 'Used', value: metrics.value.used, tone: 'ok' },
  { label: 'Unused', value: metrics.value.unused, tone: 'neutral' },
  { label: 'Forward fallback', value: metrics.value.forward, tone: 'warn' },
  { label: 'Response fallback', value: metrics.value.response, tone: 'accent' },
])
const protocolOptions = computed(() => {
  const names = new Set(store.protocols.map((protocol) => protocol.name.toLowerCase()))
  Object.keys(form.policies).forEach((protocol) => names.add(protocol.toLowerCase()))
  if (!names.size) names.add('http')
  return Array.from(names)
    .filter(Boolean)
    .sort()
    .map((protocol) => ({
      label: protocol.toUpperCase(),
      value: protocol,
    }))
})
const protocolFilterOptions = computed<Array<{ label: string; value: NamespaceProtocolFilter }>>(() => [
  { label: 'All', value: 'all' },
  ...Array.from(new Set(allRows.value.flatMap((row) => row.protocols)))
    .sort()
    .map((protocol) => ({
      label: protocol.toUpperCase(),
      value: protocol,
    })),
])
const activePolicy = computed(() => ensurePolicyForm(form.activeProtocol))
const activeProtocolSpec = computed(() => protocolSpecFor(activeProtocolKey()))
const activePolicyUsage = computed(() => {
  const namespaceID = editingNamespace.value?.id || form.id
  const protocol = activeProtocolKey()
  const related = store.drafts.filter(
    (ruleSet) => ruleSet.namespace === namespaceID && ruleSet.protocol.toLowerCase() === protocol
  )
  return {
    rulesetCount: related.length,
    ruleCount: related.reduce((count, ruleSet) => count + ruleSet.rules.length, 0),
    rulesets: related.map((ruleSet) => ({
      id: ruleSet.id,
      name: ruleSet.name || ruleSet.id,
      protocol: ruleSet.protocol.toLowerCase(),
    })),
  }
})
const validationIssues = computed(() => validateForm())
const previewActions = computed(() => ({
  rulesetMiss: fallbackFormToAction(activePolicy.value.rulesetMiss, activeProtocolSpec.value),
  ruleMiss: fallbackFormToAction(activePolicy.value.ruleMiss, activeProtocolSpec.value),
}))

async function loadAll() {
  await Promise.all([store.fetchNamespaces(), store.fetchDrafts(), store.fetchProtocols()])
}

function openCreateDialog() {
  editingNamespace.value = null
  resetForm(null)
  dialogVisible.value = true
}

function openEditDialog(namespace: NamespaceConfig, protocol?: string) {
  editingNamespace.value = namespace
  resetForm(namespace, protocol)
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
    policies: policiesFromForm(),
  }
  await store.saveNamespace(payload)
  dialogVisible.value = false
  await loadAll()
}

function resetForm(namespace: NamespaceConfig | null, protocol?: string) {
  form.id = namespace?.id || ''
  form.name = namespace?.name || ''
  form.description = namespace?.description || ''
  form.policies = namespace ? policyFormsFromNamespace(namespace) : { http: createPolicyForm('http') }
  form.activeProtocol = protocol || preferredProtocol(namespace)
}

function visiblePolicies(row: NamespaceEntryRow): NamespacePolicySummary[] {
  return row.policies.slice(0, 3)
}

function hiddenPolicyCount(row: NamespaceEntryRow) {
  return Math.max(row.policies.length - 3, 0)
}

function visibleUsage(row: NamespaceEntryRow) {
  return row.usage.rulesets.slice(0, 3)
}

function hiddenUsage(row: NamespaceEntryRow) {
  return row.usage.rulesets.slice(3)
}

function hiddenUsageCount(row: NamespaceEntryRow) {
  return Math.max(row.usage.rulesets.length - 3, 0)
}

function goRuleset(rulesetId: string) {
  router.push(`/rulesets/${encodeURIComponent(rulesetId)}/rules`)
}

function onPolicyTabChange() {
  ensurePolicyForm(form.activeProtocol)
}

function validateForm() {
  const issues: string[] = []
  if (editingNamespace.value && !form.id.trim()) {
    issues.push('Namespace ID is required')
  }
  if (editingNamespace.value && !/^[a-zA-Z0-9_-]+$/.test(form.id.trim())) {
    issues.push('Namespace ID can only contain letters, numbers, underscores, and hyphens')
  }
  for (const [protocol, policy] of Object.entries(form.policies)) {
    validateFallbackForm(policy.rulesetMiss, `${protocol} ruleset miss`, issues, protocolSpecFor(protocol))
    validateFallbackForm(policy.ruleMiss, `${protocol} rule miss`, issues, protocolSpecFor(protocol))
  }
  return issues
}

function validateFallbackForm(
  fallback: NamespaceFallbackForm,
  label: string,
  issues: string[],
  protocolSpec = activeProtocolSpec.value
) {
  if (fallback.type === 'respond') {
    const result = buildResponsePayload(protocolSpec, fallback.responsePayload, fallback.responseFieldDrafts)
    result.issues.forEach((issue) => issues.push(`${label} ${issue.message}`))
    return
  }
  if (fallback.forwardTimeoutMs < 0 || fallback.forwardTimeoutMs > 30000) {
    issues.push(`${label} forward timeout must be between 0 and 30000`)
  }
}

function fallbackFormToAction(
  fallback: NamespaceFallbackForm,
  protocolSpec = activeProtocolSpec.value
): NamespaceFallbackAction {
  if (fallback.type === 'forward') {
    return {
      type: 'forward',
      forward: {
        timeout_ms: fallback.forwardTimeoutMs || undefined,
      },
    }
  }
  const response = buildResponsePayload(protocolSpec, fallback.responsePayload, fallback.responseFieldDrafts)
  return {
    type: 'respond',
    renderer: 'static',
    response: {
      protocol: protocolSpec?.name,
      payload: response.payload,
    },
  }
}

function actionToFallbackForm(action: NamespaceFallbackAction | undefined, protocol = 'http'): NamespaceFallbackForm {
  const fallback = createFallbackForm(protocol)
  if (!action) return fallback
  fallback.type = action.type
  if (action.type === 'forward') {
    fallback.forwardTimeoutMs = action.forward?.timeout_ms || 0
    return fallback
  }
  fallback.responsePayload = action.response?.payload ?? defaultResponsePayload(protocolSpecFor(protocol))
  fallback.responseFieldDrafts = responseFieldDraftsFromPayload(protocolSpecFor(protocol), fallback.responsePayload)
  return fallback
}

function createFallbackForm(protocol = 'http'): NamespaceFallbackForm {
  const payload = defaultResponsePayload(protocolSpecFor(protocol))
  return {
    type: 'forward',
    responsePayload: payload,
    responseFieldDrafts: responseFieldDraftsFromPayload(protocolSpecFor(protocol), payload),
    forwardTimeoutMs: 5000,
  }
}

function createPolicyForm(protocol = 'http'): NamespacePolicyForm {
  return {
    rulesetMiss: createFallbackForm(protocol),
    ruleMiss: createFallbackForm(protocol),
  }
}

function ensurePolicyForm(protocol: string): NamespacePolicyForm {
  const key = protocol.trim().toLowerCase() || 'http'
  if (!form.policies[key]) {
    form.policies[key] = createPolicyForm(key)
  }
  return form.policies[key]
}

function policiesFromForm(): Record<string, NamespacePolicy> {
  ensurePolicyForm(form.activeProtocol)
  return Object.fromEntries(
    Object.entries(form.policies)
      .filter(([protocol]) => protocol.trim())
      .map(([protocol, policy]) => [
        protocol.trim().toLowerCase(),
        {
          ruleset_miss_action: fallbackFormToAction(policy.rulesetMiss, protocolSpecFor(protocol)),
          rule_miss_action: fallbackFormToAction(policy.ruleMiss, protocolSpecFor(protocol)),
        },
      ])
  )
}

function policyFormsFromNamespace(namespace: NamespaceConfig): Record<string, NamespacePolicyForm> {
  const entries = Object.entries(namespace.policies || {})
  if (!entries.length) return { http: createPolicyForm('http') }
  return Object.fromEntries(
    entries.map(([protocol, policy]) => [
      protocol.toLowerCase(),
      {
        rulesetMiss: actionToFallbackForm(policy.ruleset_miss_action, protocol),
        ruleMiss: actionToFallbackForm(policy.rule_miss_action, protocol),
      },
    ])
  )
}

function preferredProtocol(namespace: NamespaceConfig | null) {
  if (!namespace) return 'http'
  return namespace.policies?.http ? 'http' : Object.keys(namespace.policies || {})[0]?.toLowerCase() || 'http'
}

function activeProtocolKey() {
  return form.activeProtocol.trim().toLowerCase() || 'http'
}

function protocolSpecFor(protocol: string) {
  return store.protocolMap.get(protocol.trim().toLowerCase())
}

onMounted(loadAll)

watch(
  () => form.activeProtocol,
  () => {
    form.activeProtocol = activeProtocolKey()
    ensurePolicyForm(form.activeProtocol)
  },
  { immediate: true }
)
</script>

<style lang="scss" scoped>
.namespace-page {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  gap: var(--ms-space-3);
}

.namespace-toolbar {
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
  min-width: 1220px;
  display: grid;
  grid-template-columns:
    minmax(210px, 1.12fr)
    minmax(150px, 0.82fr)
    minmax(128px, 0.7fr)
    minmax(230px, 1.08fr)
    minmax(230px, 1.08fr)
    minmax(260px, 1.16fr)
    104px;
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
  min-height: 118px;
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

.namespace-description {
  display: -webkit-box;
  overflow: hidden;
  max-width: 100%;
  margin: 0;
  color: var(--ms-text-tertiary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.protocol-cell,
.linked-rulesets {
  flex-wrap: wrap;
  gap: var(--ms-space-1);
}

.protocol-chip,
.policy-protocol {
  display: inline-flex;
  align-items: center;
  padding: 4px 9px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-blue-500);
  background: var(--ms-blue-50);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  line-height: 1.35;
  text-transform: uppercase;
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

.policy-cell {
  min-width: 0;
}

.policy-stack {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.policy-summary {
  min-width: 0;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--ms-space-2);
  align-items: center;
}

.policy-summary :deep(.fallback-summary) {
  max-width: 100%;
}

.policy-protocol {
  padding: 3px 7px;
  font-size: var(--ms-text-xs);
}

.more-policies {
  width: fit-content;
  height: 28px;
  padding: 0 9px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-secondary);
  background: var(--ms-control-bg);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  cursor: pointer;

  &:hover {
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);
  }
}

.linked-rulesets button {
  max-width: 148px;
  min-height: 30px;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 0 9px;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-teal-700);
  background: var(--ms-teal-50);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  cursor: pointer;

  &:hover {
    border-color: rgba(13, 148, 136, 0.32);
    background: var(--ms-teal-100);
  }

  span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    flex: 0 0 auto;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-xs);
    text-transform: uppercase;
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
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 2px var(--ms-space-2);
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

  span {
    color: var(--ms-blue-500);
    font-size: var(--ms-text-xs);
    font-weight: var(--ms-font-semibold);
  }

  code {
    grid-column: 1 / -1;
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
  grid-template-columns: minmax(0, 1fr) minmax(300px, 340px);
  grid-template-areas:
    "profile preview"
    "policy preview";
  gap: var(--ms-space-4);
}

.profile-section,
.policy-section,
.policy-preview {
  min-width: 0;
  padding: var(--ms-space-4);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);
}

.profile-section {
  grid-area: profile;
}

.policy-section {
  grid-area: policy;
}

.policy-preview {
  grid-area: preview;
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.section-heading {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: var(--ms-space-3);

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    font-weight: var(--ms-font-bold);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
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

.policy-tabs {
  margin-bottom: var(--ms-space-3);
}

.policy-workspace {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.policy-context {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-2);
  padding: var(--ms-space-2) var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-panel-bg);

  > div {
    display: flex;
    align-items: baseline;
    gap: 4px;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }

  strong {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-base);
  }
}

.preview-row,
.linked-preview {
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

.linked-preview button {
  width: 100%;
  min-width: 0;
  min-height: 30px;
  padding: 0 var(--ms-space-2);
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: var(--ms-radius-md);
  color: var(--ms-teal-700);
  background: var(--ms-teal-50);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;

  &:hover {
    border-color: rgba(13, 148, 136, 0.32);
  }
}

@media (max-width: 980px) {
  .namespace-toolbar,
  .dialog-grid,
  .form-grid,
  .fallback-grid {
    grid-template-columns: 1fr;
  }

  .dialog-grid {
    grid-template-areas:
      "profile"
      "policy"
      "preview";
  }

  .filter-block {
    align-items: flex-start;
    flex-direction: column;
  }

  .filter-group {
    width: 100%;
    overflow-x: auto;
  }

  .filter-group button {
    flex: 1 0 auto;
  }
}
</style>
