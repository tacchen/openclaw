# 飞书 Webhook 定时推送功能研究

**日期**: 2026-04-11
**功能**: 定时推送未读文章到飞书群

---

## 📋 功能概述

### 目标

实现定时推送未读文章摘要到飞书群，用户可以：

1. 配置飞书 webhook 地址
2. 选择推送频率（每日/每周）
3. 选择推送时间（如每天 9:00）
4. 选择推送内容（未读文章、新文章、订阅源更新）

### 使用场景

- 早晨推送昨夜的新文章
- 每周推送阅读总结
- 重要订阅源更新提醒
- 自定义关键词过滤文章推送

---

## 🔧 技术方案

### 飞书 Webhook 基础

#### Webhook 地址格式

```
https://open.larksuite.com/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx
```

#### 请求方式

```bash
curl -X POST -H "Content-Type: application/json" \
    -d '{"msg_type":"text","content":{"text":"测试消息"}}' \
    https://open.larksuite.com/open-apis/bot/v2/hook/****
```

#### 支持的消息类型

| 类型 | msg_type | 说明 |
|------|----------|------|
| 文本 | `text` | 纯文本消息 |
| 富文本 | `post` | 支持富文本格式 |
| 消息卡片 | `interactive` | 支持卡片、按钮 |
| 群名片 | `share_group` | 分享群组 |

#### 限制

| 限制项 | 值 | 说明 |
|--------|-----|------|
| 频率限制 | 100 次/分钟，5 次/秒 | 避免限流 |
| 请求体大小 | ≤ 20 KB | 注意消息长度 |
| 推荐时间 | 避开整点/半点 | 10:00、17:30 等 |

---

## 🏗️ 系统架构

### 整体流程

```
┌─────────────┐
│  Cron Job  │ (定时触发)
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│  Push Service      │ (推送服务)
│  - 查询未读文章    │
│  - 生成摘要       │
│  - 格式化消息     │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Feishu Client    │ (飞书客户端)
│  - 发送 webhook    │
│  - 重试机制       │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  飞书群          │ (接收消息)
└─────────────────────┘
```

### 数据库设计

#### 新增表：PushConfig

```go
type PushConfig struct {
    ID              uint      `gorm:"primarykey"`
    UserID          uint      `gorm:"index;not null"`     // 用户 ID
    WebhookURL      string    `gorm:"not null"`          // 飞书 webhook 地址
    Frequency       string    `gorm:"not null"`          // 频率：daily/weekly/monthly
    PushTime        string    `gorm:"not null"`          // 推送时间：09:00
    FeedIDs         string    `gorm:"type:text"`          // 订阅源 ID 列表（逗号分隔）
    Category        string    `gorm:"type:text"`          // 分类过滤
    MinUnreadCount  int       `gorm:"default:1"`          // 最少未读数量
    Enabled         bool      `gorm:"default:true"`        // 是否启用
    LastPushAt      *time.Time `gorm:"index"`             // 上次推送时间
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### 新增表：PushLog

```go
type PushLog struct {
    ID          uint      `gorm:"primarykey"`
    ConfigID    uint      `gorm:"index;not null"`  // 推送配置 ID
    Status      string    `gorm:"not null"`        // success/failed
    Message     string    `gorm:"type:text"`        // 推送内容
    Response    string    `gorm:"type:text"`        // 飞书响应
    ErrorMsg    string    `gorm:"type:text"`        // 错误信息
    ArticleCount int       `gorm:"default:0"`       // 推送文章数
    CreatedAt   time.Time `gorm:"index"`
}
```

---

## 💻 实现代码

### 1. Feishu Client

**文件**: `internal/services/feishu_client.go`

```go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type FeishuClient struct {
    httpClient *http.Client
}

func NewFeishuClient() *FeishuClient {
    return &FeishuClient{
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// MessageTypes
const (
    MsgTypeText         = "text"
    MsgTypePost         = "post"
    MsgTypeInteractive  = "interactive"
)

// TextMessage 文本消息
type TextMessage struct {
    MsgType string `json:"msg_type"`
    Content struct {
        Text string `json:"text"`
    } `json:"content"`
}

// SendTextMessage 发送文本消息
func (c *FeishuClient) SendTextMessage(webhookURL, text string) error {
    msg := TextMessage{
        MsgType: MsgTypeText,
    }
    msg.Content.Text = text

    return c.sendMessage(webhookURL, msg)
}

// sendMessage 发送消息
func (c *FeishuClient) sendMessage(webhookURL string, msg interface{}) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal message error: %w", err)
    }

    // 检查消息大小（限制 20KB）
    if len(body) > 20*1024 {
        return fmt.Errorf("message too large: %d bytes (max 20KB)", len(body))
    }

    req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("create request error: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("send request error: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("feishu api error: status %d", resp.StatusCode)
    }

    return nil
}
```

### 2. Push Service

**文件**: `internal/services/push_service.go`

```go
package services

import (
    "fmt"
    "strings"
    "time"

    "rss-reader/internal/models"
    "rss-reader/internal/repository"
)

type PushService struct {
    pushConfigRepo  *repository.PushConfigRepository
    articleRepo     *repository.ArticleRepository
    feedRepo        *repository.FeedRepository
    feishuClient   *FeishuClient
}

func NewPushService(
    pushConfigRepo *repository.PushConfigRepository,
    articleRepo *repository.ArticleRepository,
    feedRepo *repository.FeedRepository,
) *PushService {
    return &PushService{
        pushConfigRepo:  pushConfigRepo,
        articleRepo:     articleRepo,
        feedRepo:        feedRepo,
        feishuClient:   NewFeishuClient(),
    }
}

// SendPush 发送推送
func (s *PushService) SendPush(config *models.PushConfig) error {
    // 查询未读文章
    articles, total, err := s.getUnreadArticles(config)
    if err != nil {
        return err
    }

    // 检查是否达到推送条件
    if int(total) < config.MinUnreadCount {
        return fmt.Errorf("unread count %d < min %d", total, config.MinUnreadCount)
    }

    // 生成消息
    message := s.generateMessage(articles, total)

    // 发送到飞书
    if err := s.feishuClient.SendTextMessage(config.WebhookURL, message); err != nil {
        return err
    }

    // 记录推送日志
    s.logPush(config.ID, message, "success", "", total, len(articles))

    return nil
}

// getUnreadArticles 获取未读文章
func (s *PushService) getUnreadArticles(config *models.PushConfig) ([]models.Article, int64, error) {
    var feedIDs []uint
    if config.FeedIDs != "" {
        // 解析逗号分隔的 feed IDs
        ids := strings.Split(config.FeedIDs, ",")
        for _, idStr := range ids {
            var id uint
            if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
                feedIDs = append(feedIDs, id)
            }
        }
    }

    articles, total, err := s.articleRepo.FindByUserID(
        config.UserID,
        1,                   // page
        20,                  // pageSize
        0,                   // feedID
        0,                   // tagID
        config.Category,      // category
        &true,               // isRead (未读)
        nil,                 // hasSummary
    )

    return articles, total, err
}

// generateMessage 生成推送消息
func (s *PushService) generateMessage(articles []models.Article, total int64) string {
    if len(articles) == 0 {
        return "🎉 暂无新文章"
    }

    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("📰 RSS 每日更新 (%d 篇新文章)\n\n", total))

    // 最多显示前 5 篇文章
    maxArticles := 5
    if len(articles) < maxArticles {
        maxArticles = len(articles)
    }

    for i := 0; i < maxArticles; i++ {
        article := articles[i]
        sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, article.Title))
        
        if article.Summary != "" {
            sb.WriteString(fmt.Sprintf("   %s\n", truncateString(article.Summary, 50)))
        }
        
        if article.Feed != nil {
            sb.WriteString(fmt.Sprintf("   来源: %s\n\n", article.Feed.Title))
        }
    }

    if total > int64(maxArticles) {
        sb.WriteString(fmt.Sprintf("... 还有 %d 篇文章\n", total-int64(maxArticles)))
    }

    return sb.String()
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen]) + "..."
}

// logPush 记录推送日志
func (s *PushService) logPush(configID uint, message, status, response string, total, count int) {
    log := &models.PushLog{
        ConfigID:    configID,
        Status:      status,
        Message:     message,
        Response:    response,
        ArticleCount: count,
    }

    // 保存日志
    s.pushConfigRepo.CreatePushLog(log)
}
```

### 3. Cron Job

**文件**: `backend/cron.go`

```go
package main

import (
    "log"
    "time"

    "github.com/robfig/cron/v3"
    "rss-reader/internal/models"
    "rss-reader/internal/repository"
    "rss-reader/internal/services"
)

// StartPushCron 启动推送定时任务
func StartPushCron(
    pushConfigRepo *repository.PushConfigRepository,
    pushService *services.PushService,
) *cron.Cron {
    c := cron.New()

    // 每天检查一次推送配置
    c.AddFunc("@daily", func() {
        log.Println("Checking push configs...")

        // 获取所有启用的推送配置
        configs, err := pushConfigRepo.FindEnabled()
        if err != nil {
            log.Printf("Error fetching push configs: %v", err)
            return
        }

        // 遍历配置，检查是否需要推送
        now := time.Now()
        for _, config := range configs {
            // 检查推送时间
            pushTime, err := time.Parse("15:04", config.PushTime)
            if err != nil {
                log.Printf("Invalid push time format: %s", config.PushTime)
                continue
            }

            pushTime = time.Date(now.Year(), now.Month(), now.Day(), 
                pushTime.Hour(), pushTime.Minute(), 0, 0, now.Location())

            // 如果当前时间 >= 推送时间 且 今天未推送过
            if now.After(pushTime) || now.Equal(pushTime) {
                // 检查今天是否已推送
                alreadyPushed, err := checkIfPushedToday(pushConfigRepo, config.ID, now)
                if err != nil {
                    log.Printf("Error checking push status: %v", err)
                    continue
                }

                if !alreadyPushed {
                    // 发送推送
                    log.Printf("Sending push for user %d...", config.UserID)
                    if err := pushService.SendPush(&config); err != nil {
                        log.Printf("Error sending push: %v", err)
                    } else {
                        log.Printf("Push sent successfully for user %d", config.UserID)
                    }
                }
            }
        }
    })

    c.Start()
    log.Println("Push cron started")
    return c
}

// checkIfPushedToday 检查今天是否已推送
func checkIfPushedToday(repo *repository.PushConfigRepository, configID uint, now time.Time) (bool, error) {
    logs, err := repo.FindPushLogsByConfigID(configID, 10)
    if err != nil {
        return false, err
    }

    today := now.Format("2006-01-02")
    for _, log := range logs {
        if log.CreatedAt.Format("2006-01-02") == today && log.Status == "success" {
            return true, nil
        }
    }

    return false, nil
}
```

### 4. API Handlers

**文件**: `internal/handlers/push_handlers.go`

```go
package handlers

import (
    "net/http"
    "rss-reader/internal/models"
    "rss-reader/internal/repository"
    "rss-reader/internal/services"

    "github.com/gin-gonic/gin"
)

// CreatePushConfig 创建推送配置
func CreatePushConfig(pushConfigRepo *repository.PushConfigRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")

        var req struct {
            WebhookURL     string `json:"webhook_url" binding:"required,url"`
            Frequency      string `json:"frequency" binding:"required,oneof=daily weekly monthly"`
            PushTime       string `json:"push_time" binding:"required"`
            FeedIDs        string `json:"feed_ids"`
            Category       string `json:"category"`
            MinUnreadCount int    `json:"min_unread_count"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        config := &models.PushConfig{
            UserID:         userID,
            WebhookURL:     req.WebhookURL,
            Frequency:      req.Frequency,
            PushTime:       req.PushTime,
            FeedIDs:        req.FeedIDs,
            Category:       req.Category,
            MinUnreadCount: req.MinUnreadCount,
            Enabled:        true,
        }

        if err := pushConfigRepo.Create(config); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, config)
    }
}

// GetPushConfigs 获取推送配置
func GetPushConfigs(pushConfigRepo *repository.PushConfigRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")

        configs, err := pushConfigRepo.FindByUserID(userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, configs)
    }
}

// UpdatePushConfig 更新推送配置
func UpdatePushConfig(pushConfigRepo *repository.PushConfigRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        id, err := parseUintParam(c, "id")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
            return
        }

        config, err := pushConfigRepo.FindByID(id)
        if err != nil || config.UserID != userID {
            c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
            return
        }

        var req struct {
            WebhookURL     string `json:"webhook_url"`
            Frequency      string `json:"frequency"`
            PushTime       string `json:"push_time"`
            FeedIDs        string `json:"feed_ids"`
            Category       string `json:"category"`
            MinUnreadCount int    `json:"min_unread_count"`
            Enabled        bool   `json:"enabled"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 更新字段
        if req.WebhookURL != "" {
            config.WebhookURL = req.WebhookURL
        }
        if req.Frequency != "" {
            config.Frequency = req.Frequency
        }
        if req.PushTime != "" {
            config.PushTime = req.PushTime
        }
        config.FeedIDs = req.FeedIDs
        config.Category = req.Category
        config.MinUnreadCount = req.MinUnreadCount
        config.Enabled = req.Enabled

        if err := pushConfigRepo.Update(config); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, config)
    }
}

// DeletePushConfig 删除推送配置
func DeletePushConfig(pushConfigRepo *repository.PushConfigRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        id, err := parseUintParam(c, "id")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
            return
        }

        config, err := pushConfigRepo.FindByID(id)
        if err != nil || config.UserID != userID {
            c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
            return
        }

        if err := pushConfigRepo.Delete(id); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Config deleted"})
    }
}

// TestPush 测试推送
func TestPush(pushConfigRepo *repository.PushConfigRepository, pushService *services.PushService) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        id, err := parseUintParam(c, "id")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
            return
        }

        config, err := pushConfigRepo.FindByID(id)
        if err != nil || config.UserID != userID {
            c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
            return
        }

        // 发送测试推送
        if err := pushService.SendPush(config); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Test push sent successfully"})
    }
}
```

---

## 📱 API 端点

### 推送配置 API

| 端点 | 方法 | 描述 | 认证 |
|------|------|------|------|
| `/api/push-configs` | GET | 获取推送配置列表 | ✅ |
| `/api/push-configs` | POST | 创建推送配置 | ✅ |
| `/api/push-configs/:id` | PUT | 更新推送配置 | ✅ |
| `/api/push-configs/:id` | DELETE | 删除推送配置 | ✅ |
| `/api/push-configs/:id/test` | POST | 测试推送 | ✅ |

### 请求示例

**创建推送配置**:

```json
POST /api/push-configs
{
  "webhook_url": "https://open.larksuite.com/open-apis/bot/v2/hook/****",
  "frequency": "daily",
  "push_time": "09:00",
  "feed_ids": "1,2,3",
  "category": "tech,news",
  "min_unread_count": 5
}
```

**响应**:

```json
{
  "id": 1,
  "user_id": 1,
  "webhook_url": "https://open.larksuite.com/open-apis/bot/v2/hook/****",
  "frequency": "daily",
  "push_time": "09:00",
  "feed_ids": "1,2,3",
  "category": "tech,news",
  "min_unread_count": 5,
  "enabled": true,
  "created_at": "2026-04-11T10:00:00Z",
  "updated_at": "2026-04-11T10:00:00Z"
}
```

---

## 🎨 前端实现

### 推送配置页面

**文件**: `frontend/src/views/PushConfig.vue`

```vue
<template>
  <div class="push-config">
    <h2>定时推送配置</h2>
    
    <!-- 配置列表 -->
    <div class="config-list">
      <div v-for="config in configs" :key="config.id" class="config-item">
        <div class="config-info">
          <h3>{{ config.frequency === 'daily' ? '每日推送' : '每周推送' }}</h3>
          <p>推送时间: {{ config.push_time }}</p>
          <p>最少未读数: {{ config.min_unread_count }} 篇</p>
          <p>状态: {{ config.enabled ? '✅ 启用' : '❌ 禁用' }}</p>
        </div>
        <div class="config-actions">
          <button @click="testPush(config.id)">测试推送</button>
          <button @click="editConfig(config.id)">编辑</button>
          <button @click="deleteConfig(config.id)">删除</button>
        </div>
      </div>
    </div>

    <!-- 添加配置按钮 -->
    <button @click="showAddModal = true">添加推送配置</button>

    <!-- 添加/编辑模态框 -->
    <div v-if="showAddModal" class="modal">
      <div class="modal-content">
        <h3>{{ editingId ? '编辑' : '添加' }}推送配置</h3>
        
        <form @submit.prevent="saveConfig">
          <label>
            飞书 Webhook URL
            <input type="text" v-model="form.webhook_url" required>
          </label>
          
          <label>
            推送频率
            <select v-model="form.frequency" required>
              <option value="daily">每日</option>
              <option value="weekly">每周</option>
              <option value="monthly">每月</option>
            </select>
          </label>
          
          <label>
            推送时间
            <input type="time" v-model="form.push_time" required>
          </label>
          
          <label>
            最少未读数
            <input type="number" v-model.number="form.min_unread_count" min="1" required>
          </label>
          
          <label>
            订阅源（留空表示全部）
            <input type="text" v-model="form.feed_ids" placeholder="1,2,3">
          </label>
          
          <label>
            分类（留空表示全部）
            <input type="text" v-model="form.category" placeholder="tech,news">
          </label>
          
          <div class="actions">
            <button type="submit">保存</button>
            <button type="button" @click="showAddModal = false">取消</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      configs: [],
      showAddModal: false,
      editingId: null,
      form: {
        webhook_url: '',
        frequency: 'daily',
        push_time: '09:00',
        feed_ids: '',
        category: '',
        min_unread_count: 5
      }
    }
  },
  async mounted() {
    await this.loadConfigs()
  },
  methods: {
    async loadConfigs() {
      const response = await fetch('/api/push-configs', {
        headers: {
          'Authorization': `Bearer ${this.getToken()}`
        }
      })
      this.configs = await response.json()
    },
    async saveConfig() {
      const url = this.editingId 
        ? `/api/push-configs/${this.editingId}`
        : '/api/push-configs'
      
      const method = this.editingId ? 'PUT' : 'POST'
      
      await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.getToken()}`
        },
        body: JSON.stringify(this.form)
      })
      
      this.showAddModal = false
      this.editingId = null
      await this.loadConfigs()
    },
    async testPush(id) {
      await fetch(`/api/push-configs/${id}/test`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.getToken()}`
        }
      })
      alert('测试推送已发送')
    },
    editConfig(id) {
      const config = this.configs.find(c => c.id === id)
      this.form = { ...config }
      this.editingId = id
      this.showAddModal = true
    },
    async deleteConfig(id) {
      if (confirm('确定要删除这个推送配置吗？')) {
        await fetch(`/api/push-configs/${id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${this.getToken()}`
          }
        })
        await this.loadConfigs()
      }
    },
    getToken() {
      return localStorage.getItem('token')
    }
  }
}
</script>

<style scoped>
.push-config {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.config-list {
  margin-bottom: 20px;
}

.config-item {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 10px;
  display: flex;
  justify-content: space-between;
}

.config-actions button {
  margin-left: 5px;
  padding: 5px 10px;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  max-width: 500px;
  width: 100%;
}

.modal-content label {
  display: block;
  margin-bottom: 10px;
}

.modal-content input,
.modal-content select {
  width: 100%;
  padding: 8px;
  margin-bottom: 5px;
}

.actions {
  margin-top: 15px;
  text-align: right;
}

.actions button {
  margin-left: 10px;
  padding: 8px 15px;
}
</style>
```

---

## 🧪 测试方案

### 1. 单元测试

**测试 Feishu Client**:

```go
func TestFeishuClient_SendTextMessage(t *testing.T) {
    client := NewFeishuClient()
    
    // 使用测试 webhook（需要飞书群）
    webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/test"
    
    err := client.SendTextMessage(webhookURL, "测试消息")
    assert.NoError(t, err)
}

func TestFeishuClient_MessageSizeLimit(t *testing.T) {
    client := NewFeishuClient()
    
    // 测试超过 20KB 的消息
    longText := strings.Repeat("a", 21*1024)
    
    webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/test"
    err := client.SendTextMessage(webhookURL, longText)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "too large")
}
```

### 2. 集成测试

**测试完整推送流程**:

```go
func TestPushService_SendPush(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    pushConfigRepo := repository.NewPushConfigRepository(db)
    articleRepo := repository.NewArticleRepository(db)
    pushService := NewPushService(pushConfigRepo, articleRepo, nil)
    
    // 创建测试用户和配置
    user := createTestUser(db)
    config := &models.PushConfig{
        UserID:     user.ID,
        WebhookURL: "https://open.larksuite.com/open-apis/bot/v2/hook/test",
        Frequency:  "daily",
        PushTime:   "09:00",
    }
    pushConfigRepo.Create(config)
    
    // 执行推送
    err := pushService.SendPush(config)
    
    // Assert
    assert.NoError(t, err)
    
    // 检查日志
    logs, _ := pushConfigRepo.FindPushLogsByConfigID(config.ID, 1)
    assert.Len(t, logs, 1)
    assert.Equal(t, "success", logs[0].Status)
}
```

---

## 📊 工作量估计

| 任务 | 时间 |
|------|------|
| 数据库设计和迁移 | 1 小时 |
| Feishu Client 实现 | 2 小时 |
| Push Service 实现 | 3 小时 |
| Cron Job 集成 | 1 小时 |
| API Handlers | 2 小时 |
| 前端页面开发 | 3 小时 |
| 测试编写 | 2 小时 |
| 调试和优化 | 1 小时 |
| **总计** | **15 小时** |

---

## ⚠️ 注意事项

### 1. 飞书限流

- 避开整点和半点时间（10:00、17:30）
- 频率限制：100 次/分钟，5 次/秒
- 错误码 11232 表示限流

### 2. Webhook 安全

- 不要在公开仓库中泄露 webhook URL
- 建议使用 IP 白名单
- 建议使用签名验证

### 3. 时区处理

- 所有时间存储使用 UTC
- 用户界面显示本地时区
- Cron 任务使用服务器时区

### 4. 错误处理

- 推送失败需要重试机制
- 记录详细的推送日志
- 提供用户手动重试功能

---

## 🚀 后续优化

### 1. 消息卡片

使用飞书消息卡片替代纯文本，支持：

- 点击打开文章
- 图片预览
- 交互按钮（标记已读、删除）

### 2. 智能推送

根据用户阅读习惯智能推荐：

- 推送用户常看的分类
- 推送高权重订阅源
- 避免推送用户已读内容

### 3. 推送统计

提供推送数据分析：

- 推送成功率
- 用户点击率
- 最佳推送时间

---

## 📚 参考资料

- [飞书自定义机器人文档](https://open.larksuite.com/document/client-docs/bot-v3/add-custom-bot?lang=zh-CN)
- [飞书 Webhook API](https://open.larksuite.com/document/client-docs/bot-v3/custom-bot-access)
- [robfig/cron 文档](https://github.com/robfig/cron/v3)

---

**文档创建时间**: 2026-04-11
**文档版本**: 1.0
**状态**: 设计完成，待实现
