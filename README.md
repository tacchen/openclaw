# RSS Reader

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

## 停止服务

```bash
docker-compose down
```
