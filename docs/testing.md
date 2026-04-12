# 测试文档

## 测试概述

本项目使用 Go 标准测试框架，当前测试覆盖率为 **~50%**（Repository 层），**100% 测试通过率**（41/41）。

## 测试结构

```
internal/repository/
├── repository.go              # 仓储层实现
├── repository_test.go         # 测试辅助函数
├── user_repository_test.go    # 用户仓储测试
├── feed_repository_test.go    # 订阅源仓储测试
├── article_repository_test.go # 文章仓储测试
└── tag_repository_test.go     # 标签仓储测试
```

## 运行测试

### 快速运行所有测试

```bash
# 使用测试脚本
bash scripts/test.sh

# 或者直接使用 go test
go test ./... -v -cover
```

### 运行特定包的测试

```bash
# Repository 层测试
go test ./internal/repository/ -v -cover

# Handlers 层测试（待添加）
go test ./internal/handlers/ -v -cover

# Services 层测试（待添加）
go test ./internal/services/ -v -cover
```

### 运行特定测试用例

```bash
# 运行 UserRepository.Create 测试
go test ./internal/repository/ -run TestUserRepository_Create -v

# 运行所有 UserRepository 测试
go test ./internal/repository/ -run TestUserRepository -v
```

### 生成测试覆盖率报告

```bash
# 生成覆盖率 HTML 报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率百分比
go test ./... -cover
```

## 测试覆盖率

### 当前状态

| 包 | 测试文件 | 测试用例 | 覆盖率 | 通过率 | 状态 |
|----|---------|---------|--------|--------|------|
| repository | ✅ | 41 | ~50% | 100% | ✅ 已完成 |
| services | ❌ | - | 0% | - | ⏳ 待添加 |
| handlers | ❌ | - | 0% | - | ⏳ 待添加 |
| models | ❌ | - | 0% | - | ⏳ 待添加 |

### 测试运行结果

**最新运行**: 2026-04-11 00:31 (UTC+8)

```
ok   rss-reader/internal/repository        6.768s
通过率: 45/45 (100%)
```

**环境**:
- 数据库: PostgreSQL 15 (Docker 容器)
- 测试数据库: rss_test
- Go 版本: 1.24.11

**说明**:
- 使用 PostgreSQL 而非 SQLite（避免 CGO 编译问题）
- 每个测试前清理表（DROP TABLE + AutoMigrate）
- 确保测试隔离性

### 测试用例统计

| Repository | 测试用例数 | 状态 |
|-----------|-----------|------|
| UserRepository | 7 | ✅ |
| FeedRepository | 11 | ✅ |
| ArticleRepository | 13 | ✅ |
| TagRepository | 10 | ✅ |
| **总计** | **41** | **✅** |

## 测试依赖

### SQLite（测试数据库）

```bash
# 已添加到 go.mod
go get gorm.io/driver/sqlite
```

**注意**: SQLite 需要编译 C 代码（CGO），首次编译可能需要一些时间。

如果遇到编译问题，可以尝试：

```bash
# 更新 GCC 工具链
sudo yum update gcc

# 或者使用纯 Go 实现的 SQLite（可选）
# go get modernc.org/sqlite
```

### PostgreSQL（生产数据库）

生产环境使用 PostgreSQL，测试**也使用 PostgreSQL**（避免 SQLite CGO 编译问题）。

## 测试指南

### 编写新测试

1. **测试辅助函数**

在 `repository_test.go` 中添加辅助函数：

```go
func createTestFeed(db *gorm.DB, userID uint, title, url, category string) *models.Feed {
    feed := &models.Feed{
        Title:    title,
        URL:      url,
        Category: category,
        UserID:   userID,
    }
    db.Create(feed)
    return feed
}
```

2. **测试用例模板**

```go
func TestRepository_MethodName(t *testing.T) {
    db := setupTestDB(t)
    repo := repository.NewRepository(db)

    // 准备测试数据
    user := createTestUser(db, "test@example.com")

    // 执行操作
    result, err := repo.MethodName(user.ID)

    // 验证结果
    if err != nil {
        t.Fatalf("Failed to execute method: %v", err)
    }

    if result == nil {
        t.Error("Expected result to not be nil")
    }
}
```

### 测试最佳实践

1. **使用 SQLite 内存数据库**
   - 快速、隔离、可重复
   - 每个测试用例使用独立的数据库

2. **表驱动测试（Table-driven tests）**

```go
func TestRepository_Validate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "valid",
            want:    true,
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "invalid",
            want:    false,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := repo.Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.want {
                t.Errorf("Validate() = %v, want %v", result, tt.want)
            }
        })
    }
}
```

3. **测试覆盖率目标**
   - Repository 层：> 70%
   - Service 层：> 60%
   - Handler 层：> 50%
   - 总体：> 60%

4. **测试命名规范**
   - `TestRepository_MethodName` - 基本测试
   - `TestRepository_MethodName_Error` - 错误情况
   - `TestRepository_MethodName_EdgeCase` - 边界情况

## 测试环境

### 本地开发

```bash
# 确保环境变量设置正确
export DATABASE_URL="sqlite::memory:"

# 运行测试
go test ./... -v -cover
```

### CI/CD

测试将在 CI/CD Pipeline 中自动运行：

```yaml
# .github/workflows/ci.yml
- name: Run tests
  run: |
    go test ./... -v -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
```

## 常见问题

### Q: 为什么要用 PostgreSQL 而不是 SQLite？

**A**: 原计划使用 SQLite，但遇到 CGO 编译问题：
- `github.com/mattn/go-sqlite3` 需要 CGO，在当前环境下编译卡住
- 尝试了 `github.com/glebarez/go-sqlite`（纯 Go），但 API 不熟悉

**解决方案**：
- 改用 PostgreSQL 测试
- ✅ 生产环境本来就用 PostgreSQL
- ✅ 不需要 CGO 编译
- ✅ 环境一致性好

### Q: 测试编译很慢

**A**: 如果需要使用 SQLite，CGO 编译较慢。可以使用编译缓存：

```bash
# 启用 Go 编译缓存
export GOCACHE=/tmp/go-cache
```

**注意**: 本项目当前使用 PostgreSQL 测试，无需 CGO。

### Q: 测试卡住

**A**: 可能是数据库连接问题。尝试：

```bash
# 清理测试进程
pkill -f "go test"

# 重新运行测试
go test ./... -v -timeout=30s
```

### Q: 测试失败

**A**: 查看详细错误信息：

```bash
# 显示详细输出
go test ./... -v

# 只显示失败的测试
go test ./... -v | grep FAIL
```

## 下一步

### 阶段一：阶段一完成 ✅

- [x] Repository 层测试
  - [x] UserRepository (7 tests)
  - [x] FeedRepository (11 tests)
  - [x] ArticleRepository (13 tests)
  - [x] TagRepository (10 tests)
- [x] 测试覆盖率 ~50%
- [x] 100% 测试通过率
- [x] 测试文档和脚本

### 阶段二：Service 层测试（下一步）

- [ ] Services 层测试
  - [ ] AuthService
  - [ ] RSSService
  - [ ] OpenAIService

- [ ] Handlers 层测试
  - [ ] Auth Handlers
  - [ ] Feed Handlers
  - [ ] Article Handlers
  - [ ] Tag Handlers

- [ ] 集成测试
  - [ ] API 端到端测试
  - [ ] RSS 抓取测试

### 测试覆盖率目标

| 阶段 | 目标覆盖率 | 实际覆盖率 | 通过率 | 截止日期 | 状态 |
|------|-----------|-----------|--------|---------|------|
| 第一阶段（Repository） | 50% | ~50% | 100% | 2026-04-10 | ✅ 完成 |
| 第二阶段（Service） | 60% | - | - | 2026-04-20 | ⏳ 待开始 |
| 第三阶段（Handler） | 50% | - | - | 2026-04-30 | ⏳ 待开始 |
| 第四阶段（集成） | 70% | - | - | 2026-05-10 | ⏳ 待开始 |

---

**最后更新**: 2026-04-11
**测试状态**: ✅ Repository 层 100% 通过
**测试框架**: Go testing package
