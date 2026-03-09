# 文章已读/未读状态功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 RSS Reader 添加文章已读/未读状态追踪功能

**Architecture:** 后端添加标记已读 API 和批量操作 API，修改文章列表查询支持未读筛选；前端在点击文章时调用标记已读 API，添加筛选开关和批量操作按钮。

**Tech Stack:** Go (Gin + GORM), Vue 3, PostgreSQL

---

## Task 1: 后端 - 添加 Repository 方法

**Files:**
- Modify: `/home/prj/rss-reader/internal/repository/article.go`

**Step 1: 添加 MarkAsRead 方法**

在 `ArticleRepository` 中添加：

```go
// MarkAsRead 标记单篇文章为已读
func (r *ArticleRepository) MarkAsRead(userID, articleID uint) error {
	return r.db.Model(&models.Article{}).
		Where("id = ? AND user_id = ?", articleID, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead 批量标记已读，支持按 feed_id 和 category 筛选
func (r *ArticleRepository) MarkAllAsRead(userID uint, feedID uint, category string) (int64, error) {
	query := r.db.Model(&models.Article{}).Where("user_id = ? AND is_read = ?", userID, false)
	
	if feedID > 0 {
		query = query.Where("feed_id = ?", feedID)
	}
	
	if category != "" {
		query = query.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.category = ?", category)
	}
	
	result := query.Update("is_read", true)
	return result.RowsAffected, result.Error
}

// GetUnreadCount 获取未读数量统计
func (r *ArticleRepository) GetUnreadCount(userID uint) (int64, map[uint]int64, map[string]int64, error) {
	var total int64
	r.db.Model(&models.Article{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&total)
	
	// By feed
	byFeed := make(map[uint]int64)
	var feedCounts []struct {
		FeedID uint
		Count  int64
	}
	r.db.Model(&models.Article{}).
		Select("feed_id, count(*) as count").
		Where("user_id = ? AND is_read = ?", userID, false).
		Group("feed_id").
		Scan(&feedCounts)
	for _, fc := range feedCounts {
		byFeed[fc.FeedID] = fc.Count
	}
	
	// By category
	byCategory := make(map[string]int64)
	var catCounts []struct {
		Category string
		Count    int64
	}
	r.db.Model(&models.Article{}).
		Select("feeds.category, count(*) as count").
		Joins("JOIN feeds ON feeds.id = articles.feed_id").
		Where("articles.user_id = ? AND articles.is_read = ?", userID, false).
		Group("feeds.category").
		Scan(&catCounts)
	for _, cc := range catCounts {
		byCategory[cc.Category] = cc.Count
	}
	
	return total, byFeed, byCategory, nil
}
```

**Step 2: 修改 FindByUserID 支持 is_read 筛选**

修改 `FindByUserID` 方法签名，添加 `isRead *bool` 参数：

```go
func (r *ArticleRepository) FindByUserID(userID uint, page, pageSize int, feedID uint, tagID uint, category string, isRead *bool) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).Where("articles.user_id = ?", userID)
	
	if feedID > 0 {
		query = query.Where("articles.feed_id = ?", feedID)
	}
	
	if tagID > 0 {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tagID)
	}
	
	if category != "" {
		query = query.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.category = ?", category)
	}
	
	if isRead != nil {
		query = query.Where("articles.is_read = ?", *isRead)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("articles.pub_date DESC").Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	for i := range articles {
		var feed models.Feed
		if err := r.db.First(&feed, articles[i].FeedID).Error; err == nil {
			articles[i].Feed = &feed
		}
		r.db.Model(&articles[i]).Association("Tags").Find(&articles[i].Tags)
	}

	return articles, total, nil
}
```

**Step 3: 提交**

```bash
cd /home/prj/rss-reader && git add internal/repository/article.go && git commit -m "feat: add read status repository methods"
```

---

## Task 2: 后端 - 添加 Handlers

**Files:**
- Modify: `/home/prj/rss-reader/internal/handlers/article.go`

**Step 1: 添加 MarkArticleRead handler**

```go
func MarkArticleRead(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		if err := articleRepo.MarkAsRead(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article marked as read"})
	}
}
```

**Step 2: 添加 MarkAllRead handler**

```go
func MarkAllRead(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var req struct {
			FeedID   uint   `json:"feed_id"`
			Category string `json:"category"`
		}
		c.ShouldBindJSON(&req)

		count, err := articleRepo.MarkAllAsRead(userID, req.FeedID, req.Category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Articles marked as read",
			"count":   count,
		})
	}
}
```

**Step 3: 添加 GetUnreadCount handler**

```go
func GetUnreadCount(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		total, byFeed, byCategory, err := articleRepo.GetUnreadCount(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total":       total,
			"by_feed":     byFeed,
			"by_category": byCategory,
		})
	}
}
```

**Step 4: 修改 GetArticles handler 支持 is_read 参数**

修改 `GetArticles` 函数：

```go
func GetArticles(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		feedIDStr := c.Query("feed_id")
		tagIDStr := c.Query("tag_id")
		category := c.Query("category")
		isReadStr := c.Query("is_read")

		var feedID uint
		if feedIDStr != "" {
			id, err := strconv.ParseUint(feedIDStr, 10, 32)
			if err == nil {
				feedID = uint(id)
			}
		}

		var tagID uint
		if tagIDStr != "" {
			id, err := strconv.ParseUint(tagIDStr, 10, 32)
			if err == nil {
				tagID = uint(id)
			}
		}

		var isRead *bool
		if isReadStr != "" {
			val := isReadStr == "true"
			isRead = &val
		}

		articles, total, err := articleRepo.FindByUserID(userID, page, pageSize, feedID, tagID, category, isRead)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"articles": articles,
			"total":    total,
			"page":     page,
			"per_page": pageSize,
		})
	}
}
```

**Step 5: 提交**

```bash
cd /home/prj/rss-reader && git add internal/handlers/article.go && git commit -m "feat: add read status handlers"
```

---

## Task 3: 后端 - 注册路由

**Files:**
- Modify: `/home/prj/rss-reader/backend/main.go`

**Step 1: 添加新路由**

在 `main.go` 中找到 articles 路由组，添加：

```go
// Article routes
articles := protected.Group("/articles")
articles.GET("", handlers.GetArticles(articleRepo))
articles.GET("/search", handlers.SearchArticles(articleRepo))
articles.GET("/unread-count", handlers.GetUnreadCount(articleRepo))
articles.PATCH("/:id/read", handlers.MarkArticleRead(articleRepo))
articles.POST("/mark-all-read", handlers.MarkAllRead(articleRepo))
articles.POST("/tags", handlers.AddArticleTag(articleRepo))
articles.DELETE("/tags", handlers.RemoveArticleTag(articleRepo))
```

**Step 2: 提交**

```bash
cd /home/prj/rss-reader && git add backend/main.go && git commit -m "feat: add read status routes"
```

---

## Task 4: 前端 - 添加 API 调用

**Files:**
- Modify: `/home/prj/rss-reader/frontend/src/api/index.js`

**Step 1: 添加 API 方法**

```javascript
// 标记文章已读
export const markArticleRead = (articleId) => 
  api.patch(`/articles/${articleId}/read`)

// 批量标记已读
export const markAllRead = (feedId, category) => 
  api.post('/articles/mark-all-read', { feed_id: feedId, category })

// 获取未读数量
export const getUnreadCount = () => 
  api.get('/articles/unread-count')
```

**Step 2: 提交**

```bash
cd /home/prj/rss-reader && git add frontend/src/api/index.js && git commit -m "feat: add read status API methods"
```

---

## Task 5: 前端 - 更新 Home.vue

**Files:**
- Modify: `/home/prj/rss-reader/frontend/src/views/Home.vue`

**Step 1: 添加状态变量**

在 `<script setup>` 中添加：

```javascript
const showUnreadOnly = ref(false)
const unreadCount = ref({ total: 0, by_feed: {}, by_category: {} })
```

**Step 2: 修改 openArticle 函数**

```javascript
import { markArticleRead, markAllRead, getUnreadCount } from '../api'

async function openArticle(article) {
  if (article.link) {
    window.open(article.link, '_blank')
  }
  // 标记已读
  if (!article.is_read) {
    try {
      await markArticleRead(article.id)
      article.is_read = true
      fetchUnreadCount()
    } catch (e) {
      console.error('Failed to mark as read:', e)
    }
  }
}
```

**Step 3: 添加批量标记已读函数**

```javascript
async function markAllAsRead(feedId, category) {
  try {
    const res = await markAllRead(feedId, category)
    fetchArticles()
    fetchUnreadCount()
  } catch (e) {
    console.error('Failed to mark all read:', e)
  }
}
```

**Step 4: 添加获取未读数量函数**

```javascript
async function fetchUnreadCount() {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data
  } catch (e) {
    console.error('Failed to fetch unread count:', e)
  }
}
```

**Step 5: 修改 fetchArticles 支持 is_read 筛选**

```javascript
async function fetchArticles() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: perPage.value }
    if (selectedFeedId.value > 0) {
      params.feed_id = selectedFeedId.value
    }
    if (selectedTagId.value > 0) {
      params.tag_id = selectedTagId.value
    }
    if (selectedCategory.value) {
      params.category = selectedCategory.value
    }
    if (showUnreadOnly.value) {
      params.is_read = 'false'
    }
    const res = await api.get('/articles', { params })
    articles.value = res.data.articles || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error('Failed to fetch articles:', e)
  } finally {
    loading.value = false
  }
}
```

**Step 6: 添加 onMounted 调用**

```javascript
onMounted(() => {
  fetchFeeds()
  fetchTags()
  fetchArticles()
  fetchUnreadCount()
})
```

**Step 7: 更新模板 - 添加未读筛选开关**

在 header-right 中添加：

```html
<div class="header-right">
  <label class="toggle-switch">
    <input type="checkbox" v-model="showUnreadOnly" @change="fetchArticles" />
    <span class="toggle-label">只看未读</span>
  </label>
  <button class="btn btn-secondary" @click="markAllAsRead(selectedFeedId, selectedCategory)" 
          :disabled="unreadCount.total === 0">
    全部标记已读
  </button>
  <div class="search-bar">
    ...
  </div>
</div>
```

**Step 8: 更新模板 - 未读文章样式**

文章列表项已有 `:class="{ unread: !article.is_read }"` 样式支持。

添加未读样式：

```css
.article-list-item.unread {
  border-left: 3px solid var(--primary);
}

.article-list-item.unread .article-list-title {
  font-weight: 600;
}
```

**Step 9: 提交**

```bash
cd /home/prj/rss-reader && git add frontend/src/views/Home.vue && git commit -m "feat: add read status UI"
```

---

## Task 6: 重新构建和测试

**Step 1: 重新构建 Docker 镜像**

```bash
cd /home/prj/rss-reader && docker-compose up -d --build
```

**Step 2: 验证功能**

1. 打开应用，确认文章列表正常加载
2. 点击一篇文章，确认自动标记已读
3. 开启"只看未读"开关，确认只显示未读文章
4. 点击"全部标记已读"，确认所有文章变为已读

**Step 3: 最终提交**

```bash
cd /home/prj/rss-reader && git add -A && git commit -m "feat: complete article read status feature"
```
