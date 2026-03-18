<template>
  <div class="app-container" :data-theme="theme">
    <!-- Sidebar -->
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <span>📡</span>
          <span>RSS Reader</span>
        </div>
        <button class="btn btn-ghost btn-icon" @click="toggleTheme" :title="theme === 'dark' ? '切换亮色' : '切换暗色'">
          {{ theme === 'dark' ? '☀️' : '🌙' }}
        </button>
      </div>
      
      <div class="sidebar-content">
        <!-- Home Button -->
        <button 
          class="home-button"
          :class="{ active: showRecommendations }"
          @click="goHome"
        >
          <span class="home-icon">🏠</span>
          <span class="home-text">首页</span>
        </button>
        

        <!-- Category Filter - Multi-select Dropdown -->
        <div class="category-filter">
          <div class="dropdown-wrapper">
            <button class="dropdown-trigger" type="button" @click.stop="showCategoryDropdown = !showCategoryDropdown">
              <span>{{ categoryFilterLabel }}</span>
              <span class="dropdown-arrow" :class="{ open: showCategoryDropdown }">▼</span>
            </button>
            <div v-show="showCategoryDropdown" class="dropdown-menu">
              <div class="dropdown-item" @click="clearCategories">
                <span class="dropdown-checkbox" :class="{ checked: selectedCategories.length === 0 }">✓</span>
                <span>全部分类</span>
              </div>
              <div class="dropdown-divider"></div>
              <div 
                v-for="cat in categories" 
                :key="cat" 
                class="dropdown-item"
                @click="toggleCategory(cat)"
              >
                <span class="dropdown-checkbox" :class="{ checked: selectedCategories.includes(cat) }">✓</span>
                <span>{{ cat }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Feeds Section -->
        <div class="sidebar-section">
          <div class="sidebar-title" style="display: flex; justify-content: space-between; align-items: center;">
            <span>订阅源</span>
            <button class="btn-add-feed" @click="showAddFeed = true" title="添加订阅源">+</button>
          </div>
          <ul class="feed-list">
            <li 
              class="feed-item" 
              :class="{ active: selectedFeedId === 0 && selectedTagId === 0 }"
              @click="selectFeed(0)"
            >
              <div class="feed-icon">📰</div>
              <div class="feed-info">
                <div class="feed-name">全部文章</div>
              </div>
            </li>
            <li 
              v-for="feed in filteredFeeds" 
              :key="feed.id"
              class="feed-item"
              :class="{ active: selectedFeedId === feed.id && selectedTagId === 0 }"
              @click="selectFeed(feed.id)"
            >
              <img v-if="feed.icon_url" :src="feed.icon_url" class="feed-icon-img" @error="$event.target.style.display='none'" /><div v-else class="feed-icon">{{ getFeedIcon(feed.category) }}</div>
              <div class="feed-info">
                <div class="feed-name">{{ feed.title || feed.url }}</div>
                <div class="feed-meta" v-if="feed.category">{{ feed.category }}</div>
              </div>
            </li>
          </ul>
        </div>

        <!-- Tags Section -->
        <div class="sidebar-section">
          <div class="sidebar-title" style="display: flex; justify-content: space-between; align-items: center;">
            <span>标签</span>
            <button class="btn btn-ghost" style="padding: 2px 6px; font-size: 12px;" @click="showAddTag = true">+ 新建</button>
          </div>
          <ul class="feed-list">
            <li 
              v-for="tag in tags" 
              :key="tag.id"
              class="feed-item"
              :class="{ active: selectedTagId === tag.id }"
              @click="selectTag(tag.id)"
            >
              <div class="feed-icon">🏷️</div>
              <div class="feed-info">
                <div class="feed-name">{{ tag.name }}</div>
              </div>
              <button class="btn btn-ghost" style="padding: 2px 6px; font-size: 12px; opacity: 0.5;" @click.stop="deleteTag(tag.id)">×</button>
            </li>
          </ul>
          <div v-if="tags.length === 0" style="color: var(--text-muted); font-size: 12px; padding: 8px;">
            暂无标签，阅读文章时可添加标签
          </div>
        </div>
      </div>
      
      <!-- User Section -->
      <div class="user-section">
        <!-- 未登录：显示登录入口 -->
        <button v-if="!authStore.isLoggedIn" class="login-link" @click="showAuthModal = true">
          <span class="login-icon">🔐</span>
          <span class="login-text">登录</span>
        </button>
        
        <!-- 已登录：显示用户信息 -->
        <template v-else>
          <div class="user-info" @click="showUserMenu = !showUserMenu">
            <div class="user-avatar">{{ userEmail?.charAt(0).toUpperCase() }}</div>
            <div class="user-details">
              <div class="user-email">{{ userEmail }}</div>
            </div>
            <span class="user-arrow" :class="{ rotated: showUserMenu }">▼</span>
          </div>
          
          <!-- User Menu Dropdown -->
          <div v-if="showUserMenu" class="user-menu">
            <div class="menu-item" @click="showProfile = true; showUserMenu = false">
              <span>👤</span>
              <span>个人信息</span>
            </div>
            <div class="menu-item" @click="showFeedManager = true; showUserMenu = false">
              <span>📡</span>
              <span>订阅源管理</span>
            </div>
            <div class="menu-divider"></div>
            <div class="menu-item menu-item-danger" @click="logout">
              <span>🚪</span>
              <span>退出登录</span>
            </div>
          </div>
        </template>
      </div>
    </aside>
    <!-- Resizer -->
    <div class="resizer" :style="{ left: (sidebarWidth - 2) + 'px' }" @mousedown="startResize"></div>

    <!-- Main Content -->
    <main class="main-content">
      <!-- Header -->
      <header class="header">
        <div class="header-left">
          <h1 class="header-title">{{ currentTitle }} <span class="header-count">({{ readFilter === "unread" ? unreadCount.total : (readFilter === "analyzed" ? analyzedCount : total) }})</span></h1>
        </div>
        <div class="header-right">
          <button v-if="readFilter !== 'read'" class="btn btn-ghost btn-compact" @click="markAllAsRead(selectedFeedId, selectedCategories)" 
                  :disabled="unreadCount.total === 0" title="全部标记已读">✓ 全部已读</button>
          <div class="dropdown-wrapper">
            <button class="dropdown-trigger dropdown-trigger-sm" type="button" @click.stop="showReadFilterDropdown = !showReadFilterDropdown">
              <span>{{ readFilterLabel }}</span>
              <span class="dropdown-arrow" :class="{ open: showReadFilterDropdown }">▼</span>
            </button>
            <div v-show="showReadFilterDropdown" class="dropdown-menu dropdown-menu-right">
              <div 
                class="dropdown-item" 
                :class="{ active: readFilter === '' }"
                @click="setReadFilter('')"
              >全部</div>
              <div 
                class="dropdown-item" 
                :class="{ active: readFilter === 'unread' }"
                @click="setReadFilter('unread')"
              >未读</div>
              <div 
                class="dropdown-item" 
                :class="{ active: readFilter === 'read' }"
                @click="setReadFilter('read')"
              >已读</div>
              <div 
                class="dropdown-item" 
                :class="{ active: readFilter === 'analyzed' }"
                @click="setReadFilter('analyzed')"
              >已分析</div>
            </div>
          </div>
          <div class="search-bar">
            <span class="search-icon">🔍</span>
            <input 
              v-model="searchQuery" 
              placeholder="搜索文章..."
              @keyup.enter="searchArticles"
            />
          </div>
          <button class="btn btn-ghost btn-icon" @click="fetchArticles" title="刷新">
            🔄
          </button>
        </div>
      </header>

      <!-- Content -->
      <div class="content">
        <!-- Recommendations View -->
        <div v-if="showRecommendations" class="recommendations-view">
          <h2 class="recommendations-title">推荐订阅源</h2>
          <p class="recommendations-subtitle">发现优质内容源，开始你的阅读之旅</p>
          
          <div class="recommendations-grid">
            <div v-for="rec in recommendedFeeds" :key="rec.url" class="recommendation-card">
              <div class="recommendation-header">
                <span class="recommendation-icon">{{ rec.icon }}</span>
                <div class="recommendation-info">
                  <h3 class="recommendation-name">{{ rec.name }}</h3>
                  <span class="recommendation-category">{{ rec.category }}</span>
                </div>
              </div>
              <p class="recommendation-desc">{{ rec.description }}</p>
              <button 
                class="btn btn-primary recommendation-btn" 
                @click="addRecommendedFeed(rec)"
                :disabled="rec.adding"
              >
                {{ rec.added ? '已添加 ✓' : (rec.adding ? '添加中...' : '+ 订阅') }}
              </button>
            </div>
          </div>
        </div>
        
        <!-- Articles View (default) -->
        <template v-else>
        <div v-if="loading" class="empty-state">
          <div class="empty-state-icon">⏳</div>
          <div class="empty-state-title">加载中...</div>
        </div>

        <div v-else-if="articles.length === 0" class="empty-state">
          <div class="empty-state-icon">📭</div>
          <div class="empty-state-title">暂无文章</div>
          <div class="empty-state-description">点击「添加订阅源」开始订阅你喜欢的 RSS</div>
        </div>

        <div v-else>
          <!-- Article List -->
          <div class="article-list">
            <div 
              v-for="article in articles" 
              :key="article.id" 
              class="article-list-item"
              :class="{ unread: !article.is_read }"
            >
              <div class="article-list-content" @click="openArticle(article)">
                <div class="article-list-meta">
                  <span class="article-list-source" v-if="article.feed">
                    {{ article.feed.title || article.feed.url }}
                  </span>
                  <span class="article-list-time">
                    {{ formatDate(article.pub_date) }}
                    <span v-if="!article.is_read" class="unread-dot"></span>
                  </span>
                </div>
                <h3 class="article-list-title">{{ article.title }}</h3>
                <p class="article-list-description" v-html="sanitizeHtml(article.description)"></p>
                <!-- Article Tags -->
                <div v-if="article.tags && article.tags.length > 0" class="article-tags">
                  <span v-for="tag in article.tags" :key="tag.id" class="article-tag">{{ tag.name }}</span>
                </div>
                <!-- AI Summary Preview -->
                <div v-if="article.summary" class="article-summary-preview" @click.stop="article.showSummary = !article.showSummary">
                  <div class="summary-preview-divider"></div>
                  <span class="summary-preview-icon">🤖</span>
                  <span class="summary-preview-text">AI概览：<template v-if="article.showSummary">{{ article.summary }}</template><template v-else>{{ article.summary.slice(0, 50) }}{{ article.summary.length > 50 ? '...' : '' }}</template></span>
                  <div v-if="getKeyPoints(article).length && article.showSummary" class="summary-keypoints">
                    <span v-for="(point, idx) in getKeyPoints(article)" :key="idx" class="keypoint-tag">{{ point }}</span>
                  </div>
                </div>
              </div>
              <div class="article-actions">
                <button class="btn btn-ghost btn-sm" @click="showTagArticle(article)" title="添加标签">🏷️</button>
                <button class="btn btn-primary btn-sm" @click="generateSummary(article)" :disabled="article.summaryLoading">{{ article.summaryLoading ? "生成中..." : "AI概览" }}</button>
              </div>

            </div>
          </div>

          <!-- Pagination -->
          <div class="pagination" v-if="total > perPage">
            <button class="pagination-btn" :disabled="page === 1" @click="prevPage">
              上一页
            </button>
            <span class="pagination-info">第 {{ page }} 页 / 共 {{ totalPages }} 页</span>
            <button class="pagination-btn" :disabled="page >= totalPages" @click="nextPage">
              下一页
            </button>
          </div>
        </div>
        </template>
      </div>
    </main>

    <!-- Modals... (simplified for space) -->
    <div v-if="showAddFeed" class="modal-overlay" @click.self="showAddFeed = false">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">添加订阅源</h3>
        </div>
        <div class="modal-body">
          <div v-if="feedError" class="message message-error">{{ feedError }}</div>
          <div class="form-group">
            <label class="form-label">RSS 地址</label>
            <input v-model="newFeed.url" class="form-input" placeholder="https://example.com/feed.xml" />
          </div>
          <div class="form-group">
            <label class="form-label">标题（可选）</label>
            <input v-model="newFeed.title" class="form-input" placeholder="自定义标题" />
          </div>
          <div class="form-group">
            <label class="form-label">分类（可选）</label>
            <input v-model="newFeed.category" class="form-input" placeholder="如：科技、新闻" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showAddFeed = false">取消</button>
          <button class="btn btn-primary" @click="createFeed" :disabled="feedLoading">
            {{ feedLoading ? '添加中...' : '添加' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showAddTag" class="modal-overlay" @click.self="showAddTag = false">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">新建标签</h3>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">标签名称</label>
            <input v-model="newTagName" class="form-input" placeholder="如：重要、待读" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showAddTag = false">取消</button>
          <button class="btn btn-primary" @click="createTag">创建</button>
        </div>
      </div>
    </div>

    <div v-if="showTagArticleModal" class="modal-overlay" @click.self="showTagArticleModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">添加标签</h3>
        </div>
        <div class="modal-body">
          <p style="margin-bottom: 16px; color: var(--text-secondary);">
            为文章「{{ selectedArticle?.title }}」添加标签
          </p>
          <div class="form-group">
            <label class="form-label">选择已有标签</label>
            <select v-model="selectedTagForArticle" class="form-input">
              <option value="">请选择</option>
              <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">或创建新标签</label>
            <div style="display: flex; gap: 8px;">
              <input v-model="newTagForArticle" class="form-input" placeholder="新标签名称" />
              <button class="btn btn-secondary" @click="createAndAddTag">创建并添加</button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showTagArticleModal = false">关闭</button>
          <button class="btn btn-primary" @click="addTagToArticle" :disabled="!selectedTagForArticle">添加</button>
        </div>
      </div>
    </div>

    <div v-if="showProfile" class="modal-overlay" @click.self="showProfile = false">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">个人信息</h3>
        </div>
        <div class="modal-body">
          <div class="profile-section">
            <div class="profile-avatar">{{ userEmail?.charAt(0).toUpperCase() }}</div>
            <div class="profile-info">
              <div class="profile-label">邮箱</div>
              <div class="profile-value">{{ userEmail }}</div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showProfile = false">关闭</button>
        </div>
      </div>
    </div>

    <div v-if="showFeedManager" class="modal-overlay" @click.self="showFeedManager = false">
      <div class="modal" style="max-width: 600px;">
        <div class="modal-header">
          <h3 class="modal-title">订阅源管理</h3>
        </div>
        <div class="modal-body" style="max-height: 400px; overflow-y: auto;">
          <div v-if="feeds.length === 0" style="text-align: center; padding: 40px; color: var(--text-muted);">
            暂无订阅源
          </div>
          <div v-else class="feed-manager-list">
            <div class="feed-manager-header">
              <label class="feed-checkbox">
                <input type="checkbox" @change="toggleSelectAllFeeds" :checked="isAllFeedsSelected" />
                <span>全选</span>
              </label>
              <span class="feed-count-text">已选 {{ selectedFeedIds.length }} / {{ feeds.length }}</span>
            </div>
            <div v-for="feed in feeds" :key="feed.id" class="feed-manager-item" :class="{ selected: selectedFeedIds.includes(feed.id) }">
              <label class="feed-checkbox">
                <input type="checkbox" :value="feed.id" v-model="selectedFeedIds" />
              </label>
              <div class="feed-manager-info" @click="toggleFeedSelection(feed.id)">
                <img v-if="feed.icon_url" :src="feed.icon_url" class="feed-icon-img" @error="$event.target.style.display='none'" /><div v-else class="feed-manager-icon">{{ getFeedIcon(feed.category) }}</div>
                <div class="feed-manager-details">
                  <div class="feed-manager-title">{{ feed.title || feed.url }}</div>
                  <div class="feed-manager-meta">
                    <span v-if="feed.category">{{ feed.category }}</span>
                    <span>{{ feed.url }}</span>
                  </div>
                </div>
              </div>
              <div class="feed-manager-actions">
                <button class="btn btn-ghost btn-sm" @click="editFeed(feed)" title="编辑">✏️</button>
                <button class="btn btn-ghost btn-sm" @click="deleteFeed(feed.id)" title="删除">🗑️</button>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="$refs.opmlInput.click()">📥 导入 OPML</button>
          <button class="btn btn-secondary" @click="exportFeeds" :disabled="selectedFeedIds.length === 0">📤 导出 OPML ({{ selectedFeedIds.length }})</button>
          <input ref="opmlInput" type="file" accept=".opml,.xml" @change="importFeeds" style="display: none" />
          <button class="btn btn-secondary" @click="showFeedManager = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 编辑订阅源弹窗 -->
    <div v-if="showEditFeed" class="modal-overlay" @click.self="showEditFeed = false">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">编辑订阅源</h3>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">标题</label>
            <input v-model="editingFeed.title" class="form-input" placeholder="自定义标题" />
          </div>
          <div class="form-group">
            <label class="form-label">分类</label>
            <input v-model="editingFeed.category" class="form-input" placeholder="如：科技、新闻" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showEditFeed = false">取消</button>
          <button class="btn btn-primary" @click="updateFeed" :disabled="editFeedLoading">
            {{ editFeedLoading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>



    <!-- Auth Modal -->
    <AuthModal v-model="showAuthModal" @success="onAuthSuccess" />
    
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import AuthModal from '../components/AuthModal.vue'
import { useRouter } from 'vue-router'
import api, { markArticleRead, markAllRead, getUnreadCount, generateSummary as apiGenerateSummary } from '../api'

const router = useRouter()
const authStore = useAuthStore()
const theme = ref(localStorage.getItem('theme') || 'light')
const sidebarWidth = ref(260)
const feeds = ref([])
const articles = ref([])
const tags = ref([])
const loading = ref(false)
const showAddFeed = ref(false)
const showAddTag = ref(false)
const showProfile = ref(false)
const showFeedManager = ref(false)
const showEditFeed = ref(false)
const showUserMenu = ref(false)
const showTagArticleModal = ref(false)
const newFeed = ref({ url: '', title: '', category: '' })
const editingFeed = ref({ id: 0, title: '', category: '' })
const selectedFeedIds = ref([])
const feedError = ref('')
const feedLoading = ref(false)
const editFeedLoading = ref(false)
const newTagName = ref('')
const selectedArticle = ref(null)
const selectedTagForArticle = ref('')
const newTagForArticle = ref('')
const searchQuery = ref('')
const selectedFeedId = ref(0)
const selectedTagId = ref(0)
const selectedCategories = ref([])  // Multi-select categories
const showCategoryDropdown = ref(false)
const showReadFilterDropdown = ref(false)
const page = ref(1)
const perPage = ref(20)
const total = ref(0)
const readFilter = ref('')
const showRecommendations = ref(false)
const showAuthModal = ref(false)
const pendingAction = ref(null)

// Recommended feeds data
const recommendedFeeds = ref([
  { name: '36氪', url: 'https://36kr.com/feed', category: '科技', icon: '🚀', description: '专注全球互联网创业与投资的科技媒体', adding: false, added: false },
  { name: '阮一峰的网络日志', url: 'https://www.ruanyifeng.com/blog/atom.xml', category: '博客', icon: '📝', description: '知名技术博主的个人博客', adding: false, added: false },
  { name: 'V2EX', url: 'https://www.v2ex.com/index.xml', category: '社区', icon: '🌐', description: '创意工作者的社区', adding: false, added: false },
  { name: '少数派', url: 'https://sspai.com/feed', category: '效率', icon: '📱', description: '高效工作与高品质生活', adding: false, added: false },
  { name: '虎嗅', url: 'https://www.huxiu.com/rss/0.xml', category: '商业', icon: '🐯', description: '有视角的商业资讯', adding: false, added: false },
  { name: 'Wired', url: 'https://www.wired.com/feed/rss', category: '科技', icon: '🔌', description: '科技如何改变文化、经济和政治', adding: false, added: false },
  { name: 'TechCrunch', url: 'https://techcrunch.com/feed/', category: '科技', icon: '🚀', description: '科技创业与投资新闻', adding: false, added: false },
  { name: 'BBC 中文网', url: 'https://feeds.bbci.co.uk/zhongwen/simp/rss.xml', category: '新闻', icon: '🌍', description: 'BBC 中文新闻', adding: false, added: false },
  { name: 'MIT Technology Review', url: 'https://www.technologyreview.com/feed/', category: '科技', icon: '🔬', description: 'MIT 科技评论', adding: false, added: false },
  { name: 'Ars Technica', url: 'https://feeds.arstechnica.com/arstechnica/index', category: '科技', icon: '⚡', description: '深度科技新闻与分析', adding: false, added: false },
  { name: 'The Verge', url: 'https://www.theverge.com/rss/index.xml', category: '科技', icon: '📱', description: '科技、科学与艺术', adding: false, added: false },
  { name: 'Hacker News', url: 'https://hnrss.org/frontpage', category: '技术', icon: '🔥', description: '热门技术讨论', adding: false, added: false },
])

// Go to home/recommendations
function goHome() {
  showRecommendations.value = true
  selectedFeedId.value = 0
  selectedTagId.value = 0
  selectedCategories.value = []
}

function closeAllDropdowns() {
  showCategoryDropdown.value = false
  showReadFilterDropdown.value = false
}

function clearCategories() {
  selectedCategories.value = []
  onCategoryChange()
}

function toggleCategory(cat) {
  const idx = selectedCategories.value.indexOf(cat)
  if (idx === -1) {
    selectedCategories.value.push(cat)
  } else {
    selectedCategories.value.splice(idx, 1)
  }
  onCategoryChange()
}

function setReadFilter(value) {
  readFilter.value = value
  showReadFilterDropdown.value = false
  fetchArticles()
}

// Add recommended feed
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
  userEmail.value = localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')).email : ''
  if (pendingAction.value) {
    pendingAction.value()
    pendingAction.value = null
  }
}

const analyzedCount = ref(0)
const unreadCount = ref({ total: 0, by_feed: {}, by_category: {} })
const userEmail = ref(localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')).email : '')

const categories = computed(() => {
  const cats = new Set()
  feeds.value.forEach(f => { if (f.category) cats.add(f.category) })
  return Array.from(cats).sort()
})

const categoryFilterLabel = computed(() => {
  if (selectedCategories.value.length === 0) return '全部分类'
  if (selectedCategories.value.length === 1) return selectedCategories.value[0]
  // 显示前2个分类名称，超过则显示 "+N"
  const firstTwo = selectedCategories.value.slice(0, 2).join(', ')
  const remaining = selectedCategories.value.length - 2
  if (remaining > 0) return `${firstTwo} +${remaining}`
  return firstTwo
})

const readFilterLabel = computed(() => {
  const labels = { '': '全部', 'unread': '未读', 'read': '已读', 'analyzed': '已分析' }
  return labels[readFilter.value] || '全部'
})

const filteredFeeds = computed(() => {
  if (selectedCategories.value.length === 0) return feeds.value
  return feeds.value.filter(f => selectedCategories.value.includes(f.category))
})

const totalPages = computed(() => Math.ceil(total.value / perPage.value))

const currentTitle = computed(() => {
  if (selectedTagId.value > 0) {
    const tag = tags.value.find(t => t.id === selectedTagId.value)
    return tag ? '标签：' + tag.name : '全部文章'
  }
  if (selectedFeedId.value > 0) {
    const feed = feeds.value.find(f => f.id === selectedFeedId.value)
    return feed ? (feed.title || feed.url) : '全部文章'
  }
  if (selectedCategories.value.length === 1) return '分类：' + selectedCategories.value[0]
  if (selectedCategories.value.length === 2) return '分类：' + selectedCategories.value.join(', ')
  if (selectedCategories.value.length > 2) return '分类：' + selectedCategories.value.slice(0, 2).join(', ') + ' +' + (selectedCategories.value.length - 2)
  return '全部文章'
})

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', theme.value)
}

function getFeedIcon(category) {
  const icons = { 'AI': '🤖', 'Safety': '🛡️', '科技': '💻', '新闻': '📰', '博客': '📝', '设计': '🎨', '财经': '💰' }
  return icons[category] || '📄'
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now - date
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days === 0) {
    const hours = Math.floor(diff / (1000 * 60 * 60))
    if (hours === 0) {
      const mins = Math.floor(diff / (1000 * 60))
      return mins <= 1 ? '刚刚' : mins + ' 分钟前'
    }
    return hours + ' 小时前'
  } else if (days === 1) return '昨天'
  else if (days < 7) return days + ' 天前'
  return date.toLocaleDateString('zh-CN')
}

const isResizing = ref(false)

function startResize(e) {
  isResizing.value = true
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

function onResize(e) {
  if (!isResizing.value) return
  const newWidth = e.clientX
  if (newWidth >= 180 && newWidth <= 500) sidebarWidth.value = newWidth
}

function stopResize() {
  isResizing.value = false
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
}

async function fetchFeeds() {
  try { feeds.value = (await api.get('/feeds')).data || [] } catch (e) { console.error(e) }
}

async function createFeed() {
  if (!newFeed.value.url) { feedError.value = '请输入 RSS 地址'; return }
  feedLoading.value = true
  feedError.value = ''
  try {
    await api.post('/feeds', newFeed.value)
    showAddFeed.value = false
    newFeed.value = { url: '', title: '', category: '' }
    fetchFeeds()
  } catch (e) { feedError.value = e.response?.data?.error || '添加失败' }
  finally { feedLoading.value = false }
}

async function deleteFeed(id) {
  if (!confirm('确定删除该订阅源？')) return
  try {
    await api.delete('/feeds/' + id)
    fetchFeeds()
    if (selectedFeedId.value === id) { selectedFeedId.value = 0; fetchArticles() }
  } catch (e) { console.error(e) }
}


function editFeed(feed) {
  editingFeed.value = { id: feed.id, title: feed.title || '', category: feed.category || '' }
  showEditFeed.value = true
}

async function updateFeed() {
  editFeedLoading.value = true
  try {
    await api.put('/feeds/' + editingFeed.value.id, {
      title: editingFeed.value.title,
      category: editingFeed.value.category
    })
    showEditFeed.value = false
    fetchFeeds()
  } catch (e) {
    console.error(e)
    alert('保存失败')
  } finally {
    editFeedLoading.value = false
  }
}

async function importFeeds(event) {
  const file = event.target.files[0]
  if (!file) return
  
  const text = await file.text()
  const parser = new DOMParser()
  const doc = parser.parseFromString(text, 'text/xml')
  const outlines = doc.querySelectorAll('outline[xmlUrl]')
  
  let imported = 0
  let failed = 0
  
  for (const outline of outlines) {
    const url = outline.getAttribute('xmlUrl')
    const title = outline.getAttribute('title') || outline.getAttribute('text') || ''
    const category = outline.getAttribute('category') || ''
    
    try {
      await api.post('/feeds', { url, title, category })
      imported++
    } catch (e) {
      failed++
      console.error('Failed to import:', url, e)
    }
  }
  
  event.target.value = ''
  fetchFeeds()
  alert(`导入完成：成功 ${imported} 个，失败 ${failed} 个`)
}

const isAllFeedsSelected = computed(() => selectedFeedIds.value.length === feeds.value.length && feeds.value.length > 0)

function toggleSelectAllFeeds(e) {
  if (e.target.checked) {
    selectedFeedIds.value = feeds.value.map(f => f.id)
  } else {
    selectedFeedIds.value = []
  }
}

function toggleFeedSelection(id) {
  const idx = selectedFeedIds.value.indexOf(id)
  if (idx === -1) {
    selectedFeedIds.value.push(id)
  } else {
    selectedFeedIds.value.splice(idx, 1)
  }
}

function exportFeeds() {
  let opml = '<?xml version="1.0" encoding="UTF-8"?>\n<opml version="2.0">\n  <head>\n    <title>RSS Feeds Export</title>\n  </head>\n  <body>\n'
  feeds.value.filter(f => selectedFeedIds.value.includes(f.id)).forEach(feed => {
    const title = feed.title || feed.url
    const category = feed.category || 'Uncategorized'
    opml += `    <outline type="rss" text="${escapeXml(title)}" title="${escapeXml(title)}" xmlUrl="${escapeXml(feed.url)}" category="${escapeXml(category)}"/>\n`
  })
  opml += '  </body>\n</opml>'
  const blob = new Blob([opml], { type: 'application/xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'rss-feeds.opml'
  a.click()
  URL.revokeObjectURL(url)
}

function escapeXml(str) {
  if (!str) return ''
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&apos;')
}

async function fetchArticles() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: perPage.value }
    if (selectedFeedId.value > 0) params.feed_id = selectedFeedId.value
    if (selectedTagId.value > 0) params.tag_id = selectedTagId.value
    if (selectedCategories.value.length > 0) params.categories = selectedCategories.value.join(',')
    if (readFilter.value === 'unread') params.is_read = 'false'
    if (readFilter.value === 'read') params.is_read = 'true'
    if (readFilter.value === 'analyzed') params.has_summary = 'true'
    const res = await api.get('/articles', { params })
    articles.value = res.data.articles || []
    total.value = res.data.total || 0
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function searchArticles() {
  if (!searchQuery.value) { fetchArticles(); return }
  loading.value = true
  try {
    const res = await api.get('/articles/search', { params: { q: searchQuery.value } })
    articles.value = res.data.articles || []
    total.value = res.data.total || 0
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function openArticle(article) {
  if (article.link) window.open(article.link, '_blank')
  if (!article.is_read) {
    try {
      await markArticleRead(article.id)
      article.is_read = true
      fetchUnreadCountData()
    } catch (e) { console.error(e) }
  }
}

function getKeyPoints(article) {
  const points = article.keyPoints || article.key_points
  if (Array.isArray(points)) return points
  if (typeof points === 'string') {
    try { return JSON.parse(points) } catch { return [] }
  }
  return []
}

async function generateSummary(article) {
  if (article.summary) {
    article.showSummary = !article.showSummary
    return
  }
  article.summaryLoading = true
  try {
    const res = await apiGenerateSummary(article.id)
    article.summary = res.data.summary
    article.keyPoints = res.data.key_points
    article.showSummary = true
  } catch (e) {
    console.error("Failed to generate summary:", e)
    alert(e.response?.data?.error || "生成概览失败")
  } finally {
    article.summaryLoading = false
  }
}

async function markAllAsRead(feedId, categories) {
  const category = Array.isArray(categories) ? categories.join(',') : categories
  try {
    await markAllRead(feedId, category)
    fetchArticles()
    fetchUnreadCountData()
  } catch (e) { console.error(e) }
}

async function fetchUnreadCountData() {
  try { unreadCount.value = (await getUnreadCount()).data } catch (e) { console.error(e) }
}

function selectFeed(id) { selectedFeedId.value = id; selectedTagId.value = 0; page.value = 1; showRecommendations.value = false; fetchArticles() }
function selectTag(id) { selectedTagId.value = id; selectedFeedId.value = 0; page.value = 1; showRecommendations.value = false; fetchArticles() }
function onCategoryChange() { selectedFeedId.value = 0; selectedTagId.value = 0; page.value = 1; showRecommendations.value = false; fetchArticles() }
function prevPage() { if (page.value > 1) { page.value--; fetchArticles() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; fetchArticles() } }

async function fetchTags() {
  try { tags.value = (await api.get('/tags')).data || [] } catch (e) { console.error(e) }
}

async function createTag() {
  if (!newTagName.value) return
  try {
    await api.post('/tags', { name: newTagName.value })
    newTagName.value = ''
    showAddTag.value = false
    fetchTags()
  } catch (e) { console.error(e) }
}

async function deleteTag(id) {
  if (!confirm('确定删除该标签？')) return
  try { await api.delete('/tags/' + id); fetchTags() } catch (e) { console.error(e) }
}

function showTagArticle(article) {
  selectedArticle.value = article
  selectedTagForArticle.value = ''
  newTagForArticle.value = ''
  showTagArticleModal.value = true
}

async function addTagToArticle() {
  if (!selectedTagForArticle.value || !selectedArticle.value) return
  try {
    await api.post('/articles/tags', { article_id: selectedArticle.value.id, tag_id: selectedTagForArticle.value })
    showTagArticleModal.value = false
    fetchArticles()
  } catch (e) { console.error(e) }
}

async function createAndAddTag() {
  if (!newTagForArticle.value || !selectedArticle.value) return
  try {
    const res = await api.post('/tags', { name: newTagForArticle.value })
    await api.post('/articles/tags', { article_id: selectedArticle.value.id, tag_id: res.data.id })
    showTagArticleModal.value = false
    fetchTags()
    fetchArticles()
  } catch (e) { console.error(e) }
}

function logout() {
  authStore.logout()
  showUserMenu.value = false
  // 留在首页，不跳转
}


// 安全渲染 HTML，只允许无风险的标签
function sanitizeHtml(html) {
  if (!html) return ""
  const allowedTags = ["a", "p", "br", "strong", "b", "em", "i", "u", "span", "div", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote", "code", "pre"]
  const allowedAttrs = { a: ["href", "title"], span: ["class"], div: ["class"], p: ["class"] }
  let result = html.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, "")
  result = result.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, "")
  result = result.replace(/\s*on\w+="[^"]*"/gi, "").replace(/\s*on\w+='[^']*'/gi, "")
  result = result.replace(/javascript:/gi, "")
  const temp = document.createElement("div")
  temp.innerHTML = result
  function cleanNode(node) {
    if (node.nodeType === Node.ELEMENT_NODE) {
      const tagName = node.tagName.toLowerCase()
      if (!allowedTags.includes(tagName)) {
        while (node.firstChild) node.parentNode.insertBefore(node.firstChild, node)
        node.parentNode.removeChild(node)
        return
      }
      const allowed = allowedAttrs[tagName] || []
      Array.from(node.attributes).forEach(attr => {
        if (!allowed.includes(attr.name)) node.removeAttribute(attr.name)
      })
    }
    Array.from(node.childNodes).forEach(cleanNode)
  }
  Array.from(temp.childNodes).forEach(cleanNode)
  return temp.innerHTML
}
onMounted(() => {
  // 根据登录状态决定初始视图
  if (!authStore.isLoggedIn) {
    showRecommendations.value = true  // 未登录显示推荐页
  } else {
    showRecommendations.value = false  // 已登录显示全部文章
    fetchArticles()
  }
  fetchFeeds()
  fetchTags()
  fetchUnreadCountData()
  
  // 点击外部区域关闭下拉菜单
  document.addEventListener('click', (e) => {
    const target = e.target
    const categoryDropdown = document.querySelector('.category-filter .dropdown-wrapper')
    const readFilterDropdown = document.querySelector('.header-right .dropdown-wrapper')
    
    if (categoryDropdown && !categoryDropdown.contains(target)) {
      showCategoryDropdown.value = false
    }
    if (readFilterDropdown && !readFilterDropdown.contains(target)) {
      showReadFilterDropdown.value = false
    }
  })
})
</script>

<style scoped>
.article-tags { margin-top: 8px; display: flex; gap: 6px; flex-wrap: wrap; }
.article-tag { font-size: 11px; padding: 2px 8px; background: var(--bg-tertiary); color: var(--text-secondary); border-radius: var(--radius-full); }
.article-actions { display: flex; flex-direction: column; gap: 4px; padding-left: 12px; border-left: 1px solid var(--border); }
.btn-sm { padding: 4px 8px; font-size: 12px; }
.btn-compact { padding: 4px 10px; font-size: 12px; height: 28px; line-height: 1; display: inline-flex; align-items: center; white-space: nowrap; }
.category-filter { margin-bottom: 16px; }
.category-filter select { font-size: 13px; }
.user-section { padding: 12px 16px; border-top: 1px solid var(--border); position: relative; }
.user-info { display: flex; align-items: center; gap: 12px; padding: 8px 12px; border-radius: var(--radius-md); cursor: pointer; }
.user-info:hover { background: var(--bg-hover); }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; background: var(--primary); color: white; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 16px; }
.user-details { flex: 1; min-width: 0; }
.user-email { font-size: 14px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--text-primary); }
.user-arrow { font-size: 10px; color: var(--text-muted); transition: transform var(--transition-fast); }
.user-arrow.rotated { transform: rotate(180deg); }
.user-menu { position: absolute; bottom: 100%; left: 16px; right: 16px; background: var(--bg-primary); border: 1px solid var(--border); border-radius: var(--radius-md); box-shadow: var(--shadow-lg); margin-bottom: 8px; overflow: hidden; }
.menu-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; cursor: pointer; }
.menu-item:hover { background: var(--bg-hover); }
.menu-item-danger { color: var(--danger); }
.menu-item-danger:hover { background: rgba(239, 68, 68, 0.1); }
.menu-divider { height: 1px; background: var(--border); }
.menu-backdrop { position: fixed; inset: 0; z-index: 50; }
.profile-section { text-align: center; }
.profile-avatar { width: 80px; height: 80px; border-radius: 50%; background: var(--primary); color: white; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 32px; margin: 0 auto 24px; }
.profile-info { padding: 12px 0; border-bottom: 1px solid var(--border); }
.profile-label { font-size: 12px; color: var(--text-muted); margin-bottom: 4px; }
.profile-value { font-size: 16px; font-weight: 500; }
.feed-manager-list { display: flex; flex-direction: column; gap: 8px; }
.feed-manager-header { display: flex; align-items: center; justify-content: space-between; padding: 8px 0; margin-bottom: 8px; border-bottom: 1px solid var(--border); }
.feed-checkbox { display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 14px; }
.feed-checkbox input { width: 16px; height: 16px; cursor: pointer; }
.feed-count-text { font-size: 13px; color: var(--text-muted); }
.feed-manager-item.selected { background: var(--primary-light); }
.feed-manager-header { display: flex; align-items: center; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid var(--border); margin-bottom: 8px; }
.feed-checkbox { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.feed-checkbox input { width: 16px; height: 16px; cursor: pointer; }
.feed-count-text { font-size: 12px; color: var(--text-muted); }
.feed-manager-item.selected { background: var(--primary-light); }
.feed-manager-item { display: flex; align-items: center; padding: 12px; background: var(--bg-secondary); border-radius: var(--radius-md); gap: 12px; }
.feed-manager-info { flex: 1; display: flex; align-items: center; gap: 12px; cursor: pointer; min-width: 0; }
.feed-manager-icon { font-size: 20px; flex-shrink: 0; }
.feed-manager-details { flex: 1; min-width: 0; }
.feed-manager-title { font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.feed-manager-meta { font-size: 12px; color: var(--text-muted); display: flex; gap: 8px; margin-top: 4px; }
.feed-manager-actions { display: flex; gap: 4px; }
.article-list-item.unread .article-list-title { font-weight: 600; }

/* 未读橙色圆点 */
.unread-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #f97316;
  border-radius: 50%;
  margin-left: 6px;
  vertical-align: middle;
  margin-top: 2px;
}
.article-summary-preview { margin-top: 8px; cursor: pointer; }
.article-summary-preview:hover { opacity: 0.8; }
.summary-preview-divider { height: 1px; background: var(--border); margin-bottom: 8px; }
.summary-preview-icon { margin-right: 4px; }
.summary-preview-text { font-size: 12px; color: #10b981; line-height: 1.4; }
.summary-keypoints { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 6px; }
.keypoint-tag { font-size: 11px; color: #f97316; background: rgba(249, 115, 22, 0.1); border: 1px solid rgba(249, 115, 22, 0.3); padding: 2px 8px; border-radius: 4px; }
.article-summary-expand { margin-top: 8px; flex-basis: 100%; }
.summary-expand-inner { }
.summary-loading { display: flex; align-items: center; gap: 8px; padding: 8px 0; color: var(--text-muted); }
.loading-spinner { width: 14px; height: 14px; border: 2px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.read-filter-select { padding: 6px 10px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg-secondary); color: var(--text-primary); font-size: 13px; cursor: pointer; }
.read-filter-select:focus { outline: none; border-color: var(--primary); }
.toggle-switch { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; }
.toggle-switch input { width: 16px; height: 16px; cursor: pointer; }
.toggle-label { color: var(--text-secondary); }
.header-count { font-size: 14px; font-weight: normal; color: var(--text-muted); margin-left: 8px; }
.resizer { width: 5px; background: transparent; cursor: col-resize; position: fixed; top: 0; height: 100vh; z-index: 101; }
.resizer:hover { background: var(--primary); }
/* AI Summary */
.article-summary { margin-top: 12px; padding: 12px; background: var(--bg-tertiary); border-radius: var(--radius-md); border-left: 3px solid var(--primary); }
.summary-header { font-size: 12px; font-weight: 600; color: var(--primary); margin-bottom: 8px; }
.summary-text { font-size: 13px; line-height: 1.5; color: var(--text-primary); margin-bottom: 8px; }
.summary-points { margin: 0; padding-left: 20px; }
.summary-points li { font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }

/* Home Button */
.home-button {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 12px;
  background: transparent;
  border: none;
  width: 100%;
  text-align: left;
}

.home-button:hover {
  background: var(--bg-hover);
}

.home-button.active {
  background: var(--primary);
  color: white;
}

.home-icon {
  font-size: 18px;
}

.home-text {
  font-size: 14px;
  font-weight: 500;
}

/* Recommendations View */
.recommendations-view {
  padding: 20px 0;
}

.recommendations-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.recommendations-subtitle {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 24px;
}

.recommendations-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.recommendation-card {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
  transition: box-shadow var(--transition-normal), transform var(--transition-normal);
}

.recommendation-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.recommendation-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.recommendation-icon {
  font-size: 28px;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.recommendation-info {
  flex: 1;
}

.recommendation-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px 0;
}

.recommendation-category {
  font-size: 12px;
  color: var(--text-muted);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.recommendation-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 16px;
  min-height: 40px;
}

.recommendation-btn {
  width: 100%;
}

/* 添加订阅源按钮 - 简洁加号 */
.btn-add-feed {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.btn-add-feed:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* Login Link */
.login-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  width: 100%;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
}

.login-link:hover {
  background: var(--bg-hover);
}

.login-icon {
  font-size: 18px;
}

.login-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

/* Dropdown Styles */
.dropdown-wrapper {
  position: relative;
  z-index: 60;
}
.dropdown-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.dropdown-trigger:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}
.dropdown-trigger-sm {
  padding: 6px 10px;
  width: auto;
  min-width: 80px;
}
.dropdown-arrow {
  font-size: 10px;
  color: var(--text-muted);
  transition: transform 0.2s;
}
.dropdown-arrow.open {
  transform: rotate(180deg);
}
.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 150px;
  max-height: 280px;
  overflow-y: auto;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  z-index: 100;
  padding: 4px 0;
}
.dropdown-menu-right {
  left: auto;
  right: 0;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  transition: background 0.15s;
}
.dropdown-item:hover {
  background: var(--bg-hover);
}
.dropdown-item.active {
  color: var(--primary);
  font-weight: 500;
}
.dropdown-checkbox {
  width: 16px;
  height: 16px;
  border: 1px solid var(--border);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  color: transparent;
  transition: all 0.15s;
}
.dropdown-checkbox.checked {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}
.dropdown-divider {
  height: 1px;
  background: var(--border);
  margin: 4px 0;
}
</style>
