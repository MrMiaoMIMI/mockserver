<template>
  <main class="login-page">
    <section class="login-panel" aria-label="MockServer login">
      <div class="login-brand">
        <div class="brand-mark">
          <el-icon><Connection /></el-icon>
        </div>
        <div>
          <strong>MockServer</strong>
          <span>admin console</span>
        </div>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent>
        <el-form-item label="Email" prop="email">
          <el-input
            v-model.trim="form.email"
            autocomplete="email"
            clearable
            placeholder="name@example.com"
            size="large"
            @keyup.enter="submit"
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-button class="login-button" :loading="auth.loading" size="large" type="primary" @click="submit">
          <el-icon><Right /></el-icon>
          <span>Sign in</span>
        </el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { Connection, Message, Right } from '@element-plus/icons-vue'
import { useAuthStore } from '@/store'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const form = reactive({
  email: '',
})

const rules: FormRules<typeof form> = {
  email: [
    { required: true, message: 'Email is required', trigger: 'blur' },
    { type: 'email', message: 'Enter a valid email', trigger: ['blur', 'change'] },
  ],
}

async function submit() {
  await formRef.value?.validate()
  await auth.debugLogin(form.email)
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/rulesets'
  await router.replace(redirect)
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: var(--ms-space-5);
  background:
    linear-gradient(135deg, rgba(20, 184, 166, 0.14), transparent 38%),
    linear-gradient(315deg, rgba(37, 99, 235, 0.12), transparent 42%),
    var(--ms-canvas);
}

.login-panel {
  width: min(420px, 100%);
  display: flex;
  flex-direction: column;
  gap: var(--ms-space-5);
  padding: var(--ms-space-6);
  border: 1px solid var(--ms-border);
  border-radius: var(--ms-radius-lg);
  background: rgba(255, 255, 255, 0.92);
  box-shadow: var(--ms-shadow-lg);
}

.login-brand {
  display: flex;
  align-items: center;
  gap: var(--ms-space-3);

  strong,
  span {
    display: block;
  }

  strong {
    color: var(--ms-text-primary);
    font-family: var(--ms-font-display);
    font-size: var(--ms-text-xl);
    font-weight: var(--ms-font-bold);
  }

  span {
    margin-top: 3px;
    color: var(--ms-text-secondary);
    font-size: var(--ms-text-sm);
  }
}

.brand-mark {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--ms-radius-lg);
  color: #ffffff;
  background: linear-gradient(135deg, #2563eb, #0f766e);
}

.login-button {
  width: 100%;
}
</style>
