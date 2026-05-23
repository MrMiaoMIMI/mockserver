<template>
  <div class="result-inspector">
    <el-empty v-if="!parsed" description="No result yet" />

    <template v-else>
      <section v-if="validation" class="result-panel">
        <header>
          <span>validation</span>
          <strong :class="{ ok: validation.valid }">{{
            validation.valid ? 'pass' : 'blocked'
          }}</strong>
        </header>
        <div v-if="validation.issues?.length" class="issue-list">
          <div
            v-for="issue in validation.issues"
            :key="`${issue.path}-${issue.message}`"
            class="issue-row"
          >
            <code :title="issue.path">{{ issue.path }}</code>
            <span :title="issue.message">{{ issue.message }}</span>
          </div>
        </div>
        <div v-if="validationWarnings.length" class="issue-list">
          <div
            v-for="warning in validationWarnings"
            :key="`${warning.path}-${warning.message}`"
            class="issue-row is-warning"
          >
            <code :title="warning.path">{{ warning.path }}</code>
            <span :title="warning.message">{{ warning.message }}</span>
          </div>
        </div>
      </section>

      <section v-if="simulationDiagnostics" class="result-panel diagnostic-panel">
        <header>
          <span>simulation</span>
          <strong :class="{ ok: simulationDiagnostics.matched }">
            {{ simulationDiagnostics.statusLabel }}
          </strong>
        </header>
        <div class="summary-grid">
          <div
            v-for="metric in simulationDiagnostics.metrics"
            :key="metric.label"
            :class="['metric-card', metric.tone ? `is-${metric.tone}` : '']"
          >
            <span>{{ metric.label }}</span>
            <code>{{ metric.value }}</code>
          </div>
        </div>

        <div v-if="simulationDiagnostics.candidateRules.length" class="candidate-list">
          <span>candidate rules</span>
          <code v-for="ruleId in simulationDiagnostics.candidateRules" :key="ruleId">
            {{ ruleId }}
          </code>
        </div>

        <section v-if="diagnosisPath.length" class="trace-section">
          <div class="trace-heading">diagnosis path</div>
          <div class="diagnosis-path">
            <div
              v-for="step in diagnosisPath"
              :key="step.id"
              :class="['diagnosis-step', { ok: step.matched }]"
            >
              <span>{{ step.matched ? 'pass' : 'miss' }}</span>
              <code>{{ step.label }}</code>
            </div>
          </div>
        </section>

        <section v-if="simulationDiagnostics.rulesetTraces.length" class="trace-section">
          <div class="trace-heading">ruleset outcome</div>
          <div
            v-for="trace in simulationDiagnostics.rulesetTraces"
            :key="trace.id"
            class="diagnostic-row"
            :class="{ ok: trace.matched }"
          >
            <span>{{ trace.matched ? 'pass' : 'miss' }}</span>
            <div>
              <code>{{ trace.label }}</code>
              <small>{{ trace.message }}</small>
              <small v-if="trace.detail">{{ trace.detail }}</small>
            </div>
          </div>
        </section>

        <section v-if="simulationDiagnostics.conditionTraces.length" class="trace-section">
          <div class="trace-heading">rule condition checks</div>
          <div
            v-for="trace in simulationDiagnostics.conditionTraces"
            :key="trace.id"
            class="diagnostic-row"
            :class="{ ok: trace.matched }"
            :style="{ '--trace-depth': trace.depth || 0 }"
          >
            <span>{{ trace.matched ? 'pass' : 'miss' }}</span>
            <div>
              <code>{{ trace.label }}</code>
              <small>{{ trace.message }}</small>
              <small v-if="trace.detail">{{ trace.detail }}</small>
            </div>
          </div>
        </section>

        <section v-if="simulationDiagnostics.selectorTraces.length" class="trace-section">
          <div class="trace-heading">selector checks</div>
          <div
            v-for="trace in simulationDiagnostics.selectorTraces"
            :key="trace.id"
            class="diagnostic-row"
            :class="{ ok: trace.matched }"
          >
            <span>{{ trace.matched ? 'pass' : 'miss' }}</span>
            <div>
              <code>{{ trace.label }}</code>
              <small>{{ trace.message }}</small>
              <small v-if="trace.detail">{{ trace.detail }}</small>
            </div>
          </div>
        </section>
      </section>

      <section v-if="diff" class="result-panel">
        <header>
          <span>rollback diff</span>
          <strong :class="{ ok: !diff.changed }">{{ diff.changed ? 'changed' : 'same' }}</strong>
        </header>
        <div v-if="fieldDiffs.length" class="issue-list">
          <div v-for="item in fieldDiffs" :key="`${item.path}-${item.message}`" class="issue-row">
            <code :title="item.path">{{ item.path }}</code>
            <span :title="item.message">{{ item.message }}</span>
          </div>
        </div>
      </section>

      <JsonEditor
        v-if="showRaw"
        :model-value="rawJson"
        readonly
        class="raw-result"
        title="Raw JSON"
        :show-format="false"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import JsonEditor from '@/components/common/JsonEditor.vue'
import { buildSimulationDiagnostics } from '@/utils/simulationDiagnostics'

const props = withDefaults(
  defineProps<{
    rawJson: string
    showRaw?: boolean
  }>(),
  {
    showRaw: true,
  }
)

const parsed = computed<Record<string, any> | null>(() => {
  if (!props.rawJson) return null
  try {
    return JSON.parse(props.rawJson) as Record<string, any>
  } catch {
    return null
  }
})

const payload = computed(() => parsed.value?.result || parsed.value)
const isValidationPayload = (
  value: unknown
): value is { valid?: boolean; issues?: any[]; warnings?: any[] } => {
  return Boolean(
    value &&
    typeof value === 'object' &&
    ('valid' in value || Array.isArray((value as { issues?: unknown[] }).issues))
  )
}

const simulation = computed(() => {
  if (payload.value?.simulation) return payload.value.simulation
  if (payload.value?.trace && typeof payload.value?.matched === 'boolean') return payload.value
  return null
})
const validation = computed(() => {
  if (isValidationPayload(payload.value?.validation)) {
    return payload.value.validation
  }
  if (simulation.value || payload.value?.diff) {
    return null
  }
  return isValidationPayload(payload.value) ? payload.value : null
})
const validationWarnings = computed(() => {
  return Array.isArray(validation.value?.warnings) ? validation.value.warnings : []
})
const simulationDiagnostics = computed(() => buildSimulationDiagnostics(simulation.value))
const diagnosisPath = computed(() => {
  if (!simulationDiagnostics.value) return []
  return [
    ...simulationDiagnostics.value.rulesetTraces.map((trace) => ({
      id: `ruleset-${trace.id}`,
      label: trace.label,
      matched: trace.matched,
    })),
    ...simulationDiagnostics.value.conditionTraces.map((trace) => ({
      id: `condition-${trace.id}`,
      label: trace.label,
      matched: trace.matched,
    })),
    ...simulationDiagnostics.value.selectorTraces.map((trace) => ({
      id: `selector-${trace.id}`,
      label: trace.label,
      matched: trace.matched,
    })),
  ].slice(0, 8)
})
const diff = computed(() => payload.value?.diff || null)
const fieldDiffs = computed(() => {
  const result = [...(diff.value?.ruleset_field_diffs || [])]
  for (const ruleDiff of diff.value?.rule_diffs || []) {
    result.push(...(ruleDiff.field_diffs || []))
  }
  return result
})
</script>

<style lang="scss" scoped>
.result-inspector {
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
  overflow: auto;
}

.result-panel {
  flex-shrink: 0;
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--ms-space-3);
    margin-bottom: var(--ms-space-3);
  }

  header span {
    color: var(--ms-text-tertiary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  header strong {
    padding: 4px 9px;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
    font-size: var(--ms-text-sm);
  }

  header strong.ok {
    color: var(--ms-green-600);
    background: var(--ms-green-50);
  }
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--ms-space-2);

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--ms-space-1);
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    background: var(--ms-control-bg);
  }

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
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

.metric-card {
  &.is-ok {
    background: var(--ms-green-50);
  }

  &.is-warn {
    background: var(--ms-amber-50);
  }

  &.is-muted {
    background: var(--ms-control-bg);
  }
}

.candidate-list {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--ms-space-2);
  margin-top: var(--ms-space-3);

  span {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    min-width: 0;
    overflow: hidden;
    max-width: 180px;
    padding: 4px 8px;
    border-radius: var(--ms-radius-pill);
    color: var(--ms-teal-700);
    background: var(--ms-teal-50);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.trace-section {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  margin-top: var(--ms-space-3);
}

.trace-heading {
  color: var(--ms-text-tertiary);
  font-family: var(--ms-font-display);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-bold);
  text-transform: uppercase;
}

.diagnosis-path {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--ms-space-2);
}

.diagnosis-step {
  min-width: 0;
  max-width: 220px;
  display: inline-flex;
  align-items: center;
  gap: var(--ms-space-1);
  padding: 5px 8px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-red-600);
  background: rgba(248, 113, 113, 0.08);

  span {
    flex: 0 0 auto;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  code {
    min-width: 0;
    overflow: hidden;
    color: inherit;
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &.ok {
    color: var(--ms-green-600);
    background: var(--ms-green-50);
  }
}

.diagnostic-row {
  min-width: 0;
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  padding-left: calc(var(--ms-space-2) + var(--trace-depth, 0) * 14px);
  border-radius: var(--ms-radius-md);
  background: rgba(248, 113, 113, 0.08);

  > span {
    color: var(--ms-red-600);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
  }

  &.ok {
    background: var(--ms-green-50);

    > span {
      color: var(--ms-green-600);
    }
  }

  div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  code,
  small {
    min-width: 0;
    overflow: hidden;
    font-size: var(--ms-text-sm);
    line-height: 1.45;
    text-overflow: ellipsis;
  }

  code {
    color: var(--ms-text-secondary);
  }

  small {
    color: var(--ms-text-tertiary);
  }
}

.issue-list,
.explain-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
}

.issue-row,
.explain-row {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(80px, 0.6fr) minmax(0, 1fr);
  gap: var(--ms-space-2);
  padding: var(--ms-space-2);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  code,
  span,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  code {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }

  span,
  small {
    font-size: var(--ms-text-sm);
    line-height: 1.45;
  }
}

.issue-row.is-warning {
  background: var(--ms-amber-50);

  code,
  span {
    color: var(--ms-amber-600);
  }
}

.explain-row {
  grid-template-columns: minmax(80px, 0.6fr) 48px minmax(0, 1fr);

  span {
    color: var(--ms-red-600);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
  }

  span.ok {
    color: var(--ms-green-600);
  }

  small {
    color: var(--ms-text-tertiary);
  }
}

.raw-result {
  min-height: 320px;
  flex: 1;
}

@media (max-width: 720px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
