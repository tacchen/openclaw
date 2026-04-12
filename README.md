# RSS Reader

[![CI](https://github.com/your-username/rss-reader/workflows/CI/badge.svg)](https://github.com/your-username/rss-reader/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/rss-reader)](https://goreportcard.com/report/github.com/your-username/rss-reader)

一个基于 Go + Vue 3 的 RSS 新闻订阅系统。

## 技术栈

- **后端**: Go 1.21+ (Gin + GORM)
- **前端**: Vue 3 + Vite
- **数据库**: PostgreSQL 15
- **部署**: Docker Compose

## 功能

- ✅ 用户注册/登录
- ✅ RSS 源管理（增删改查）
- ✅ 自动定时抓取（每 30 分钟）
- ✅ 文章列表展示（分页）
- ✅ 按分类筛选
- ✅ 标题搜索

## 快速开始

### 前置条件

- Docker
- Docker Compose
- Go 1.21+ (本地开发)

### 启动服务

```bash
cd /home/prj/rss-reader
docker-compose up -d --build
```

服务将在 **80 端口** 启动。

### 访问

打开浏览器访问: http://服务器IP

## 目录结构

```
rss-reader/
├── backend/
│   └── main.go           # 入口文件
├── internal/
│   ├── config/           # 配置
│   ├── handlers/         # HTTP 处理器
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── services/         # 业务逻辑层
├── frontend/             # Vue 3 前端
│   ├── src/
│   │   ├── api/          # API 封装
│   │   ├── router/       # 路由
│   │   ├── stores/       # Pinia 状态管理
│   │   └── views/        # 页面组件
│   └── package.json
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| GET | /api/feeds | 获取订阅列表 |
| POST | /api/feeds | 添加订阅源 |
| PUT | /api/feeds/:id | 更新订阅源 |
| DELETE | /api/feeds/:id | 删除订阅源 |
| GET | /api/articles | 获取文章列表 |
| GET | /api/articles/search | 标题搜索 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 服务端口 |
| DATABASE_URL | - | PostgreSQL 连接串 |
| JWT_SECRET | - | JWT 密钥 |
| OPENAI_API_KEY | - | OpenAI API Key（可选）|
| OPENAI_BASE_URL | - | OpenAI API URL（可选）|
| OPENAI_MODEL | - | OpenAI 模型（可选）|

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./... -v -cover

# 使用测试脚本
bash scripts/test.sh

# 生成测试覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**注意**: 测试使用 SQLite 内存数据库，首次编译需要一些时间（CGO）。

### 代码检查

```bash
# 运行 golangci-lint
golangci-lint run

# 代码格式化
gofmt -w .
```

### 本地开发

```bash
# 安装依赖
go mod download

cd frontend
npm install

# 启动后端（需要 PostgreSQL 数据库）
export DATABASE_URL="postgres://user:password@localhost:5432/rss?sslmode=disable"
go run backend/main.go

# 启动前端
cd frontend
npm run dev
```

## 停止服务

```bash
docker-compose down
```

## 文档

- [API 文档](./docs/api.md) - 完整的 API 参考
- [架构文档](./docs/architecture.md) - 系统架构设计
- [测试文档](./docs/testing.md) - 测试指南和覆盖率
- [AGENTS.md](./AGENTS.md) - AI Agent 行为指南
- [Harness 评估](./docs/harness-assessment.md) - Harness Engineering 状态
- [Harness 迁移计划](/home/doc/rssreader-harness-migration.md) - 完整的改造路线图

## CI/CD

项目使用 GitHub Actions 进行持续集成：

- 自动运行测试
- 代码质量检查 (golangci-lint)
- 构建验证
- 测试覆盖率报告

查看 CI 状态: https://github.com/your-username/rss-reader/actions
