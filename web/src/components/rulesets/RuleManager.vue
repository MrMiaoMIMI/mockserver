<template>
  <section class="rule-manager">
    <header class="manager-header">
      <div>
        <span>rule manager</span>
        <strong>{{ visibleRows.length }} / {{ allRows.length }} rules</strong>
        <small v-if="activeRule" :title="ruleTechnicalLabel(activeRule)">
          selected {{ ruleDisplayName(activeRule) }}
        </small>
      </div>
      <el-button type="primary" :icon="Plus" :disabled="!ruleSet" @click="$emit('create')">
        Add rule
      </el-button>
    </header>

    <el-empty v-if="!ruleSet" description="Select a draft ruleset first" />
    <el-empty v-else-if="!ruleSet.rules.length" description="No rules yet" />
    <template v-else>
      <div class="manager-toolbar">
        <div class="toolbar-main">
          <el-input
            v-model="filters.query"
            :prefix-icon="Search"
            placeholder="Search rule id, name, condition, action, status"
            clearable
          />

          <el-popover
            placement="bottom-end"
            trigger="click"
            :width="360"
            popper-class="rule-filter-popover"
          >
            <template #reference>
              <el-button :icon="Filter" :type="hasActiveFilters ? 'primary' : 'default'">
                Filters
                <span v-if="activeFilterChips.length" class="filter-count">
                  {{ activeFilterChips.length }}
                </span>
              </el-button>
            </template>
            <div class="filter-popover">
              <header>
                <strong>Filter rules</strong>
                <button type="button" :disabled="!hasActiveFilters" @click="resetFilters">
                  Reset
                </button>
              </header>
              <label>
                <span>State</span>
                <FilterSegment
                  v-model="filters.enabled"
                  :options="enabledOptions"
                  aria-label="Enabled filter"
                />
              </label>
              <label>
                <span>Action type</span>
                <el-select v-model="filters.actionType" placeholder="Any action" clearable>
                  <el-option
                    v-for="actionType in actionTypes"
                    :key="actionType"
                    :label="actionType"
                    :value="actionType"
                  />
                </el-select>
              </label>
              <label>
                <span>Diagnostics</span>
                <FilterSegment
                  v-model="filters.status"
                  :options="statusOptions"
                  aria-label="Diagnostic status filter"
                />
              </label>
            </div>
          </el-popover>
        </div>

        <div class="filter-summary" aria-label="Rule status counts">
          <div v-if="activeFilterChips.length" class="active-filter-chips">
            <StateChip
              v-for="chip in activeFilterChips"
              :key="chip"
              :label="chip"
              tone="accent"
            />
            <button type="button" @click="resetFilters">clear</button>
          </div>
          <small>{{ countSummary }}</small>
        </div>
      </div>

      <el-empty
        v-if="!visibleRows.length"
        class="filtered-empty"
        description="No rules match the current filters"
      />

      <div v-else class="rule-stack" role="listbox" aria-label="Rule stack">
      <article
        v-for="row in visibleRows"
        :key="row.rule.id"
        class="rule-card"
        :class="{ active: row.selected }"
        role="option"
        tabindex="0"
        :aria-selected="row.selected"
        @click="setActiveRule(row.rule.id)"
        @keydown.enter.prevent="setActiveRule(row.rule.id)"
        @keydown.space.prevent="setActiveRule(row.rule.id)"
      >
        <div class="rule-content">
          <div class="rule-heading">
            <div class="rule-title">
              <span class="status-dot" :class="{ 'is-on': row.rule.enabled }" />
              <div>
                <strong :title="ruleDisplayName(row.rule)">{{ ruleDisplayName(row.rule) }}</strong>
                <small :title="row.rule.id">{{ ruleTechnicalLabel(row.rule) }}</small>
              </div>
            </div>
            <div class="rule-badges">
              <StateChip :label="`p${row.rule.priority}`" tone="neutral" />
              <StateChip
                v-for="status in row.statuses"
                :key="status.key"
                :label="status.label"
                :tone="status.tone"
                :title="status.title"
              />
            </div>
          </div>
          <div class="rule-action-summary">
            <span>action</span>
            <code :title="row.action">{{ row.action }}</code>
          </div>
          <details class="rule-condition">
            <summary>Condition</summary>
            <code :title="row.condition">{{ row.condition }}</code>
          </details>
        </div>
        <div class="rule-actions">
          <button
            class="icon-button"
            type="button"
            :title="`Edit ${row.rule.id}`"
            :aria-label="`Edit ${row.rule.id}`"
            :disabled="isBusy(row.rule)"
            @click.stop="requestEditRule(row.rule)"
          >
            <el-icon><EditPen /></el-icon>
          </button>
          <el-dropdown trigger="click" @command="(command: string) => handleRuleCommand(command, row.rule)">
            <button
              class="icon-button"
              type="button"
              :title="`More actions for ${row.rule.id}`"
              :aria-label="`More actions for ${row.rule.id}`"
              :disabled="isBusy(row.rule)"
              @click.stop
            >
              <el-icon><MoreFilled /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="move-up" :disabled="!canMoveRule(row.rule, -1)">
                  Move up
                </el-dropdown-item>
                <el-dropdown-item command="move-down" :disabled="!canMoveRule(row.rule, 1)">
                  Move down
                </el-dropdown-item>
                <el-dropdown-item command="duplicate">Duplicate</el-dropdown-item>
                <el-dropdown-item divided command="delete">Delete</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </article>
    </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  EditPen,
  Filter,
  MoreFilled,
  Plus,
  Search,
} from '@element-plus/icons-vue'
import FilterSegment, { type FilterSegmentOption } from '@/components/common/FilterSegment.vue'
import StateChip from '@/components/common/StateChip.vue'
import { useMockserverStore } from '@/store'
import type { Rule, RuleSet } from '@/types'
import {
  availableActionTypes,
  buildRuleRows,
  buildRuleStatusCounts,
  defaultRuleFilters,
  orderedRules,
  resolveStableSelectedRuleId,
  type RuleDiagnosticState,
  type RuleEnabledFilter,
  type RuleStatusFilter,
} from '@/utils/ruleCollection'
import {
  cloneRule,
  nextDuplicatePriority,
  ruleDisplayName,
  ruleTechnicalLabel,
} from '@/utils/ruleSummaries'
import { generateRuleId } from '@/utils/ruleFormAdapter'

const props = defineProps<{
  ruleSet: RuleSet | null
  activeRuleId?: string
  diagnostics?: Record<string, RuleDiagnosticState>
}>()

const emit = defineEmits<{
  updated: [ruleSet: RuleSet]
  select: [ruleId: string]
  create: []
  edit: [rule: Rule]
}>()

const store = useMockserverStore()
const localActiveRuleId = ref('')
const busyRuleKey = ref('')
const filters = reactive(defaultRuleFilters())

const selectedRuleId = computed(() => props.activeRuleId || localActiveRuleId.value)
const allRows = computed(() => {
  return buildRuleRows(props.ruleSet, defaultRuleFilters(), selectedRuleId.value, props.diagnostics)
})
const visibleRows = computed(() => {
  return buildRuleRows(props.ruleSet, filters, selectedRuleId.value, props.diagnostics)
})
const counts = computed(() => buildRuleStatusCounts(allRows.value))
const actionTypes = computed(() => availableActionTypes(props.ruleSet))
const activeRule = computed(() => {
  return orderedRules(props.ruleSet?.rules || []).find((rule) => rule.id === selectedRuleId.value) || null
})
const activeFilterChips = computed(() => {
  const chips: string[] = []
  const query = filters.query.trim()
  if (query) chips.push(`search "${query}"`)
  if (filters.enabled !== 'all') chips.push(filters.enabled)
  if (filters.actionType) chips.push(filters.actionType)
  if (filters.status !== 'all') {
    chips.push(statusOptions.find((option) => option.value === filters.status)?.label || filters.status)
  }
  return chips
})
const hasActiveFilters = computed(() => activeFilterChips.value.length > 0)
const countSummary = computed(() => {
  return `${visibleRows.value.length}/${allRows.value.length} shown · ${counts.value.enabled} enabled · ${counts.value.invalid} invalid · ${counts.value.matched} matched`
})

const enabledOptions: Array<FilterSegmentOption & { value: RuleEnabledFilter }> = [
  { label: 'all', value: 'all' },
  { label: 'enabled', value: 'enabled' },
  { label: 'disabled', value: 'disabled' },
]

const statusOptions: Array<FilterSegmentOption & { value: RuleStatusFilter }> = [
  { label: 'all', value: 'all' },
  { label: 'matched', value: 'matched' },
  { label: 'missed', value: 'missed' },
  { label: 'not reached', value: 'not_reached' },
  { label: 'fallback', value: 'fallback' },
  { label: 'invalid', value: 'invalid' },
]

watch(
  () => props.ruleSet?.rules.map((rule) => rule.id).join('\u0000') || '',
  () => {
    const nextRuleId = resolveStableSelectedRuleId(props.ruleSet, selectedRuleId.value)
    if (!nextRuleId) {
      localActiveRuleId.value = ''
      emit('select', '')
      return
    }
    if (nextRuleId !== selectedRuleId.value) {
      setActiveRule(nextRuleId)
    }
  },
  { immediate: true }
)

function setActiveRule(ruleId: string) {
  localActiveRuleId.value = ruleId
  emit('select', ruleId)
}

function requestEditRule(rule: Rule) {
  setActiveRule(rule.id)
  emit('edit', rule)
}

function handleRuleCommand(command: string, rule: Rule) {
  if (command === 'move-up') {
    void moveRule(rule, -1)
    return
  }
  if (command === 'move-down') {
    void moveRule(rule, 1)
    return
  }
  if (command === 'duplicate') {
    void duplicateRule(rule)
    return
  }
  if (command === 'delete') {
    void deleteRule(rule)
  }
}

async function moveRule(rule: Rule, direction: -1 | 1) {
  if (!props.ruleSet || !canMoveRule(rule, direction)) return
  setActiveRule(rule.id)
  await withRuleOperation(rule, direction === -1 ? 'move-up' : 'move-down', async () => {
    const rules = orderedRules(props.ruleSet!.rules)
    const index = rules.findIndex((item) => item.id === rule.id)
    const target = rules[index + direction]
    if (!target) return
    const nextRuleSet = cloneRuleSet(props.ruleSet!)
    const currentRule = nextRuleSet.rules.find((item) => item.id === rule.id)
    const targetRule = nextRuleSet.rules.find((item) => item.id === target.id)
    if (!currentRule || !targetRule) return
    const nextPriority = targetRule.priority
    targetRule.priority = currentRule.priority
    currentRule.priority = nextPriority
    const updated = await store.saveDraft(nextRuleSet)
    emit('updated', updated)
  })
}

async function duplicateRule(rule: Rule) {
  if (!props.ruleSet) return
  setActiveRule(rule.id)
  const duplicated = cloneRule(rule)
  duplicated.name = `${ruleDisplayName(rule)} copy`
  duplicated.id = generateRuleId(props.ruleSet)
  duplicated.priority = nextDuplicatePriority(rule.priority, props.ruleSet.rules)
  await withRuleOperation(rule, 'duplicate', async () => {
    const updated = await store.addRule(props.ruleSet!.id, duplicated)
    emit('updated', updated)
    setActiveRule(duplicated.id)
  })
}

async function deleteRule(rule: Rule) {
  if (!props.ruleSet) return
  setActiveRule(rule.id)
  await withRuleOperation(rule, 'delete', async () => {
    await ElMessageBox.confirm(`Delete rule ${rule.name || rule.id} (${rule.id})?`, 'Delete rule', {
      type: 'warning',
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
    })
    const updated = await store.deleteRule(props.ruleSet!.id, rule.id)
    emit('updated', updated)
  })
}

function canMoveRule(rule: Rule, direction: -1 | 1) {
  const rules = orderedRules(props.ruleSet?.rules || [])
  const index = rules.findIndex((item) => item.id === rule.id)
  return index >= 0 && index + direction >= 0 && index + direction < rules.length
}

function isBusy(rule: Rule) {
  return busyRuleKey.value.startsWith(`${rule.id}:`)
}

function resetFilters() {
  Object.assign(filters, defaultRuleFilters())
}

async function withRuleOperation(rule: Rule, action: string, operation: () => Promise<void>) {
  busyRuleKey.value = `${rule.id}:${action}`
  try {
    await operation()
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(toErrorMessage(error))
    }
  } finally {
    busyRuleKey.value = ''
  }
}

function cloneRuleSet(ruleSet: RuleSet): RuleSet {
  return JSON.parse(JSON.stringify(ruleSet)) as RuleSet
}

function isDialogCancel(error: unknown) {
  return error === 'cancel' || error === 'close' || toErrorMessage(error).toLowerCase() === 'cancel'
}

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}
</script>

<style lang="scss" scoped>
.rule-manager {
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-xl);
  background: var(--ms-panel-bg);
  box-shadow: var(--ms-shadow-xs), var(--ms-shadow-inset);
}

.manager-header {
  flex-shrink: 0;
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-4);
  padding: 0 var(--ms-space-4);
  border-bottom: 1px solid var(--ms-border-light);

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

  strong {
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-lg);
    font-weight: var(--ms-font-bold);
  }

  small {
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.rule-stack {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  overflow: auto;
  padding: var(--ms-space-4);
}

.manager-toolbar {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
  padding: var(--ms-space-2) var(--ms-space-3);
  border-bottom: 1px solid var(--ms-border-light);
}

.toolbar-main {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--ms-space-2);
  align-items: center;
}

.filter-count {
  min-width: 18px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: 4px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-inverse);
  background: rgba(255, 255, 255, 0.24);
  font-family: var(--ms-font-mono);
  font-size: 11px;
  font-weight: var(--ms-font-bold);
}

.filter-summary {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-2);

  small {
    min-width: 0;
    overflow: hidden;
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
    line-height: 1.4;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.active-filter-chips {
  min-width: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--ms-space-1);

  button {
    border: none;
    color: var(--ms-teal-700);
    background: transparent;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    cursor: pointer;
  }
}

.filtered-empty {
  min-height: 220px;
}

:global(.rule-filter-popover) {
  padding: var(--ms-space-3);
}

:global(.rule-filter-popover .filter-popover) {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

:global(.rule-filter-popover header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-2);
}

:global(.rule-filter-popover header strong) {
  color: var(--ms-text-primary);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-md);
}

:global(.rule-filter-popover header button) {
  border: none;
  color: var(--ms-teal-700);
  background: transparent;
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
  cursor: pointer;
}

:global(.rule-filter-popover header button:disabled) {
  color: var(--ms-text-muted);
  cursor: not-allowed;
}

:global(.rule-filter-popover label) {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

:global(.rule-filter-popover label > span) {
  color: var(--ms-text-tertiary);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.rule-card {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg-soft);
  cursor: pointer;
  transition:
    background var(--ms-transition-fast),
    border-color var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast);

  &:hover,
  &:focus-visible {
    border-color: rgba(37, 99, 235, 0.32);
    background: var(--ms-panel-bg-hover);
    outline: none;
  }

  &.active {
    border-color: rgba(37, 99, 235, 0.48);
    background: var(--ms-teal-50);
    box-shadow: inset 3px 0 0 var(--ms-teal-600), var(--ms-shadow-xs);
  }
}

.rule-content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.rule-heading {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ms-space-3);
}

.rule-title {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: var(--ms-space-3);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  strong,
  small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-base);
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.rule-badges {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--ms-space-1);
  flex: 0 0 auto;
}

.rule-action-summary {
  min-width: 0;
  display: grid;
  grid-template-columns: 62px minmax(0, 1fr);
  gap: var(--ms-space-2);
  align-items: center;

  span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.rule-condition {
  min-width: 0;

  summary {
    width: max-content;
    max-width: 100%;
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    cursor: pointer;
  }

  code {
    display: block;
    min-width: 0;
    margin-top: var(--ms-space-1);
    padding: var(--ms-space-2);
    overflow: hidden;
    border: 1px solid var(--ms-border-light);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: var(--ms-control-bg);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.rule-actions {
  display: flex;
  align-items: center;
  align-self: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--ms-space-2);
  max-width: 84px;
}

.icon-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

@media (max-width: 980px) {
  .rule-card,
  .rule-action-summary {
    grid-template-columns: 1fr;
  }

  .rule-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
    max-width: none;
  }

  .rule-badges {
    justify-content: flex-start;
  }
}

@media (max-width: 620px) {
  .manager-header {
    height: auto;
    min-height: 0;
    align-items: stretch;
    flex-direction: column;
    gap: var(--ms-space-2);
    margin-bottom: var(--ms-space-2);
    padding: var(--ms-space-3);
    background: var(--ms-panel-bg);

    :deep(.el-button) {
      width: 100%;
      margin-left: 0;
    }
  }

  .rule-stack {
    padding: var(--ms-space-3);
  }

  .filter-row {
    grid-template-columns: 1fr;
  }

  .status-control {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .rule-card {
    padding: var(--ms-space-3);
  }

  .rule-heading {
    flex-direction: column;
  }

  .rule-title {
    width: 100%;
  }
}

@media (max-width: 1500px) and (min-width: 981px) {
  .rule-actions {
    justify-content: flex-start;
    max-width: none;
  }
}
</style>
