import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAuthStore } from '@/store'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { title: 'Login', public: true },
    },
    {
      path: '/',
      component: AppLayout,
      redirect: '/rulesets',
      children: [
        {
          path: '/rulesets',
          name: 'RuleSets',
          component: () => import('@/views/rulesets/RuleSetList.vue'),
          meta: { title: 'Rulesets' },
        },
        {
          path: '/namespaces',
          name: 'Namespaces',
          component: () => import('@/views/namespaces/NamespaceList.vue'),
          meta: { title: 'Namespace' },
        },
        {
          path: '/rulesets/:id/rules',
          name: 'RuleSetRules',
          component: () => import('@/views/rulesets/RuleSetRules.vue'),
          meta: { title: 'Rules', hidden: true },
        },
        {
          path: '/dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: 'Traffic' },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/NotFound.vue'),
      meta: { title: 'Not found' },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (to.name === 'Login' && auth.isAuthenticated) {
      return '/rulesets'
    }
    return true
  }
  if (!auth.isAuthenticated) {
    return {
      path: '/login',
      query: { redirect: to.fullPath },
    }
  }
  return true
})

router.afterEach((to) => {
  document.title = `${String(to.meta.title || 'Console')} - MockServer`
})

export default router
