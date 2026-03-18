# 登录流程改造设计文档

## 背景

当前登录逻辑：未登录用户访问 Home 页面会被重定向到独立的 `/login` 页面。

目标：改为弹窗登录，允许未登录用户浏览首页，点击需要登录的操作时弹出登录弹窗。

## 需求

1. **未登录用户可进入 Home 页面**（不再强制跳转）
2. **侧边栏底部登录入口**：文字链接 + icon（如 `🔐 登录`）
3. **弹窗登录**：点击后弹出登录弹窗，不是跳转页面
4. **弹窗内可切换登录/注册**：通过底部文字链接切换
5. **未登录点击推荐订阅**：弹出登录弹窗，登录成功后继续执行订阅

## 设计

### 1. 路由改动

**文件**: `frontend/src/router/index.js`

- 移除 Home 路由的 `meta: { requiresAuth: true }`
- 保留 `/login` 和 `/register` 路由（可选，用于直接访问）
- 路由守卫逻辑调整：已登录用户访问登录页仍重定向到首页

### 2. 新增登录/注册弹窗组件

**新建文件**: `frontend/src/components/AuthModal.vue`

```vue
<template>
  <div class="modal-overlay" @click.self="close">
    <div class="modal">
      <div class="modal-header">
        <h3>{{ isLogin ? '登录' : '注册' }}</h3>
      </div>
      <div class="modal-body">
        <!-- 登录表单 -->
        <template v-if="isLogin">
          <input v-model="email" placeholder="邮箱" type="email" />
          <input v-model="password" placeholder="密码" type="password" />
          <button @click="handleLogin">登录</button>
          <p class="switch-link">没有账号？<a @click="isLogin = false">去注册</a></p>
        </template>
        <!-- 注册表单 -->
        <template v-else>
          <input v-model="email" placeholder="邮箱" type="email" />
          <input v-model="password" placeholder="密码" type="password" />
          <button @click="handleRegister">注册</button>
          <p class="switch-link">已有账号？<a @click="isLogin = true">去登录</a></p>
        </template>
      </div>
    </div>
  </div>
</template>
```

**Props**:
- `modelValue: boolean` - 控制显示/隐藏（v-model）
- `onSuccess: Function` - 登录/注册成功回调，用于执行被拦截的操作

**功能**:
- 登录/注册切换
- 登录成功后调用 `onSuccess` 回调并关闭弹窗
- 调用 `authStore.login()` / `authStore.register()`

### 3. 侧边栏底部改造

**文件**: `frontend/src/views/Home.vue`

**未登录状态**:
```vue
<div class="user-section">
  <button class="login-link" @click="showAuthModal = true">
    <span>🔐</span>
    <span>登录</span>
  </button>
</div>
```

**已登录状态**（保持现有）:
```vue
<div class="user-section">
  <div class="user-info" @click="showUserMenu = !showUserMenu">
    ...
  </div>
</div>
```

**样式**: 登录链接使用文字 + 图标，hover 时有背景变化

### 4. 推荐订阅拦截逻辑

**文件**: `frontend/src/views/Home.vue`

```js
async function addRecommendedFeed(rec) {
  if (rec.added || rec.adding) return
  
  // 检查登录状态
  if (!authStore.isLoggedIn) {
    pendingAction.value = () => doAddFeed(rec)
    showAuthModal.value = true
    return
  }
  
  doAddFeed(rec)
}

async function doAddFeed(rec) {
  rec.adding = true
  try {
    await api.post('/feeds', { url: rec.url, title: rec.name, category: rec.category })
    rec.added = true
    fetchFeeds()
  } catch (e) {
    console.error('Failed to add feed:', e)
    alert('添加失败: ' + (e.response?.data?.error || '未知错误'))
  } finally {
    rec.adding = false
  }
}

function onAuthSuccess() {
  showAuthModal.value = false
  if (pendingAction.value) {
    pendingAction.value()
    pendingAction.value = null
  }
}
```

### 5. Auth Store 改造

**文件**: `frontend/src/stores/auth.js`

- 确保 `isLoggedIn` 响应式更新
- 添加 `init()` 方法在应用启动时调用（检查 token 有效性）

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `frontend/src/router/index.js` | 修改：移除 Home 的 requiresAuth |
| `frontend/src/stores/auth.js` | 修改：确保 init() 方法可用 |
| `frontend/src/components/AuthModal.vue` | 新增：登录/注册弹窗组件 |
| `frontend/src/views/Home.vue` | 修改：底部登录入口 + 弹窗集成 + 订阅拦截 |

## 交互流程

```
未登录用户进入首页
    ↓
看到首页内容（推荐订阅源等）
    ↓
点击「订阅」
    ↓
弹出登录弹窗
    ↓
登录成功
    ↓
自动执行订阅操作
```

```
未登录用户
    ↓
点击左下角「🔐 登录」
    ↓
弹出登录弹窗
    ↓
登录成功
    ↓
弹窗关闭，底部变为用户头像菜单
```
