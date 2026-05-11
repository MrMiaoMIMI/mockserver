<template>
  <div class="fallback-summary">
    <span :class="['fallback-type', `is-${summary.tone}`]">
      {{ summary.label }}
    </span>
    <code :title="summary.detail">{{ summary.detail }}</code>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { NamespaceFallbackAction } from '@/types'
import { fallbackSummary } from '@/utils/entryLists'

const props = defineProps<{
  action: NamespaceFallbackAction
}>()

const summary = computed(() => fallbackSummary(props.action))
</script>

<style lang="scss" scoped>
.fallback-summary {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);
}

.fallback-type {
  flex: 0 0 auto;
  padding: 4px 9px;
  border-radius: var(--ms-radius-pill);
  color: var(--ms-green-600);
  background: var(--ms-green-50);
  font-size: var(--ms-text-sm);
  font-weight: var(--ms-font-semibold);

  &.is-warn {
    color: var(--ms-amber-600);
    background: var(--ms-amber-50);
  }

  &.is-danger {
    color: var(--ms-red-600);
    background: var(--ms-red-50);
  }

  &.is-accent {
    color: var(--ms-blue-500);
    background: var(--ms-blue-50);
  }
}

code {
  min-width: 0;
  overflow: hidden;
  color: var(--ms-text-secondary);
  font-size: var(--ms-text-sm);
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
