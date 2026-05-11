<template>
  <section class="rule-overview-workbench">
    <el-empty v-if="!ruleSet" description="Select a draft ruleset" />

    <template v-else-if="rule">
      <header class="overview-hero">
        <div>
          <span>selected rule</span>
          <strong :title="ruleDisplayName(rule)">{{ ruleDisplayName(rule) }}</strong>
          <small :title="rule.id">{{ ruleTechnicalLabel(rule) }}</small>
        </div>
        <div class="overview-actions">
          <el-button :icon="EditPen" @click="$emit('edit', rule)">Edit</el-button>
          <el-button type="primary" :icon="VideoPlay" @click="$emit('simulate')">
            Simulate
          </el-button>
        </div>
      </header>

      <div class="overview-metrics">
        <div>
          <span>priority</span>
          <strong>{{ rule.priority }}</strong>
        </div>
        <div>
          <span>state</span>
          <strong>{{ rule.enabled ? 'enabled' : 'disabled' }}</strong>
        </div>
        <div>
          <span>action</span>
          <strong>{{ actionTypeLabel(rule.action.type) }}</strong>
        </div>
        <div>
          <span>diagnostics</span>
          <strong>{{ diagnosticLabel }}</strong>
        </div>
      </div>

      <section class="overview-panel">
        <header>
          <span>ruleset gate</span>
          <strong>{{ selectorRows.length }} nodes</strong>
        </header>
        <ConditionTreePreview :condition="selectorCondition" aria-label="Ruleset selector gate" />
      </section>

      <section class="overview-panel">
        <header>
          <span>rule condition</span>
          <strong>{{ conditionRows.length }} nodes</strong>
        </header>
        <ConditionTreePreview :condition="rule.when" :aria-label="`Condition tree for ${rule.id}`" />
      </section>

      <section class="overview-panel">
        <header>
          <span>action</span>
          <strong>Response behavior</strong>
        </header>
        <code :title="action">{{ action }}</code>
      </section>

      <section v-if="diagnostic?.messages?.length" class="overview-panel">
        <header>
          <span>latest signals</span>
          <strong>{{ diagnostic.messages.length }} notes</strong>
        </header>
        <div class="signal-list">
          <span v-for="message in diagnostic.messages" :key="message" :title="message">
            {{ message }}
          </span>
        </div>
      </section>

      <footer class="overview-footer">
        <el-button :icon="Check" @click="$emit('validate')">Check readiness</el-button>
        <el-button :icon="EditPen" @click="$emit('edit', rule)">Edit rule</el-button>
      </footer>
    </template>

    <el-empty v-else description="Select a rule on the left to inspect it here">
      <el-button type="primary" :icon="Plus" @click="$emit('create')">Add rule</el-button>
    </el-empty>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Check, EditPen, Plus, VideoPlay } from '@element-plus/icons-vue'
import ConditionTreePreview from '@/components/rulesets/ConditionTreePreview.vue'
import type { Rule, RuleSet } from '@/types'
import type { RuleDiagnosticState } from '@/utils/ruleCollection'
import { buildConditionTreeRows } from '@/utils/conditionTreeRows'
import { actionSummary, actionTypeLabel, ruleDisplayName, ruleTechnicalLabel } from '@/utils/ruleSummaries'

const props = defineProps<{
  ruleSet: RuleSet | null
  rule: Rule | null
  diagnostic?: RuleDiagnosticState
}>()

defineEmits<{
  create: []
  edit: [rule: Rule]
  simulate: []
  validate: []
}>()

const conditionRows = computed(() => (props.rule ? buildConditionTreeRows(props.rule.when) : []))
const selectorCondition = computed(() => ({ all: props.ruleSet?.selector.all || [] }))
const selectorRows = computed(() => (props.ruleSet ? buildConditionTreeRows(selectorCondition.value) : []))
const action = computed(() => (props.rule ? actionSummary(props.rule.action) : ''))
const diagnosticLabel = computed(() => {
  if (props.diagnostic?.validation === 'invalid') return 'invalid'
  if (props.diagnostic?.simulation === 'not_reached') return 'not reached'
  if (props.diagnostic?.simulation) return props.diagnostic.simulation
  return 'none'
})
</script>

<style lang="scss" scoped>
.rule-overview-workbench {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-3);
}

.overview-hero {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--ms-space-3);
  padding: var(--ms-space-3);
  border: 1px solid rgba(37, 99, 235, 0.22);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-teal-50);

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

  strong,
  small {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-lg);
  }

  small {
    color: var(--ms-text-tertiary);
    font-size: var(--ms-text-sm);
  }
}

.overview-actions,
.overview-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--ms-space-2);
}

.overview-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
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
    text-transform: uppercase;
  }

  strong {
    overflow: hidden;
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.overview-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);
  padding: var(--ms-space-3);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-panel-bg);

  header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--ms-space-3);
  }

  header span {
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
    text-transform: uppercase;
  }

  header strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-md);
  }

  code {
    overflow: hidden;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
    line-height: 1.55;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.signal-list {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-2);

  span {
    min-width: 0;
    overflow: hidden;
    padding: var(--ms-space-2);
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-secondary);
    background: var(--ms-control-bg);
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 720px) {
  .overview-hero,
  .overview-metrics {
    grid-template-columns: 1fr;
  }

  .overview-actions,
  .overview-footer {
    justify-content: stretch;

    :deep(.el-button) {
      width: 100%;
      margin-left: 0;
    }
  }
}
</style>
