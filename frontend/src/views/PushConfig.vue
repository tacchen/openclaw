<template>
  <div class="push-config-page">
    <div class="page-header">
      <h1>推送配置</h1>
      <button class="btn btn-primary" @click="showCreateDialog = true">新建配置</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="config" class="config-card">
      <div class="config-header">
        <h3>当前配置</h3>
        <button class="btn btn-sm" @click="editConfig">编辑</button>
        <button class="btn btn-sm btn-danger" @click="deleteConfig">删除</button>
      </div>
      <div class="config-details">
        <div class="config-item">
          <label>推送频率：</label>
          <span>{{ config.frequency }}</span>
        </div>
        <div class="config-item">
          <label>推送时间：</label>
          <span>{{ config.push_time }}</span>
        </div>
        <div class="config-item">
          <label>最小未读数：</label>
          <span>{{ config.min_unread_count }}</span>
        </div>
        <div class="config-item">
          <label>订阅源过滤：</label>
          <span>{{ config.feed_ids?.length || 0 }} 个</span>
        </div>
        <div class="config-item">
          <label>分类过滤：</label>
          <span>{{ config.category_ids?.length || 0 }} 个</span>
        </div>
      </div>
      <div class="config-actions">
        <button class="btn btn-secondary" @click="testPush">测试推送</button>
      </div>
    </div>
    <div v-else class="no-config">
      <p>暂无推送配置，点击"新建配置"创建一个</p>
    </div>

    <!-- Create/Edit Dialog -->
    <div v-if="showCreateDialog || showEditDialog" class="modal-overlay" @click.self="closeDialog">
      <div class="modal">
        <h2>{{ showEditDialog ? '编辑配置' : '新建配置' }}</h2>
        <form @submit.prevent="saveConfig">
          <div class="form-group">
            <label>Webhook URL</label>
            <input type="url" v-model="formData.webhook_url" required />
          </div>
          <div class="form-group">
            <label>推送频率</label>
            <select v-model="formData.frequency" required>
              <option value="daily">每日</option>
              <option value="weekly">每周</option>
              <option value="monthly">每月</option>
            </select>
          </div>
          <div class="form-group">
            <label>推送时间</label>
            <input type="time" v-model="formData.push_time" required />
          </div>
          <div class="form-group">
            <label>最小未读数</label>
            <input type="number" v-model="formData.min_unread_count" min="0" />
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-secondary" @click="closeDialog">取消</button>
            <button type="submit" class="btn btn-primary">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const config = ref(null)
const loading = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const formData = ref({
  webhook_url: '',
  frequency: 'daily',
  push_time: '09:00',
  min_unread_count: 1
})

const fetchConfig = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/push-configs', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    const data = await response.json()
    if (data.length > 0) {
      config.value = data[0]
    }
  } catch (error) {
    console.error('获取配置失败:', error)
  } finally {
    loading.value = false
  }
}

const editConfig = () => {
  if (config.value) {
    formData.value = { ...config.value }
    showEditDialog.value = true
  }
}

const saveConfig = async () => {
  try {
    const token = localStorage.getItem('token')
    const url = showEditDialog.value ? `/api/push-configs/${config.value.id}` : '/api/push-configs'
    const method = showEditDialog.value ? 'PUT' : 'POST'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(formData.value)
    })

    if (response.ok) {
      closeDialog()
      await fetchConfig()
    }
  } catch (error) {
    console.error('保存配置失败:', error)
  }
}

const deleteConfig = async () => {
  if (!confirm('确定删除推送配置吗？')) return

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/push-configs/${config.value.id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (response.ok) {
      config.value = null
    }
  } catch (error) {
    console.error('删除配置失败:', error)
  }
}

const testPush = async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/push-configs/${config.value.id}/test`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (response.ok) {
      alert('测试推送成功')
    }
  } catch (error) {
    console.error('测试推送失败:', error)
    alert('测试推送失败')
  }
}

const closeDialog = () => {
  showCreateDialog.value = false
  showEditDialog.value = false
  formData.value = {
    webhook_url: '',
    frequency: 'daily',
    push_time: '09:00',
    min_unread_count: 1
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.push-config-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
}

.loading {
  text-align: center;
  padding: 40px;
}

.config-card {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 20px;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--border);
}

.config-header h3 {
  margin: 0;
}

.config-details {
  margin-bottom: 20px;
}

.config-item {
  display: flex;
  padding: 10px 0;
  border-bottom: 1px solid var(--border);
}

.config-item:last-child {
  border-bottom: none;
}

.config-item label {
  width: 120px;
  font-weight: 500;
  color: var(--text-secondary);
}

.config-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.no-config {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

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
  border-radius: var(--radius-md);
  padding: 20px;
  max-width: 500px;
  width: 90%;
}

.modal h2 {
  margin: 0 0 20px 0;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-primary);
  color: var(--text-primary);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-weight: 500;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.btn-danger {
  background: var(--danger);
  color: white;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}
</style>
