<template>
  <div
    :class="['filter-segment', { 'is-compact': compact }]"
    :aria-label="ariaLabel"
    role="group"
  >
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      :class="{ active: modelValue === option.value }"
      :disabled="option.disabled"
      :title="option.title"
      @click="$emit('update:modelValue', option.value)"
    >
      <span>{{ option.label }}</span>
      <small v-if="option.count !== undefined">{{ option.count }}</small>
    </button>
  </div>
</template>

<script setup lang="ts">
export interface FilterSegmentOption {
  label: string
  value: string
  count?: string | number
  title?: string
  disabled?: boolean
}

withDefaults(
  defineProps<{
    modelValue: string
    options: FilterSegmentOption[]
    ariaLabel?: string
    compact?: boolean
  }>(),
  {
    ariaLabel: 'Filter',
    compact: false,
  }
)

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<style lang="scss" scoped>
.filter-segment {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-control-bg);

  button {
    min-width: 0;
    height: 32px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--ms-space-1);
    padding: 0 var(--ms-space-3);
    overflow: hidden;
    border: none;
    border-radius: var(--ms-radius-md);
    color: var(--ms-text-tertiary);
    background: transparent;
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-semibold);
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
    transition:
      color var(--ms-transition-fast),
      background var(--ms-transition-fast),
      box-shadow var(--ms-transition-fast);

    &.active {
      color: var(--ms-text-inverse);
      background: var(--ms-teal-600);
      box-shadow: var(--ms-shadow-xs);
    }

    &:hover:not(.active):not(:disabled),
    &:focus-visible:not(.active):not(:disabled) {
      color: var(--ms-teal-700);
      background: var(--ms-teal-50);
      outline: none;
    }

    &:disabled {
      cursor: not-allowed;
      opacity: 0.5;
    }
  }

  small {
    color: inherit;
    font-family: var(--ms-font-mono);
    font-size: var(--ms-text-sm);
  }
}

.is-compact button {
  height: 28px;
  padding: 0 var(--ms-space-2);
}

@media (max-width: 640px) {
  .filter-segment {
    width: 100%;
    overflow-x: auto;

    button {
      flex: 1 0 auto;
    }
  }
}
</style>
