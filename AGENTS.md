# AGENTS.md - RSS Reader Agent 工作指南

## 项目上下文

**项目名称**: RSS Reader - RSS 新闻订阅系统
**项目目标**: 让用户订阅、管理和阅读 RSS 新闻源
**技术栈**:
- **后端**: Go 1.21+ (Gin + GORM + PostgreSQL)
- **前端**: Vue 3 (Vite)
- **部署**: Docker Compose

**架构模式**: 三层架构
```
Handlers (HTTP) → Services (业务逻辑) → Repository (数据访问) → Database
```

**代码库位置**: `/home/prj/rss-reader`

---

## 项目结构

```
rss-reader/
├── backend/main.go              # 入口文件
├── internal/
│   ├── config/                  # 配置管理
│   │   └── config.go            # Load() 函数加载环境变量
│   ├── handlers/                # HTTP 处理器
│   │   └── handlers.go          # 所有路由和处理函数
│   ├── services/                # 业务逻辑
│   │   └── services.go          # AuthService, RSSService, OpenAIService
│   ├── repository/              # 数据访问层
│   │   └── repository.go        # 所有数据库操作
│   ├── models/                  # 数据模型
│   │   └── models.go            # User, Feed, Article, Tag, ArticleTag
│   ├── utils/                   # 工具函数
│   │   └── favicon.go
│   └── migration/               # 数据库迁移
│       └── README.md
├── frontend/                    # Vue 3 前端
│   ├── src/
│   │   ├── api/                 # API 封装
│   │   ├── router/              # 路由
│   │   ├── stores/              # Pinia 状态管理
│   │   └── views/               # 页面组件
│   └── package.json
├── docs/                        # 机器可读文档
│   ├── architecture.md         # 架构文档（待创建）
│   ├── api.md                   # API 文档（待创建）
│   └── plans/                   # 功能设计文档
├── .agents/                     # Agent 专用目录（待创建）
│   ├── context/                 # 上下文模板
│   ├── tools/                   # 自定义工具
│   ├── validators/              # 验证器
│   └── memory/                  # 记忆管理
├── AGENTS.md                    # 本文件
├── README.md                    # 项目说明
└── COMMIT_CONVENTION.md          # Git 提交规范
```

---

## 行为规范

### 1. 前置检查

在修改代码前，**必须**完成以下步骤：

- [ ] 阅读相关文档（docs/ 目录下的架构和 API 文档）
- [ ] 阅读现有代码（特别是要修改的模块）
- [ ] 理解业务逻辑（通过 services 层）
- [ ] 检查是否有现有测试

### 2. 代码变更规则

**每当代码变更时，必须：**

1. **包含对应的测试**
   - 修改 Repository → 添加 Repository 测试
   - 修改 Service → 添加 Service 测试
   - 修改 Handler → 添加 Handler 测试
   - 修改 Model → 添加相关测试

2. **保持向后兼容**
   - API 变更需提供废弃期
   - 数据库变更需编写迁移脚本
   - 重大变更需先讨论

3. **遵循代码风格**
   - Go: 使用 `gofmt` 格式化
   - 遵循 `golangci-lint` 检查规则
   - 前端: 遵循 ESLint 和 Prettier

### 3. 不确定情况处理

**遇到不确定的问题时：**

1. 先询问，不要猜测
2. 提供多个解决方案，说明优缺点
3. 等待人工确认后再实施

**禁止行为：**
- ❌ 不要凭空捏造不存在的 API 或函数
- ❌ 不要假设数据库表结构，查看 models/models.go
- ❌ 不要直接修改生产数据库
- ❌ 不要提交 `.env` 文件

### 4. 测试优先原则

**测试驱动开发（TDD）优先：**

1. 先写测试（失败的测试）
2. 再写实现代码
3. 运行测试直到通过
4. 重构优化

**测试要求：**
- Repository 层: 使用 SQLite 内存数据库
- Service 层: Mock Repository
- Handler 层: Mock Service，使用 httptest

---

## 工具使用规则

### Go 工具

| 工具 | 用途 | 命令 |
|------|------|------|
| **go test** | 运行测试 | `go test ./... -v -cover` |
| **go build** | 检查编译 | `go build ./...` |
| **gofmt** | 格式化代码 | `gofmt -w .` |
| **go vet** | 静态分析 | `go vet ./...` |
| **golangci-lint** | 综合检查 | `golangci-lint run` |

### 前端工具

| 工具 | 用途 | 命令 |
|------|------|------|
| **npm test** | 运行测试 | `cd frontend && npm test` |
| **npm run lint** | ESLint 检查 | `cd frontend && npm run lint` |
| **npm run build** | 构建前端 | `cd frontend && npm run build` |

### Docker 工具

| 工具 | 用途 | 命令 |
|------|------|------|
| **docker-compose up** | 启动服务 | `docker-compose up -d --build` |
| **docker-compose down** | 停止服务 | `docker-compose down` |
| **docker-compose logs** | 查看日志 | `docker-compose logs -f app` |

---

## 记忆规则

### 1. 长期记忆（MEMORY.md）

**重要决策和知识写入 MEMORY.md：**

- 架构决策和原因
- 重要的技术选型
- 已知问题和解决方案
- 最佳实践和模式

**示例：**
```markdown
## Architecture Decision: 为什么使用 GORM

- 决策日期: 2026-04-10
- 原因: 类型安全、自动化迁移、易于测试
- 替代方案: sqlx, raw SQL
```

### 2. 每日记录

**每日工作记录到 `memory/YYYY-MM-DD.md`：**

- 完成的任务
- 遇到的问题和解决方案
- 学到的经验
- 待办事项

### 3. 功能文档

**新功能文档写入 `docs/features/` 或 `docs/plans/`：**

- 功能设计文档
- API 变更说明
- 数据库变更说明

---

## 代码风格

### Go 代码风格

1. **遵循标准规范**
   - 使用 `gofmt` 自动格式化
   - 遵循 `Effective Go` 最佳实践
   - 使用有意义的变量名

2. **错误处理**
   ```go
   // 错误：忽略错误
   db.Create(user)

   // 正确：处理错误
   if err := db.Create(user).Error; err != nil {
       return fmt.Errorf("failed to create user: %w", err)
   }
   ```

3. **函数长度**
   - 单个函数不超过 50 行
   - 职责单一
   - 可读性强

4. **注释**
   - 公开函数必须有注释
   - 复杂逻辑必须有注释
   - 不注释显而易见的代码

### Vue 3 代码风格

1. **组件命名**
   - 使用 PascalCase（如 `ArticleList.vue`）
   - 描述性强

2. **Props 定义**
   ```typescript
   interface Props {
     article: Article
     onMarkRead: (id: number) => void
   }
   ```

3. **组合式 API**
   - 优先使用 `<script setup>`
   - 使用 TypeScript 类型定义

---

## Git 提交规范

**遵循 COMMIT_CONVENTION.md：**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档变更
- `style`: 代码格式（不影响功能）
- `refactor`: 重构代码
- `perf`: 性能优化
- `test`: 添加测试
- `build`: 构建系统或依赖变更
- `ci`: CI 配置变更
- `chore`: 其他杂项

**示例：**
```bash
git commit -m "feat(repository): add UserRepository.Create method"
git commit -m "fix(handlers): handle nil feed ID gracefully"
git commit -m "test(repository): add unit tests for FeedRepository"
```

---

## 安全规则

### 1. 敏感信息

**永远不要提交：**
- ❌ `.env` 文件
- ❌ API Keys（OpenAI, JWT Secret 等）
- ❌ 密码
- ❌ 证书文件

**正确做法：**
- 使用环境变量
- 使用 secrets 管理工具
- `.env` 已在 `.gitignore` 中

### 2. 认证和授权

**所有 API 端点（除了 `/api/auth/*）必须经过认证：**

- 使用 JWT Token
- 验证 Token 有效性
- 检查用户权限

### 3. 数据库安全

**直接访问数据库规则：**
- ❌ Agent 不能直接连接生产数据库
- ✅ 使用 Repository 层操作数据库
- ✅ 使用事务保护关键操作

### 4. 输入验证

**所有用户输入必须验证：**
- HTTP 请求参数
- 查询参数
- POST/PUT 请求体

---

## 模块职责

### Handlers 层 (`internal/handlers/`)

**职责：**
- HTTP 请求处理
- 参数验证和绑定
- 调用 Service 层
- 返回 HTTP 响应

**禁止：**
- ❌ 直接访问数据库
- ❌ 复杂业务逻辑
- ❌ 调用其他 Handler

**示例：**
```go
func GetFeeds(c *gin.Context, repo *repository.FeedRepository) {
    // 1. 获取用户（从 JWT）
    user := getUserFromJWT(c)

    // 2. 调用 Repository 获取数据
    feeds, err := repo.GetByUserID(user.ID)

    // 3. 返回响应
    c.JSON(200, feeds)
}
```

### Services 层 (`internal/services/`)

**职责：**
- 业务逻辑
- 协调多个 Repository
- 调用外部 API（RSS, OpenAI）
- 事务管理

**禁止：**
- ❌ HTTP 处理
- ❌ 数据库细节

**示例：**
```go
func (s *RSSService) FetchFeed(feed *models.Feed) error {
    // 1. 调用外部 API
    items, err := s.parser.Parse(feed.URL)
    if err != nil {
        return err
    }

    // 2. 业务逻辑处理
    articles := s.convertToArticles(feed, items)

    // 3. 调用 Repository 保存
    return s.articleRepo.CreateBatch(articles)
}
```

### Repository 层 (`internal/repository/`)

**职责：**
- 数据库访问
- GORM 查询
- CRUD 操作

**禁止：**
- ❌ 业务逻辑
- ❌ 调用外部 API

**示例：**
```go
func (r *FeedRepository) GetByUserID(userID uint) ([]*models.Feed, error) {
    var feeds []*models.Feed
    err := r.db.Where("user_id = ?", userID).Find(&feeds).Error
    return feeds, err
}
```

### Models 层 (`internal/models/`)

**职责：**
- 数据模型定义
- 数据库表映射
- 结构体标签

**示例：**
```go
type Feed struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Title     string         `gorm:"size:255;not null" json:"title"`
    URL       string         `gorm:"size:500;not null;uniqueIndex:idx_user_url" json:"url"`
    Category  string         `gorm:"size:100" json:"category"`
    UserID    uint           `gorm:"not null;index:idx_user_url" json:"user_id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}
```

---

## 数据模型

### User（用户）
```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;size:100;not null"`
    Password  string    `gorm:"size:255;not null"` // bcrypt 哈希
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### Feed（订阅源）
```go
type Feed struct {
    ID        uint
    Title     string
    URL       string         `gorm:"uniqueIndex:idx_user_url"`
    Category  string
    UserID    uint
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Article（文章）
```go
type Article struct {
    ID          uint
    Title       string
    URL         string         `gorm:"uniqueIndex:idx_feed_url"`
    Summary     string
    Content     string
    PublishedAt time.Time
    FeedID      uint
    UserID      uint
    IsRead      bool           `gorm:"default:false"`
    Aisummary   string         // AI 生成的摘要
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Tag（标签）
```go
type Tag struct {
    ID        uint
    Name      string `gorm:"uniqueIndex:idx_user_name"`
    UserID    uint
    CreatedAt time.Time
}
```

### ArticleTag（文章-标签关联）
```go
type ArticleTag struct {
    ArticleID uint `gorm:"primaryKey"`
    TagID     uint `gorm:"primaryKey"`
}
```

---

## API 设计原则

### 1. RESTful 风格

- 使用 HTTP 动词（GET, POST, PUT, PATCH, DELETE）
- 资源导向的 URL 设计
- 使用状态码（200, 201, 400, 401, 404, 500）

### 2. 统一错误响应格式

```json
{
  "error": "error message"
}
```

### 3. 分页参数

- `page`: 页码（从 1 开始）
- `pageSize`: 每页数量（默认 20，最大 100）

**示例：**
```bash
GET /api/articles?page=1&pageSize=20
```

### 4. 认证

**除 `/api/auth/*` 外，所有 API 必须认证：**

```http
Authorization: Bearer <jwt_token>
```

### 5. 常见响应状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|---------|
| 200 | OK | 请求成功 |
| 201 | Created | 资源创建成功 |
| 400 | Bad Request | 参数错误 |
| 401 | Unauthorized | 未认证或 Token 无效 |
| 404 | Not Found | 资源不存在 |
| 500 | Internal Server Error | 服务器错误 |

---

## 开发流程

### 1. 修改现有代码

```bash
# 1. 拉取最新代码
git pull

# 2. 创建新分支
git checkout -b feature/my-feature

# 3. 阅读相关文档和代码
cat docs/architecture.md
cat docs/api.md
cat internal/handlers/handlers.go

# 4. 修改代码
# 5. 添加测试
# 6. 运行测试
go test ./... -v -cover

# 7. 运行 linter
golangci-lint run

# 8. 格式化代码
gofmt -w .

# 9. 提交代码
git add .
git commit -m "feat: add my feature"

# 10. 推送到远程
git push origin feature/my-feature
```

### 2. 添加新功能

```bash
# 1. 编写测试（TDD）
# 2. 实现 Model（数据模型）
# 3. 实现 Repository
# 4. 实现 Service
# 5. 实现 Handler
# 6. 更新 API 文档（docs/api.md）
# 7. 更新前端 API 调用
# 8. 运行测试
# 9. 代码审查
# 10. 合并到主分支
```

### 3. 修复 Bug

```bash
# 1. 复现 Bug
# 2. 编写测试（失败的测试）
# 3. 修复代码
# 4. 运行测试（通过）
# 5. 检查回归测试
# 6. 提交修复
```

---

## 测试指南

### 1. Repository 层测试

**使用 SQLite 内存数据库：**

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to connect to test database: %v", err)
    }

    db.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{})

    return db
}

func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := repository.NewUserRepository(db)

    user := &models.User{
        Username: "testuser",
        Password: "hashed_password",
    }

    err := repo.Create(user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    if user.ID == 0 {
        t.Error("User ID should not be 0")
    }
}
```

### 2. Service 层测试

**Mock Repository：**

```go
type MockUserRepository struct {
    users []*models.User
    err   error
}

func (m *MockUserRepository) Create(user *models.User) error {
    if m.err != nil {
        return m.err
    }
    m.users = append(m.users, user)
    return nil
}

func TestAuthService_Register(t *testing.T) {
    mockRepo := &MockUserRepository{}
    authService := services.NewAuthService(mockRepo, "secret")

    err := authService.Register("testuser", "password")

    if err != nil {
        t.Fatalf("Failed to register: %v", err)
    }

    if len(mockRepo.users) != 1 {
        t.Error("Should create one user")
    }
}
```

### 3. Handler 层测试

**使用 httptest：**

```go
func TestHandler_GetFeeds(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // Setup router
    r := gin.Default()
    mockRepo := &MockFeedRepository{}
    r.GET("/api/feeds", func(c *gin.Context) {
        handlers.GetFeeds(c, mockRepo)
    })

    // Create request
    req, _ := http.NewRequest("GET", "/api/feeds", nil)
    w := httptest.NewRecorder()

    // Perform request
    r.ServeHTTP(w, req)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

---

## 常见任务

### 添加新的 API 端点

1. 在 `internal/handlers/handlers.go` 中添加 handler 函数
2. 在 `backend/main.go` 中注册路由
3. 在 `internal/repository/repository.go` 中添加数据访问方法（如需要）
4. 在 `internal/services/services.go` 中添加业务逻辑（如需要）
5. 更新 `docs/api.md`
6. 添加测试
7. 运行测试和 linter

### 修改数据库模型

1. 修改 `internal/models/models.go`
2. 创建迁移脚本 `internal/migration/XXX_new_field.sql`
3. 更新 Repository 方法（如需要）
4. 添加测试
5. 在本地测试迁移
6. 更新 `docs/database-schema.md`（如有）

### 添加新功能

1. 编写功能设计文档 `docs/plans/YYYY-MM-DD-feature-name.md`
2. 实现 Model（数据模型）
3. 实现 Repository
4. 实现 Service
5. 实现 Handler
6. 更新前端 API 调用
7. 更新前端 UI
8. 添加测试
9. 代码审查
10. 合并到主分支

---

## 故障排查

### 问题：go test 卡住

**原因：** 可能是数据库连接问题

**解决：**
```bash
# 使用 SQLite 内存数据库进行测试
go test ./... -tags=sqlite
```

### 问题：golangci-lint 报错

**原因：** 代码风格不符合规范

**解决：**
```bash
# 查看具体错误
golangci-lint run --out-format=line-number

# 自动修复部分问题
golangci-lint run --fix
```

### 问题：Docker 容器无法启动

**原因：** 可能是端口冲突或配置错误

**解决：**
```bash
# 查看日志
docker-compose logs app

# 检查端口占用
lsof -i :80

# 重建容器
docker-compose up -d --build --force-recreate
```

---

## 参考资源

### 官方文档
- [Go 文档](https://go.dev/doc/)
- [Gin 框架](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [Vue 3 文档](https://vuejs.org/)
- [Vite 文档](https://vitejs.dev/)

### 项目文档
- [README.md](../README.md) - 项目说明
- [COMMIT_CONVENTION.md](../COMMIT_CONVENTION.md) - Git 提交规范
- [docs/architecture.md](docs/architecture.md) - 架构文档（待创建）
- [docs/api.md](docs/api.md) - API 文档（待创建）
- [harness-assessment.md](docs/harness-assessment.md) - Harness 评估报告

### 外部资源
- [Harness Engineering](/home/doc/harness-engineering.md) - Harness 调研报告
- [Harness 快速参考](/home/doc/harness-engineering-quickref.md)
- [Effective Go](https://go.dev/doc/effective_go)

---

## 最后更新

**更新日期**: 2026-04-10
**更新内容**: 创建 AGENTS.md，定义 AI Agent 行为规范

---

**记住：**
- 🧠 先思考，再行动
- 📝 测试优先
- 🔄 持续改进
- 🤝 代码审查很重要
