<template>
  <div class="condition-tree" role="tree" :aria-label="ariaLabel">
    <div
      v-for="row in rows"
      :key="row.number"
      class="condition-row"
      role="treeitem"
      :style="{ '--condition-depth': row.depth }"
    >
      <span class="condition-index">{{ row.number }}</span>
      <div>
        <strong>{{ row.label }}</strong>
        <code v-if="row.detail" :title="row.detail">{{ row.detail }}</code>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Condition } from '@/types'
import { buildConditionTreeRows } from '@/utils/conditionTreeRows'

const props = withDefaults(
  defineProps<{
    condition: Condition
    ariaLabel?: string
  }>(),
  {
    ariaLabel: 'Condition tree',
  }
)

const rows = computed(() => buildConditionTreeRows(props.condition))
</script>

<style lang="scss" scoped>
.condition-tree {
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-1);
}

.condition-row {
  --condition-depth: 0;
  min-width: 0;
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr);
  align-items: start;
  gap: var(--ms-space-2);
  padding: 8px 10px 8px calc(10px + var(--condition-depth) * 18px);
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-md);
  background: var(--ms-control-bg);

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
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
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
  }

  code {
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
  }
}

.condition-index {
  align-self: start;
  justify-self: start;
  padding: 2px 7px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-teal-700);
  background: var(--ms-teal-50);
  font-family: var(--ms-font-mono);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);
}
</style>
