<template>
  <div class="header">
    <div class="header-left">
      <button
        class="icon-button"
        :title="collapse ? 'Expand sidebar' : 'Collapse sidebar'"
        type="button"
        @click="emit('toggle-sidebar')"
      >
        <el-icon>
          <Expand v-if="collapse" />
          <Fold v-else />
        </el-icon>
      </button>
      <div class="route-title">
        <strong>{{ sectionLabel }}</strong>
      </div>
    </div>

    <div class="header-right">
      <button class="icon-button" title="Refresh current data" type="button" @click="refreshAll">
        <el-icon><Refresh /></el-icon>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Expand, Fold, Refresh } from '@element-plus/icons-vue'
import { useMockserverStore } from '@/store'

defineProps<{
  collapse: boolean
}>()

const emit = defineEmits<{
  'toggle-sidebar': []
}>()

const route = useRoute()
const store = useMockserverStore()

const sectionLabel = computed(() => {
  if (route.path.startsWith('/rulesets')) return 'rule control'
  if (route.path.startsWith('/namespaces')) return 'namespace'
  if (route.path.startsWith('/dashboard')) return 'traffic'
  return 'mockserver'
})

async function refreshAll() {
  await Promise.allSettled([
    store.fetchDrafts(),
    store.fetchPublished(),
    store.fetchNamespaces(),
    store.fetchMetrics(),
    store.fetchTrafficEvents({ limit: 50, include_indexes: true }),
  ])
}

onMounted(refreshAll)
</script>

<style lang="scss" scoped>
.header {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ms-space-4);
  padding: 0 var(--ms-space-4);
}

.header-left,
.header-right {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: var(--ms-space-3);
}

.route-title {
  min-width: 0;
  display: flex;
  align-items: center;

  strong {
    overflow: hidden;
    color: var(--ms-teal-700);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-base);
    font-weight: var(--ms-font-bold);
    text-overflow: ellipsis;
    text-transform: uppercase;
    white-space: nowrap;
  }
}

@media (max-width: 640px) {
  .header {
    padding: 0 var(--ms-space-3);
  }
}
</style>
