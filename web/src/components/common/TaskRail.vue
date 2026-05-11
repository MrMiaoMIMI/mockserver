<template>
  <div class="task-rail" role="tablist" :aria-label="ariaLabel">
    <button
      v-for="task in tasks"
      :key="task.name"
      type="button"
      role="tab"
      :aria-selected="modelValue === task.name"
      :class="{ active: modelValue === task.name }"
      :title="task.description"
      @click="$emit('update:modelValue', task.name)"
    >
      <span>{{ task.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
export interface TaskRailItem {
  name: string
  label: string
  description?: string
}

withDefaults(
  defineProps<{
    modelValue: string
    tasks: TaskRailItem[]
    ariaLabel?: string
  }>(),
  {
    ariaLabel: 'Tasks',
  }
)

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<style lang="scss" scoped>
.task-rail {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--ms-border-light);
  border-radius: var(--ms-radius-lg);
  background: var(--ms-surface-muted);
}

button {
  min-width: 0;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 10px;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: var(--ms-radius-md);
  color: var(--ms-text-secondary);
  background: transparent;
  text-align: center;
  cursor: pointer;
  transition:
    border-color var(--ms-transition-fast),
    background var(--ms-transition-fast),
    box-shadow var(--ms-transition-fast);

  span {
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    color: var(--ms-text-primary);
    font-size: var(--ms-text-sm);
    font-weight: var(--ms-font-bold);
  }

  &:hover {
    border-color: rgba(37, 99, 235, 0.28);
    background: var(--ms-control-bg);
  }

  &.active {
    border-color: rgba(37, 99, 235, 0.42);
    background: var(--ms-control-bg);
    box-shadow: var(--ms-shadow-xs);

    span {
      color: var(--ms-teal-700);
    }
  }
}

@media (max-width: 640px) {
  .task-rail {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
