<template>
  <div class="push-config-page">
    <!-- Page Header -->
    <div class="page-header">
      <button class="btn btn-back" @click="goBack">← 返回</button>
      <h1>推送配置</h1>
      <button class="btn btn-primary" @click="showCreateDialog = true" v-if="!config">新建配置</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading">加载中...</div>

    <!-- Config Card -->
    <div v-else-if="config" class="config-card">
      <div class="config-header">
        <h3>当前配置</h3>
        <div class="config-actions">
          <button class="btn btn-sm btn-secondary" @click="editConfig">编辑</button>
          <button class="btn btn-sm btn-danger" @click="deleteConfig">删除</button>
        </div>
      </div>
      <div class="config-details">
        <div class="config-item">
          <label>Webhook URL</label>
          <span class="config-value config-webhook">{{ config.webhook_url }}</span>
        </div>
        <div class="config-item">
          <label>推送频率</label>
          <span class="config-value">{{ frequencyLabel }}</span>
        </div>
        <div class="config-item">
          <label>推送时间</label>
          <span class="config-value">{{ config.push_time }}</span>
        </div>
        <div class="config-item">
          <label>最小未读数</label>
          <span class="config-value">{{ config.min_unread_count }}</span>
        </div>
        <div class="config-item">
          <label>最多推送文章数</label>
          <span class="config-value">{{ config.max_article_count }}</span>
        </div>
        <div class="config-item">
          <label>订阅源过滤</label>
          <span class="config-value">{{ config.feed_ids?.length || 0 }} 个</span>
        </div>
        <div class="config-item">
          <label>分类过滤</label>
          <span class="config-value">{{ config.category_ids?.length || 0 }} 个</span>
        </div>
        <div class="config-item" v-if="config.last_push_at">
          <label>上次推送</label>
          <span class="config-value">{{ formatDateTime(config.last_push_at) }}</span>
        </div>
      </div>
      <div class="config-actions">
        <button class="btn btn-secondary" @click="testPush">📨 测试推送</button>
      </div>
    </div>

    <!-- No Config -->
    <div v-else class="no-config">
      <p>暂无推送配置</p>
      <button class="btn btn-primary" @click="showCreateDialog = true">创建配置</button>
    </div>

    <!-- Create/Edit Dialog -->
    <div v-if="showCreateDialog || showEditDialog" class="modal-overlay" @click.self="closeDialog">
      <div class="modal">
        <h2>{{ showEditDialog ? '编辑配置' : '新建配置' }}</h2>
        <form @submit.prevent="saveConfig">
          <div class="form-group">
            <label>Webhook URL <span class="required">*</span></label>
            <input type="url" v-model="formData.webhook_url" required placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" />
            <small class="form-hint">飞书机器人 Webhook URL</small>
          </div>
          <div class="form-group">
            <label>推送频率 <span class="required">*</span></label>
            <select v-model="formData.frequency" required>
              <option value="daily">每日</option>
              <option value="weekly">每周</option>
              <option value="monthly">每月</option>
            </select>
          </div>
          <div class="form-group">
            <label>推送时间 <span class="required">*</span></label>
            <input type="time" v-model="formData.push_time" required />
            <small class="form-hint">推送时间（24 小时制，如 09:00）</small>
          </div>
          <div class="form-group">
            <label>最小未读数</label>
            <input type="number" v-model="formData.min_unread_count" min="0" />
            <small class="form-hint">只有当未读文章数量达到此值时才推送</small>
          </div>
          <div class="form-group">
            <label>最多推送文章数</label>
            <input type="number" v-model="formData.max_article_count" min="1" max="50" />
            <small class="form-hint">每次推送最多包含的文章数量（1-50）</small>
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
import { ref, computed, onMounted } from 'vue'
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
  min_unread_count: 1,
  max_article_count: 10
})

const frequencyLabel = computed(() => {
  if (!config.value) return ''
  const labels = {
    daily: '每日',
    weekly: '每周',
    monthly: '每月'
  }
  return labels[config.value.frequency] || config.value.frequency
})

const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const goBack = () => {
  router.push('/')
}

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
    alert('获取配置失败，请稍后重试')
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
      alert('保存成功')
    } else {
      const error = await response.json()
      alert(error.error || '保存失败，请稍后重试')
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    alert('保存失败，请稍后重试')
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
      alert('删除成功')
    } else {
      alert('删除失败，请稍后重试')
    }
  } catch (error) {
    console.error('删除配置失败:', error)
    alert('删除失败，请稍后重试')
  }
}

const testPush = async () => {
  if (!config.value?.id) return

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/push-configs/${config.value.id}/test`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (response.ok) {
      alert('✅ 测试推送成功，请检查飞书消息')
    } else {
      const error = await response.json()
      alert('❌ 测试推送失败：' + (error.error || '未知错误'))
    }
  } catch (error) {
    console.error('测试推送失败:', error)
    alert('❌ 测试推送失败，请稍后重试')
  }
}

const closeDialog = () => {
  showCreateDialog.value = false
  showEditDialog.value = false
  formData.value = {
    webhook_url: '',
    frequency: 'daily',
    push_time: '09:00',
    min_unread_count: 1,
    max_article_count: 10
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.push-config-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  gap: 16px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  flex: 1;
}

/* Buttons */
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-back {
  background: var(--bg-secondary);
  color: var(--text-primary);
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
  padding: 6px 12px;
  font-size: 13px;
}

/* Loading */
.loading {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
  font-size: 16px;
}

/* Config Card */
.config-card {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border);
}

.config-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.config-actions {
  display: flex;
  gap: 8px;
}

/* Config Details */
.config-details {
  margin-bottom: 24px;
}

.config-item {
  display: flex;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
  align-items: center;
}

.config-item:last-child {
  border-bottom: none;
}

.config-item label {
  width: 140px;
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 14px;
}

.config-value {
  flex: 1;
  color: var(--text-primary);
  font-size: 14px;
}

.config-webhook {
  font-family: monospace;
  font-size: 13px;
  word-break: break-all;
  color: var(--text-muted);
}

.required {
  color: var(--danger);
  margin-left: 4px;
}

.form-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-muted);
  font-weight: normal;
}

/* No Config */
.no-config {
  text-align: center;
  padding: 80px 20px;
  color: var(--text-secondary);
}

.no-config p {
  font-size: 16px;
  margin-bottom: 20px;
}

/* Modal */
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
  padding: 24px;
  max-width: 520px;
  width: 90%;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.modal h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  font-size: 14px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(74, 144, 226, 0.1);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid var(--border);
}

/* Responsive */
@media (max-width: 600px) {
  .push-config-page {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    gap: 12px;
  }

  .page-header h1 {
    font-size: 20px;
    text-align: center;
  }

  .config-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .config-actions {
    justify-content: flex-end;
  }

  .config-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .config-item label {
    width: auto;
  }

  .modal {
    padding: 20px;
  }
}
</style>
