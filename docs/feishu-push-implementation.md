# 飞书 Webhook 推送功能 - 使用说明

**功能**: 当 RSS 订阅源抓取到新文章时，自动推送到飞书群

---

## 🚀 快速开始

### 1. 配置飞书 Webhook URL

在 `.env` 文件中设置飞书 webhook URL：

```bash
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/968f7d62-b536-4ee5-a5d5-839931519065
```

### 2. 启动服务

```bash
cd /home/prj/rss-reader
./backend/main
```

服务启动后会输出：

```
Feishu webhook configured: https://open.feishu.cn/open-apis/bot/v2/hook/...9931519065
Feishu webhook enabled: https://open.feishu.cn/open-apis/bot/v2/hook/...9931519065
Server starting on port 8080...
```

### 3. 添加订阅源

添加一个新的 RSS 订阅源，服务会立即抓取：

```bash
curl -X POST http://localhost:8080/api/feeds \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.zhihu.com/rss",
    "title": "知乎热榜",
    "category": "tech"
  }'
```

### 4. 等待推送

当抓取到新文章时，会自动发送到飞书群：

```
📰 新文章推送

标题：xxx
来源：知乎热榜
描述：xxx
链接：https://...
```

---

## 📋 功能特性

### ✅ 已实现

- [x] 飞书 Webhook 客户端
- [x] 新文章自动推送
- [x] 格式化消息（标题、来源、描述、链接）
- [x] Webhook URL 隐藏（日志中不泄露完整 URL）
- [x] 可选配置（不设置不影响其他功能）

### 📝 消息格式

推送消息包含以下信息：

```
📰 新文章推送

标题：[文章标题]
来源：[订阅源标题]
描述：[文章描述，最多100字]
链接：[文章链接]
```

---

## ⚙️ 配置说明

### 环境变量

| 变量 | 必填 | 说明 | 默认值 |
|------|------|------|--------|
| `FEISHU_WEBHOOK_URL` | 否 | 飞书群 webhook 地址 | 空（禁用推送）|

### 示例

```bash
# .env 文件
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx
```

---

## 🔧 代码结构

### 新增文件

| 文件 | 说明 |
|------|------|
| `internal/services/feishu_client.go` | 飞书客户端，封装 webhook API |
| `.env.example` | 配置文件示例 |

### 修改文件

| 文件 | 修改内容 |
|------|---------|
| `internal/services/rss.go` | 添加飞书推送逻辑 |
| `internal/config/config.go` | 添加 webhook URL 配置 |
| `backend/main.go` | 初始化飞书客户端 |
| `.env` | 添加 webhook URL |

---

## 🧪 测试

### 手动测试

1. 启动服务
2. 添加一个新的订阅源
3. 查看飞书群是否收到推送

### 自动测试

RSS 抓取定时任务（每 30 分钟）会自动检查新文章并推送。

---

## ⚠️ 注意事项

### 1. 飞书限流

- **频率限制**: 100 次/分钟，5 次/秒
- **消息大小**: ≤ 20 KB
- **建议**: 避开整点和半点（10:00、17:30 等）

### 2. Webhook 安全

- **不要泄露**: 不要在公开仓库中提交 webhook URL
- **建议配置**:
  - IP 白名单
  - 自定义关键词
  - 签名验证

### 3. 推送策略

- **只推送新文章**: 已存在的文章不会重复推送
- **批量推送**: 多个订阅源有新文章时会分别推送
- **失败重试**: 目前单次尝试，失败会记录日志

---

## 📊 推送效果

### 成功场景

- ✅ 添加新订阅源，立即推送新文章
- ✅ 定时抓取（30 分钟），推送新文章
- ✅ 格式化消息，包含完整信息

### 失败场景

- ❌ Webhook URL 错误 → 记录错误日志
- ❌ 飞书限流 → 记录错误日志（错误码 11232）
- ❌ 消息过大 → 记录错误日志（> 20 KB）

---

## 🚀 后续优化

### 短期

1. **批量推送**: 多篇文章合并为一条消息
2. **推送去重**: 避免同一文章多次推送
3. **错误重试**: 失败后自动重试

### 中期

1. **消息卡片**: 使用飞书消息卡片（支持点击、按钮）
2. **推送过滤**: 支持按分类、订阅源过滤
3. **推送日志**: 记录推送历史和状态

### 长期

1. **智能推送**: 根据阅读习惯推荐内容
2. **推送统计**: 推送成功率、点击率
3. **用户配置**: 前端页面配置推送规则

---

## 🐛 故障排查

### 问题 1: 没有收到推送

**检查步骤**:

1. 检查 `.env` 文件中是否设置了 `FEISHU_WEBHOOK_URL`
2. 检查服务日志中是否有 `Feishu webhook enabled` 输出
3. 检查飞书机器人是否还在群中
4. 检查 webhook URL 是否正确

### 问题 2: 推送格式错误

**检查步骤**:

1. 查看服务日志中的错误信息
2. 检查飞书群设置（安全配置）
3. 检查 webhook URL 是否匹配

### 问题 3: 频繁限流

**解决方案**:

1. 修改 cron 表达式，避免整点推送
2. 减少推送频率（如改为每小时一次）
3. 使用批量推送减少请求数

---

## 📄 参考资料

- [飞书自定义机器人文档](https://open.larksuite.com/document/client-docs/bot-v3/add-custom-bot?lang=zh-CN)
- [飞书 Webhook API](https://open.larksuite.com/document/client-docs/bot-v3/custom-bot-access)

---

**文档创建时间**: 2026-04-11
**功能状态**: ✅ 已实现并测试
**当前配置**: 已启用，Webhook URL: https://open.feishu.cn/open-apis/bot/v2/hook/968f7d62-b536-4ee5-a5d5-839931519065
