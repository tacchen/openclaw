# API 文档

本文档描述 RSS Reader 的所有 API 端点。

**基础 URL**: `http://localhost:8080`

**认证方式**: JWT Bearer Token（除 `/api/auth/*` 外，所有 API 需要认证）

**请求头**:
```http
Content-Type: application/json
Authorization: Bearer <jwt_token>
```

---

## 认证 API

### POST /api/auth/register

注册新用户。

**请求体**:
```json
{
  "username": "string (必填, 3-50 字符)",
  "password": "string (必填, 6-100 字符)"
}
```

**成功响应**: 201 Created
```json
{
  "message": "User created successfully"
}
```

**错误响应**: 400 Bad Request
```json
{
  "error": "Username already exists"
}
```

---

### POST /api/auth/login

用户登录，返回 JWT Token。

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**成功响应**: 200 OK
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误响应**: 401 Unauthorized
```json
{
  "error": "Invalid username or password"
}
```

**Token 有效期**: 24 小时（可在配置中修改）

---

## Feeds API（需要认证）

### GET /api/feeds

获取当前用户的所有订阅源。

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| category | string | 否 | - | 按分类筛选 |

**请求示例**:
```http
GET /api/feeds?category=tech
```

**成功响应**: 200 OK
```json
[
  {
    "id": 1,
    "title": "Hacker News",
    "url": "https://news.ycombinator.com/rss",
    "category": "tech",
    "created_at": "2026-04-10T00:00:00Z",
    "updated_at": "2026-04-10T00:00:00Z"
  }
]
```

---

### POST /api/feeds

添加新的订阅源。

**请求体**:
```json
{
  "title": "string (必填, 1-255 字符)",
  "url": "string (必填, 1-500 字符, 必须是有效的 RSS URL)",
  "category": "string (可选, 1-100 字符)"
}
```

**成功响应**: 201 Created
```json
{
  "id": 2,
  "title": "TechCrunch",
  "url": "https://techcrunch.com/feed/",
  "category": "news",
  "created_at": "2026-04-10T12:00:00Z",
  "updated_at": "2026-04-10T12:00:00Z"
}
```

**错误响应**: 400 Bad Request
```json
{
  "error": "Invalid RSS URL"
}
```

**注意**: 添加订阅源后会自动触发一次 RSS 抓取

---

### PUT /api/feeds/:id

更新订阅源信息。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Feed ID |

**请求体**:
```json
{
  "title": "string (可选)",
  "url": "string (可选)",
  "category": "string (可选)"
}
```

**成功响应**: 200 OK
```json
{
  "id": 1,
  "title": "Updated Title",
  "url": "https://example.com/feed",
  "category": "updated-category",
  "created_at": "2026-04-10T00:00:00Z",
  "updated_at": "2026-04-10T12:30:00Z"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Feed not found"
}
```

---

### DELETE /api/feeds/:id

删除订阅源（会级联删除该订阅源的所有文章）。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Feed ID |

**成功响应**: 200 OK
```json
{
  "message": "Feed deleted successfully"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Feed not found"
}
```

---

## Articles API（需要认证）

### GET /api/articles

获取文章列表。

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| feed_id | integer | 否 | - | 按订阅源筛选 |
| is_read | boolean | 否 | - | 按已读状态筛选 |
| page | integer | 否 | 1 | 页码 |
| pageSize | integer | 否 | 20 | 每页数量 (1-100) |

**请求示例**:
```http
GET /api/articles?feed_id=1&page=1&pageSize=20
```

**成功响应**: 200 OK
```json
{
  "articles": [
    {
      "id": 1,
      "title": "Example Article Title",
      "url": "https://example.com/article",
      "summary": "Article summary...",
      "content": "Full article content...",
      "published_at": "2026-04-10T10:00:00Z",
      "is_read": false,
      "ai_summary": "AI generated summary...",
      "feed": {
        "id": 1,
        "title": "Hacker News",
        "category": "tech"
      },
      "tags": [
        {
          "id": 1,
          "name": "ai"
        }
      ],
      "created_at": "2026-04-10T10:30:00Z",
      "updated_at": "2026-04-10T10:30:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "pageSize": 20
}
```

---

### GET /api/articles/search

按标题搜索文章。

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| q | string | 是 | - | 搜索关键词 |
| page | integer | 否 | 1 | 页码 |
| pageSize | integer | 否 | 20 | 每页数量 |

**请求示例**:
```http
GET /api/articles/search?q=AI&page=1&pageSize=20
```

**成功响应**: 200 OK（格式同 `/api/articles`）

---

### GET /api/articles/unread-count

获取未读文章数量。

**成功响应**: 200 OK
```json
{
  "unread_count": 42
}
```

---

### PATCH /api/articles/:id/read

标记单篇文章为已读。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Article ID |

**请求体**:
```json
{
  "is_read": true
}
```

**成功响应**: 200 OK
```json
{
  "message": "Article marked as read"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Article not found"
}
```

---

### POST /api/articles/mark-all-read

批量标记所有文章为已读。

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| feed_id | integer | 否 | - | 只标记指定订阅源的文章 |

**请求示例**:
```http
POST /api/articles/mark-all-read?feed_id=1
```

**成功响应**: 200 OK
```json
{
  "message": "All articles marked as read",
  "affected_count": 100
}
```

---

## AI Summary API（需要认证）

### POST /api/articles/:id/summary

为文章生成 AI 摘要（使用 OpenAI API）。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Article ID |

**成功响应**: 200 OK
```json
{
  "id": 1,
  "ai_summary": "AI generated summary of the article..."
}
```

**错误响应**:
- 404 Not Found: 文章不存在
- 500 Internal Server Error: OpenAI API 调用失败

**注意**: 如果未配置 `OPENAI_API_KEY`，此功能不可用

---

### GET /api/articles/:id/summary

获取文章的 AI 摘要（如果已生成）。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Article ID |

**成功响应**: 200 OK
```json
{
  "ai_summary": "Previously generated summary..."
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Summary not found"
}
```

---

## Tags API（需要认证）

### GET /api/tags

获取当前用户的所有标签。

**成功响应**: 200 OK
```json
[
  {
    "id": 1,
    "name": "ai",
    "created_at": "2026-04-10T00:00:00Z"
  },
  {
    "id": 2,
    "name": "startup",
    "created_at": "2026-04-10T00:00:00Z"
  }
]
```

---

### POST /api/tags

创建新标签。

**请求体**:
```json
{
  "name": "string (必填, 1-100 字符)"
}
```

**成功响应**: 201 Created
```json
{
  "id": 3,
  "name": "new-tag",
  "created_at": "2026-04-10T12:00:00Z"
}
```

**错误响应**: 400 Bad Request
```json
{
  "error": "Tag name already exists"
}
```

---

### DELETE /api/tags/:id

删除标签（会自动移除所有文章的该标签）。

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | Tag ID |

**成功响应**: 200 OK
```json
{
  "message": "Tag deleted successfully"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Tag not found"
}
```

---

## Article Tags API（需要认证）

### POST /api/articles/tags

为文章添加标签。

**请求体**:
```json
{
  "article_id": "integer (必填)",
  "tag_id": "integer (必填)"
}
```

**成功响应**: 200 OK
```json
{
  "message": "Tag added to article"
}
```

**错误响应**: 400 Bad Request
```json
{
  "error": "Tag already added to article"
}
```

---

### DELETE /api/articles/tags

移除文章的标签。

**请求体**:
```json
{
  "article_id": "integer (必填)",
  "tag_id": "integer (必填)"
}
```

**成功响应**: 200 OK
```json
{
  "message": "Tag removed from article"
}
```

---

## 错误响应格式

所有错误响应遵循统一格式：

```json
{
  "error": "错误描述信息"
}
```

### 常见错误码

| 状态码 | 含义 | 示例 |
|--------|------|------|
| 400 | Bad Request | 参数错误、请求格式错误 |
| 401 | Unauthorized | 未认证、Token 无效或过期 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源已存在（如用户名、订阅源） |
| 500 | Internal Server Error | 服务器内部错误 |

---

## 分页响应格式

分页响应包含元数据：

```json
{
  "articles": [...],
  "total": 100,
  "page": 1,
  "pageSize": 20
}
```

**计算总页数**:
```javascript
const totalPages = Math.ceil(total / pageSize);
```

---

## 数据类型

### User（用户）
```json
{
  "id": 1,
  "username": "string",
  "created_at": "2026-04-10T00:00:00Z",
  "updated_at": "2026-04-10T00:00:00Z"
}
```

### Feed（订阅源）
```json
{
  "id": 1,
  "title": "string",
  "url": "string",
  "category": "string",
  "user_id": 1,
  "created_at": "2026-04-10T00:00:00Z",
  "updated_at": "2026-04-10T00:00:00Z"
}
```

### Article（文章）
```json
{
  "id": 1,
  "title": "string",
  "url": "string",
  "summary": "string",
  "content": "string",
  "published_at": "2026-04-10T00:00:00Z",
  "is_read": false,
  "ai_summary": "string",
  "feed_id": 1,
  "user_id": 1,
  "created_at": "2026-04-10T00:00:00Z",
  "updated_at": "2026-04-10T00:00:00Z"
}
```

### Tag（标签）
```json
{
  "id": 1,
  "name": "string",
  "user_id": 1,
  "created_at": "2026-04-10T00:00:00Z"
}
```

---

## 使用示例

### JavaScript (Fetch)

```javascript
// 1. 登录
const loginResponse = await fetch('http://localhost:8080/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'password123',
  }),
});

const { token } = await loginResponse.json();

// 2. 获取订阅源
const feedsResponse = await fetch('http://localhost:8080/api/feeds', {
  headers: {
    'Authorization': `Bearer ${token}`,
  },
});

const feeds = await feedsResponse.json();

// 3. 添加订阅源
const addFeedResponse = await fetch('http://localhost:8080/api/feeds', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  },
  body: JSON.stringify({
    title: 'TechCrunch',
    url: 'https://techcrunch.com/feed/',
    category: 'news',
  }),
});

const newFeed = await addFeedResponse.json();
```

### cURL

```bash
# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# 获取订阅源
curl http://localhost:8080/api/feeds \
  -H "Authorization: Bearer <token>"

# 添加订阅源
curl -X POST http://localhost:8080/api/feeds \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"TechCrunch","url":"https://techcrunch.com/feed/","category":"news"}'
```

---

## 最后更新

**更新日期**: 2026-04-10
**API 版本**: v1.0

---

**注意事项**:

1. 所有时间戳使用 ISO 8601 格式（UTC）
2. Token 过期时间：24 小时
3. RSS 抓取间隔：30 分钟（自动任务）
4. AI 摘要功能需要配置 `OPENAI_API_KEY`
5. 单个用户最多可以添加 100 个订阅源
6. 分页 `pageSize` 最大值为 100

---

## 推送配置 API（需要认证）

### POST /api/push-configs

创建推送配置。

**请求体**:
```json
{
  "webhook_url": "string (必填)",
  "frequency": "string (必填, daily/weekly/monthly)",
  "push_time": "string (必填, HH:MM 格式)",
  "min_unread_count": "integer (可选, 默认 1)",
  "feed_ids": "[integer] (可选, 空数组表示全部)",
  "category_ids": "[integer] (可选, 空数组表示全部)"
}
```

**成功响应**: 201 Created
```json
{
  "id": 1,
  "user_id": 1,
  "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
  "frequency": "daily",
  "push_time": "09:00",
  "min_unread_count": 5,
  "feed_ids": [1, 2, 3],
  "category_ids": [],
  "last_push_at": null,
  "created_at": "2026-04-12T11:00:00Z",
  "updated_at": "2026-04-12T11:00:00Z"
}
```

**错误响应**: 400 Bad Request
```json
{
  "error": "Invalid frequency: hourly"
}
```

---

### GET /api/push-configs

获取当前用户的所有推送配置。

**成功响应**: 200 OK
```json
[
  {
    "id": 1,
    "user_id": 1,
    "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    "frequency": "daily",
    "push_time": "09:00",
    "min_unread_count": 5,
    "feed_ids": [1, 2, 3],
    "category_ids": [],
    "last_push_at": "2026-04-12T09:00:00Z",
    "created_at": "2026-04-12T11:00:00Z",
    "updated_at": "2026-04-12T11:00:00Z"
  }
]
```

---

### GET /api/push-configs/:id

获取指定推送配置。

**URL 参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 推送配置 ID |

**成功响应**: 200 OK
```json
{
  "id": 1,
  "user_id": 1,
  "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
  "frequency": "daily",
  "push_time": "09:00",
  "min_unread_count": 5,
  "feed_ids": [1, 2, 3],
  "category_ids": [],
  "last_push_at": "2026-04-12T09:00:00Z",
  "created_at": "2026-04-12T11:00:00Z",
  "updated_at": "2026-04-12T11:00:00Z"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Config not found"
}
```

---

### PUT /api/push-configs/:id

更新推送配置。

**URL 参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 推送配置 ID |

**请求体**:
```json
{
  "webhook_url": "string (可选)",
  "frequency": "string (可选)",
  "push_time": "string (可选)",
  "min_unread_count": "integer (可选)",
  "feed_ids": "[integer] (可选)",
  "category_ids": "[integer] (可选)"
}
```

**成功响应**: 200 OK
```json
{
  "message": "Config updated successfully"
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Config not found"
}
```

---

### DELETE /api/push-configs/:id

删除推送配置。

**URL 参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 推送配置 ID |

**成功响应**: 200 OK
```json
{
  "message": "Config deleted successfully"
}
```

---

### POST /api/push-configs/:id/test

测试推送配置。

**URL 参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 推送配置 ID |

**成功响应**: 200 OK
```json
{
  "message": "Test push sent successfully"
}
```

**测试消息格式**:
```
🔔 推送配置测试

这是一条测试消息，如果您收到此消息，说明推送配置正常。
```

---

### GET /api/push-logs

获取推送日志。

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | integer | 否 | 1 | 页码 |
| page_size | integer | 否 | 20 | 每页数量（最大 100） |
| status | string | 否 | - | 过滤状态（success/failed） |
| start_date | string | 否 | - | 开始日期（RFC3339 格式） |
| end_date | string | 否 | - | 结束日期（RFC3339 格式） |

**请求示例**:
```http
GET /api/push-logs?page=1&page_size=20&status=success&start_date=2026-04-01T00:00:00Z
```

**成功响应**: 200 OK
```json
{
  "logs": [
    {
      "id": 1,
      "user_id": 1,
      "push_config_id": 1,
      "status": "success",
      "article_count": 10,
      "message": "Push sent successfully",
      "error_message": null,
      "sent_at": "2026-04-12T09:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

---

### GET /api/push-configs/:id/stats

获取推送统计。

**URL 参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 推送配置 ID |

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| from | string | 否 | 7 天前 | 开始日期（RFC3339 格式） |
| to | string | 否 | 现在 | 结束日期（RFC3339 格式） |

**请求示例**:
```http
GET /api/push-configs/1/stats?from=2026-04-01T00:00:00Z&to=2026-04-12T23:59:59Z
```

**成功响应**: 200 OK
```json
{
  "total_pushes": 14,
  "success_count": 12,
  "failed_count": 2,
  "success_rate": 85.71,
  "avg_article_count": 8.5
}
```

**错误响应**: 404 Not Found
```json
{
  "error": "Config not found"
}
```

---

## 推送测试 API（需要认证）

### POST /api/push/test

手动触发每日汇总推送（测试用）。

**成功响应**: 200 OK
```json
{
  "message": "Test push sent successfully"
}
```

**注意**: 此接口用于测试全局推送功能，会根据所有用户的配置发送推送。
