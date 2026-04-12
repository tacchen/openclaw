package handlers

import (
	"net/http"
	"strconv"
	"time"

	"rss-reader/internal/models"
	"rss-reader/internal/services"

	"github.com/gin-gonic/gin"
)

// PushConfigRequest 推送配置请求
type PushConfigRequest struct {
	WebhookURL     string   `json:"webhook_url" binding:"required"`
	Frequency      string   `json:"frequency" binding:"required"`
	PushTime       string   `json:"push_time" binding:"required"`
	MinUnreadCount int      `json:"min_unread_count"`
	FeedIDs        []int64  `json:"feed_ids,omitempty"`
	CategoryIDs    []int64  `json:"category_ids,omitempty"`
}

type PushConfig struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	WebhookURL     string     `json:"webhook_url"`
	Frequency      string     `json:"frequency"`
	PushTime       string     `json:"push_time"`
	MinUnreadCount int        `json:"min_unread_count"`
	FeedIDs        []int64    `json:"feed_ids,omitempty"`
	CategoryIDs    []int64    `json:"category_ids,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreatePushConfig 创建推送配置
func CreatePushConfig(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PushConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 从上下文获取用户 ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		modelsConfig := &models.PushConfig{
			WebhookURL:     req.WebhookURL,
			Frequency:      req.Frequency,
			PushTime:       req.PushTime,
			MinUnreadCount: req.MinUnreadCount,
		}
		// 将数组类型复制
		if req.FeedIDs != nil {
			modelsConfig.FeedIDs = make([]int64, len(req.FeedIDs))
			copy(modelsConfig.FeedIDs, req.FeedIDs)
		}
		if req.CategoryIDs != nil {
			modelsConfig.CategoryIDs = make([]int64, len(req.CategoryIDs))
			copy(modelsConfig.CategoryIDs, req.CategoryIDs)
		}

		if err := pushService.CreateConfig(userID.(int), modelsConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, modelsConfig)
	}
}

// GetPushConfigs 获取推送配置列表
func GetPushConfigs(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configs, err := pushService.GetConfigs(userID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, configs)
	}
}

// GetPushConfig 获取单个推送配置
func GetPushConfig(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		config, err := pushService.GetConfig(userID.(int), configID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// UpdatePushConfig 更新推送配置
func UpdatePushConfig(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		var req PushConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := &models.PushConfig{
			WebhookURL:     req.WebhookURL,
			Frequency:      req.Frequency,
			PushTime:       req.PushTime,
			MinUnreadCount: req.MinUnreadCount,
		}
		// 将数组类型复制
		if req.FeedIDs != nil {
			updates.FeedIDs = make([]int64, len(req.FeedIDs))
			copy(updates.FeedIDs, req.FeedIDs)
		}
		if req.CategoryIDs != nil {
			updates.CategoryIDs = make([]int64, len(req.CategoryIDs))
			copy(updates.CategoryIDs, req.CategoryIDs)
		}

		if err := pushService.UpdateConfig(userID.(int), configID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
	}
}

// DeletePushConfig 删除推送配置
func DeletePushConfig(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		if err := pushService.DeleteConfig(userID.(int), configID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
	}
}

// TestPushConfig 测试推送配置
func TestPushConfig(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		if err := pushService.TestConfig(userID.(int), configID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Test push sent successfully"})
	}
}

// GetPushLogs 获取推送日志
func GetPushLogs(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// 分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		// 过滤参数
		filter := make(map[string]interface{})
		if status := c.Query("status"); status != "" {
			filter["status"] = status
		}
		if startDate := c.Query("start_date"); startDate != "" {
			filter["start_date"] = startDate
		}
		if endDate := c.Query("end_date"); endDate != "" {
			filter["end_date"] = endDate
		}

		logs, total, err := pushService.GetPushLogs(userID.(int), filter, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"logs":       logs,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		})
	}
}

// GetPushStats 获取推送统计
func GetPushStats(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		// 验证配置所有权
		_, err = pushService.GetConfig(userID.(int), configID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
			return
		}

		// 时间范围（默认最近 7 天）
		from := time.Now().AddDate(0, 0, -7)
		to := time.Now()

		if fromStr := c.Query("from"); fromStr != "" {
			if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
				from = parsed
			}
		}
		if toStr := c.Query("to"); toStr != "" {
			if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
				to = parsed
			}
		}

		stats, err := pushService.GetStats(configID, from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
