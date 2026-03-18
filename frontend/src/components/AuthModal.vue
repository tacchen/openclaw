<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="close">
    <div class="modal">
      <div class="modal-header">
        <h3 class="modal-title">{{ isLogin ? '登录' : '注册' }}</h3>
      </div>
      <div class="modal-body">
        <div v-if="error" class="message message-error">{{ error }}</div>
        
        <div class="form-group">
          <label class="form-label">邮箱</label>
          <input 
            v-model="email" 
            type="email" 
            class="form-input" 
            placeholder="请输入邮箱"
            @keyup.enter="handleSubmit"
          />
        </div>
        
        <div class="form-group">
          <label class="form-label">密码</label>
          <input 
            v-model="password" 
            type="password" 
            class="form-input" 
            placeholder="请输入密码"
            @keyup.enter="handleSubmit"
          />
        </div>
        
        <button 
          class="btn btn-primary btn-block" 
          @click="handleSubmit"
          :disabled="loading"
        >
          {{ loading ? '处理中...' : (isLogin ? '登录' : '注册') }}
        </button>
        
        <p class="switch-link">
          <template v-if="isLogin">
            没有账号？<a @click="switchToRegister">去注册</a>
          </template>
          <template v-else>
            已有账号？<a @click="switchToLogin">去登录</a>
          </template>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps({
  modelValue: Boolean,
  defaultMode: {
    type: String,
    default: 'login' // 'login' or 'register'
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const authStore = useAuthStore()
const isLogin = ref(props.defaultMode === 'login')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

watch(() => props.modelValue, (val) => {
  if (val) {
    // 打开时重置状态
    error.value = ''
    isLogin.value = props.defaultMode === 'login'
  }
})

function close() {
  emit('update:modelValue', false)
}

function switchToRegister() {
  isLogin.value = false
  error.value = ''
}

function switchToLogin() {
  isLogin.value = true
  error.value = ''
}

async function handleSubmit() {
  if (!email.value || !password.value) {
    error.value = '请填写邮箱和密码'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    if (isLogin.value) {
      await authStore.login(email.value, password.value)
    } else {
      await authStore.register(email.value, password.value)
      // 注册成功后自动登录
      await authStore.login(email.value, password.value)
    }
    
    emit('success')
    close()
  } catch (e) {
    error.value = e.response?.data?.error || (isLogin.value ? '登录失败' : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 400px;
  margin: 16px;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}

.modal-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 6px;
  color: var(--text-secondary);
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
}

.form-input:focus {
  outline: none;
  border-color: var(--primary);
}

.btn-block {
  width: 100%;
  padding: 12px;
  margin-top: 8px;
}

.message-error {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger);
  padding: 10px 12px;
  border-radius: var(--radius-md);
  font-size: 14px;
  margin-bottom: 16px;
}

.switch-link {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: var(--text-muted);
}

.switch-link a {
  color: var(--primary);
  cursor: pointer;
  text-decoration: none;
}

.switch-link a:hover {
  text-decoration: underline;
}
</style>
