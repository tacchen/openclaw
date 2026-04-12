# RSS Reader 架构文档

## 概述

RSS Reader 是一个基于 Go + Vue 3 的三层架构 Web 应用，使用 Gin 作为 HTTP 框架，GORM 作为 ORM，PostgreSQL 作为数据库。

**设计原则**：
- 职责分离（Separation of Concerns）
- 依赖倒置（Dependency Inversion）
- 单一职责（Single Responsibility）
- 开放封闭（Open/Closed）

---

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│                      (Vue 3 + Vite)                         │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │   Views    │  │   Stores   │  │    API     │            │
│  │ (组件/页面) │  │ (Pinia)    │  │  (axios)   │            │
│  └────────────┘  └────────────┘  └────────────┘            │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP/REST (JSON)
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend (Go)                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Handlers Layer (HTTP 处理)                          │   │
│  │  - 路由定义                                            │   │
│  │  - 请求参数验证                                         │   │
│  │  - JWT 认证中间件                                       │   │
│  │  - 响应格式化                                           │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │ 调用                                   │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Services Layer (业务逻辑)                            │   │
│  │  - AuthService: 用户认证                                │   │
│  │  - RSSService: RSS 抓取和解析                           │   │
│  │  - OpenAIService: AI 摘要生成                          │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │ 调用                                   │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Repository Layer (数据访问)                         │   │
│  │  - UserRepository: 用户 CRUD                          │   │
│  │  - FeedRepository: 订阅源 CRUD                         │   │
│  │  - ArticleRepository: 文章 CRUD                        │   │
│  │  - TagRepository: 标签 CRUD                            │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │ 使用 GORM                              │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Models Layer (数据模型)                              │   │
│  │  - User, Feed, Article, Tag, ArticleTag              │   │
│  └──────────────────┬───────────────────────────────────┘   │
└─────────────────────┼───────────────────────────────────────┘
                      │ GORM ORM
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   PostgreSQL 数据库                         │
│  users | feeds | articles | tags | article_tags             │
└─────────────────────────────────────────────────────────────┘
```

---

## 模块划分

### 1. Handlers Layer (`internal/handlers/`)

**职责**：
- HTTP 请求处理
- 请求参数验证和绑定
- 调用 Service 层执行业务逻辑
- 返回 HTTP 响应
- 错误处理和状态码设置

**包含内容**：
- 路由定义 (`router`)
- 请求处理函数 (`handlers`)
- JWT 认证中间件 (`AuthMiddleware`)
- CORS 配置

**文件**: `handlers.go`

**关键函数**：
```go
// 认证
Register(authService)          // 用户注册
Login(authService)             // 用户登录

// 订阅源
GetFeeds(feedRepo)             // 获取订阅源列表
CreateFeed(feedRepo, rssService)  // 添加订阅源
UpdateFeed(feedRepo)           // 更新订阅源
DeleteFeed(feedRepo)           // 删除订阅源

// 文章
GetArticles(articleRepo)        // 获取文章列表
SearchArticles(articleRepo)     // 搜索文章
GetUnreadCount(articleRepo)    // 获取未读数量
MarkArticleRead(articleRepo)    // 标记文章已读
MarkAllRead(articleRepo)        // 批量标记已读

// AI 摘要
GenerateArticleSummary(articleRepo, openaiService)  // 生成 AI 摘要
GetArticleSummary(articleRepo)                      // 获取 AI 摘要

// 标签
GetTags(tagRepo)               // 获取标签
CreateTag(tagRepo)             // 创建标签
DeleteTag(tagRepo)             // 删除标签

// 文章标签关联
AddArticleTag(articleRepo)     // 添加标签到文章
RemoveArticleTag(articleRepo)  // 移除文章标签
```

**禁止事项**：
- ❌ 直接访问数据库
- ❌ 实现业务逻辑
- ❌ 调用其他 Handler

---

### 2. Services Layer (`internal/services/`)

**职责**：
- 业务逻辑实现
- 协调多个 Repository
- 调用外部 API（RSS 解析、OpenAI API）
- 事务管理
- 密码哈希和验证

**包含内容**：
- `AuthService`: 用户注册、登录、密码验证
- `RSSService`: RSS 抓取、解析、文章保存
- `OpenAIService`: AI 摘要生成

**文件**: `services.go`

**关键逻辑**：

**AuthService**：
```go
- Register(username, password): 用户注册
  - 检查用户名是否已存在
  - 哈希密码（bcrypt）
  - 创建用户记录

- Login(username, password): 用户登录
  - 查找用户
  - 验证密码
  - 生成 JWT Token
```

**RSSService**：
```go
- FetchAllFeeds(): 抓取所有订阅源
  - 遍历所有订阅源
  - 解析 RSS XML
  - 保存新文章（去重）
  - 为新文章生成 AI 摘要

- ParseFeed(url): 解析单个 RSS 源
  - HTTP GET 获取 RSS 内容
  - XML 解析
  - 转换为 Article 模型
```

**OpenAIService**：
```go
- GenerateSummary(content): 生成摘要
  - 调用 OpenAI API
  - 提示词工程
  - 返回摘要文本
```

**禁止事项**：
- ❌ HTTP 请求处理
- ❌ 数据库细节（GORM 查询）

---

### 3. Repository Layer (`internal/repository/`)

**职责**：
- 数据库访问
- GORM 查询封装
- CRUD 操作
- 复杂查询实现

**包含内容**：
- `UserRepository`: 用户 CRUD
- `FeedRepository`: 订阅源 CRUD
- `ArticleRepository`: 文章 CRUD
- `TagRepository`: 标签 CRUD
- `ArticleTagRepository`: 文章-标签关联

**文件**: `repository.go`

**关键方法**：

**UserRepository**：
```go
Create(user)              // 创建用户
FindByUsername(username)  // 按用户名查找
FindByID(id)              // 按 ID 查找
```

**FeedRepository**：
```go
Create(feed)              // 创建订阅源
GetByUserID(userID)       // 获取用户的所有订阅源
FindByID(id)              // 按 ID 查找
Update(feed)              // 更新订阅源
Delete(id)                // 删除订阅源
```

**ArticleRepository**：
```go
Create(article)           // 创建文章
CreateBatch(articles)      // 批量创建
GetByUserID(userID, page, pageSize)    // 分页获取
GetByFeedID(feedID, page, pageSize)    // 按订阅源获取
SearchByTitle(keyword, page, pageSize)  // 搜索
Update(article)           // 更新
MarkAsRead(id)            // 标记已读
MarkAllAsRead(userID)     // 批量标记已读
GetUnreadCount(userID)    // 未读数量
```

**TagRepository**：
```go
Create(tag)               // 创建标签
GetByUserID(userID)       // 获取用户的所有标签
FindByID(id)              // 按 ID 查找
Delete(id)                // 删除标签
```

**禁止事项**：
- ❌ 业务逻辑
- ❌ 调用外部 API

---

### 4. Models Layer (`internal/models/`)

**职责**：
- 数据模型定义
- 数据库表映射
- 结构体标签（GORM, JSON）
- 验证规则

**文件**: `models.go`

**数据模型**：

```go
// User - 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Feed - 订阅源模型
type Feed struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `gorm:"size:255;not null" json:"title"`
    URL       string    `gorm:"size:500;not null;uniqueIndex:idx_user_url" json:"url"`
    Category  string    `gorm:"size:100" json:"category"`
    UserID    uint      `gorm:"not null;index:idx_user_url" json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Article - 文章模型
type Article struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Title       string    `gorm:"size:500;not null" json:"title"`
    URL         string    `gorm:"size:1000;not null;uniqueIndex:idx_feed_url" json:"url"`
    Summary     string    `gorm:"type:text" json:"summary"`
    Content     string    `gorm:"type:text" json:"content"`
    PublishedAt time.Time `json:"published_at"`
    IsRead      bool      `gorm:"default:false" json:"is_read"`
    Aisummary   string    `gorm:"type:text" json:"ai_summary"`
    FeedID      uint      `gorm:"not null;index:idx_feed_url;index" json:"feed_id"`
    UserID      uint      `gorm:"not null;index" json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Tag - 标签模型
type Tag struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"size:100;not null;uniqueIndex:idx_user_name" json:"name"`
    UserID    uint      `gorm:"not null;index:idx_user_name" json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

// ArticleTag - 文章-标签关联模型
type ArticleTag struct {
    ArticleID uint `gorm:"primaryKey" json:"article_id"`
    TagID     uint `gorm:"primaryKey" json:"tag_id"`
}
```

**关键索引**：
- `idx_user_url`: 用户 + 订阅源 URL 唯一索引（防止重复订阅）
- `idx_feed_url`: 订阅源 + 文章 URL 唯一索引（防止重复文章）
- `idx_user_name`: 用户 + 标签名 唯一索引（防止重复标签）

---

### 5. Config Layer (`internal/config/`)

**职责**：
- 环境变量加载
- 配置管理

**文件**: `config.go`

**配置项**：
```go
type Config struct {
    Port         string // 服务器端口
    DatabaseURL  string // PostgreSQL 连接串
    JWTSecret    string // JWT 密钥
    OpenAIAPIKey string // OpenAI API Key（可选）
}
```

---

### 6. Utils Layer (`internal/utils/`)

**职责**：
- 工具函数
- 辅助功能

**文件**: `favicon.go`（处理 favicon）

---

## 数据库设计

### 表关系

```
users (用户)
  ├── feeds (订阅源) [1:N]
  │     └── articles (文章) [1:N]
  │           └── article_tags (文章-标签) [N:M]
  │                 └── tags (标签)
  └── tags (标签) [1:N]
```

### 外键约束

| 表 | 外键 | 引用表 | 删除规则 |
|----|------|--------|---------|
| feeds | user_id | users | CASCADE |
| articles | user_id | users | CASCADE |
| articles | feed_id | feeds | CASCADE |
| tags | user_id | users | CASCADE |
| article_tags | article_id | articles | CASCADE |
| article_tags | tag_id | tags | CASCADE |

### 唯一约束

| 表 | 约束 | 说明 |
|----|------|------|
| users | username | 用户名唯一 |
| feeds | user_id + url | 用户不能重复订阅同一源 |
| articles | feed_id + url | 同一订阅源不重复文章 |
| tags | user_id + name | 用户标签名唯一 |

---

## 请求流程

### 典型的 API 请求流程

```
1. 前端发送 HTTP 请求
   ↓
2. Gin Router 接收请求
   ↓
3. CORS 中间件处理跨域
   ↓
4. AuthMiddleware 验证 JWT Token（除 /api/auth/*）
   ↓
5. Handler 函数处理请求
   - 解析请求参数
   - 验证参数
   ↓
6. Handler 调用 Service
   ↓
7. Service 执行业务逻辑
   - 调用 Repository 获取数据
   - 调用外部 API（如需要）
   ↓
8. Repository 执行数据库查询（GORM）
   ↓
9. PostgreSQL 返回数据
   ↓
10. Service 返回结果
   ↓
11. Handler 格式化响应
   ↓
12. 返回 HTTP 响应给前端
```

### 示例：获取文章列表

```go
// 1. 前端请求
GET /api/articles?page=1&pageSize=20

// 2. Handler: GetArticles
func GetArticles(c *gin.Context, repo *repository.ArticleRepository) {
    // 3. 从 JWT 获取用户
    user := getUserFromJWT(c)

    // 4. 解析查询参数
    page, _ := c.GetQuery("page")
    pageSize, _ := c.GetQuery("pageSize")

    // 5. 调用 Repository
    articles, total, err := repo.GetByUserID(
        user.ID,
        page,
        pageSize,
    )

    // 6. 返回响应
    c.JSON(200, gin.H{
        "articles": articles,
        "total":    total,
        "page":     page,
        "pageSize": pageSize,
    })
}

// 7. Repository: GetByUserID
func (r *ArticleRepository) GetByUserID(
    userID uint,
    page, pageSize string,
) ([]*models.Article, int64, error) {
    var articles []*models.Article
    var total int64

    offset, limit := parsePagination(page, pageSize)

    // 8. GORM 查询
    r.db.Model(&models.Article{}).
        Where("user_id = ?", userID).
        Count(&total)

    err := r.db.
        Where("user_id = ?", userID).
        Offset(offset).
        Limit(limit).
        Order("published_at DESC").
        Find(&articles).Error

    return articles, total, err
}
```

---

## 定时任务

### RSS 自动抓取

**工具**: `github.com/robfig/cron/v3`

**执行间隔**: 每 30 分钟

**触发流程**：
```
Cron 触发
  ↓
RSSService.FetchAllFeeds()
  ↓
遍历所有订阅源
  ↓
ParseFeed(url)
  ↓
保存新文章（去重）
  ↓
为每篇新文章生成 AI 摘要（如配置了 OpenAI API Key）
  ↓
完成
```

**代码位置**: `backend/main.go`

---

## 安全机制

### 1. JWT 认证

**算法**: HS256
**密钥**: 从环境变量 `JWT_SECRET` 读取
**过期时间**: 24 小时

**Token 结构**：
```json
{
  "user_id": 1,
  "username": "testuser",
  "exp": 1744435200
}
```

### 2. 密码哈希

**算法**: bcrypt
**成本**: 10（默认）

**存储**：哈希后的密码存储在 `users.password`

### 3. 数据库安全

- 使用 GORM 参数化查询（防止 SQL 注入）
- 外键约束（防止孤立数据）
- 事务保护关键操作

### 4. CORS 配置

```go
allowOrigins:     []string{"*"}
allowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
allowHeaders:     []string{"Origin", "Content-Type", "Authorization"}
```

---

## 错误处理

### 错误类型

```go
// Repository 层错误
ErrNotFound      = "not found"
ErrDuplicate     = "duplicate entry"

// Service 层错误
ErrInvalidInput  = "invalid input"
ErrUnauthorized  = "unauthorized"

// Handler 层错误
ErrBadRequest    = 400
ErrUnauthorized  = 401
ErrNotFound      = 404
ErrInternalError = 500
```

### 错误响应格式

```json
{
  "error": "error message"
}
```

---

## 性能优化

### 已实现

1. **数据库索引**
   - `idx_user_url` (feeds)
   - `idx_feed_url` (articles)
   - `idx_user_name` (tags)

2. **分页查询**
   - 避免一次性加载所有数据
   - 默认每页 20 条

3. **定时任务**
   - 每 30 分钟抓取，避免频繁请求

### 可优化

1. **缓存**
   - Redis 缓存热点数据
   - Feed 内容缓存

2. **连接池**
   - 数据库连接池配置
   - HTTP 客户端连接池

3. **查询优化**
   - 避免 N+1 查询
   - 使用 GORM Preload

---

## 部署架构

### Docker Compose

```
┌─────────────────────────────────────────────┐
│         Docker Compose 网络                  │
│                                             │
│  ┌─────────────────┐  ┌──────────────────┐   │
│  │   app 容器      │  │   db 容器        │   │
│  │  (Go + Vue)     │  │  (PostgreSQL)    │   │
│  │   Port: 8080    │  │   Port: 5432     │   │
│  └─────────────────┘  └──────────────────┘   │
│          │                    │              │
│          └────────────────────┘              │
│                                             │
└─────────────────────────────────────────────┘
```

### 服务端口

- **HTTP 服务**: 8080（内部）→ 80（对外）
- **PostgreSQL**: 5432

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 8080 |
| DATABASE_URL | PostgreSQL 连接串 | - |
| JWT_SECRET | JWT 密钥 | - |
| OPENAI_API_KEY | OpenAI API Key | - |
| OPENAI_BASE_URL | OpenAI API URL | - |
| OPENAI_MODEL | OpenAI 模型 | - |

---

## 扩展性设计

### 添加新的 API 端点

1. 在 `handlers.go` 中添加 handler 函数
2. 在 `main.go` 中注册路由
3. 如需新数据操作，添加 Repository 方法
4. 如需新业务逻辑，添加 Service 方法
5. 更新 `docs/api.md`

### 添加新的数据模型

1. 在 `models.go` 中定义模型
2. 创建迁移脚本
3. 添加对应的 Repository
4. 添加对应的 Service
5. 添加对应的 Handler
6. 更新 API 文档

---

## 开发规范

### Go 代码规范

1. **命名**
   - 包名：小写、单词
   - 函数名：PascalCase（导出）/ camelCase（私有）
   - 变量名：camelCase

2. **错误处理**
   ```go
   if err != nil {
       return fmt.Errorf("context: %w", err)
   }
   ```

3. **注释**
   - 公开函数必须有注释
   - 复杂逻辑必须有注释

### Vue 代码规范

1. **组件命名**: PascalCase
2. **文件命名**: kebab-case
3. **组合式 API**: 优先使用 `<script setup>`

---

## 最后更新

**更新日期**: 2026-04-10
**架构版本**: v1.0
