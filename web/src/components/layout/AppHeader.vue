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
      <div v-if="auth.userEmail" class="user-pill" :title="auth.userEmail">
        <el-icon><User /></el-icon>
        <span>{{ auth.userEmail }}</span>
      </div>
      <button class="icon-button" title="Refresh current data" type="button" @click="refreshAll">
        <el-icon><Refresh /></el-icon>
      </button>
      <button class="icon-button" title="Sign out" type="button" @click="logout">
        <el-icon><SwitchButton /></el-icon>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Expand, Fold, Refresh, SwitchButton, User } from '@element-plus/icons-vue'
import { useAuthStore, useMockserverStore } from '@/store'

defineProps<{
  collapse: boolean
}>()

const emit = defineEmits<{
  'toggle-sidebar': []
}>()

const route = useRoute()
const router = useRouter()
const store = useMockserverStore()
const auth = useAuthStore()

const sectionLabel = computed(() => {
  if (route.path.startsWith('/rulesets')) return 'rule control'
  if (route.path.startsWith('/namespaces')) return 'namespace'
  if (route.path.startsWith('/traffic')) return 'traffic'
  return 'mockserver'
})

async function refreshAll() {
  await Promise.allSettled([
    store.fetchDrafts(),
    store.fetchPublished(),
    store.fetchNamespaces(),
    store.fetchMetrics(),
    store.fetchTrafficSummary(),
  ])
}

async function logout() {
  auth.logout()
  await router.push('/login')
}

onMounted(async () => {
  if (auth.isAuthenticated && !auth.userEmail) {
    await auth.fetchCurrentUser().catch(() => undefined)
  }
  await refreshAll()
})
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

.user-pill {
  min-width: 0;
  max-width: 260px;
  height: 34px;
  display: flex;
  align-items: center;
  gap: var(--ms-space-2);
  padding: 0 var(--ms-space-3);
  border: 1px solid var(--ms-border);
  border-radius: var(--ms-radius-pill);
  color: var(--ms-text-secondary);
  background: rgba(255, 255, 255, 0.7);

  span {
    min-width: 0;
    overflow: hidden;
    font-size: var(--ms-text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 640px) {
  .header {
    padding: 0 var(--ms-space-3);
  }
}
</style>
