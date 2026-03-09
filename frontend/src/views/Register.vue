<template>
  <div class="auth-page">
    <div class="auth-container">
      <h2>注册</h2>
      <div v-if="error" class="message message-error">{{ error }}</div>
      <div v-if="success" class="message message-success">{{ success }}</div>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label class="form-label">邮箱</label>
          <input v-model="email" type="email" class="form-input" required />
        </div>
        <div class="form-group">
          <label class="form-label">密码</label>
          <input v-model="password" type="password" class="form-input" required minlength="6" />
        </div>
        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>
      <p class="auth-switch">
        已有账号？ <router-link to="/login">登录</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)

async function handleRegister() {
  error.value = ''
  success.value = ''
  loading.value = true
  try {
    await authStore.register(email.value, password.value)
    success.value = '注册成功！正在跳转登录...'
    setTimeout(() => router.push('/login'), 1500)
  } catch (e) {
    error.value = e.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
}
</style>
