package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rss-reader/internal/models"

	"gorm.io/gorm"
)

// PushService 推送服务
type PushService struct {
	db           *gorm.DB
	feishuClient *FeishuClient
}

// NewPushService 创建推送服务
func NewPushService(db *gorm.DB, feishuClient *FeishuClient) *PushService {
	ps := &PushService{
		feishuClient: feishuClient,
	}
	ps.db = db
	return ps
}

// CreateConfig 创建用户推送配置
func (s *PushService) CreateConfig(userID uint, config *models.PushConfig) error {
	config.UserID = userID
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Initialize nil arrays to empty slices
	if config.FeedIDs == nil {
		config.FeedIDs = []int64{}
	}
	if config.CategoryIDs == nil {
		config.CategoryIDs = []int64{}
	}

	// 验证频率
	validFrequencies := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validFrequencies[config.Frequency] {
		return fmt.Errorf("invalid frequency: %s", config.Frequency)
	}

	// 验证推送时间格式
	if _, err := time.Parse("15:04", config.PushTime); err != nil {
		return fmt.Errorf("invalid push time format: %s (expected HH:MM)", config.PushTime)
	}

	// 验证最小未读数
	if config.MinUnreadCount < 0 {
		return fmt.Errorf("min_unread_count must be >= 0")
	}

	return s.db.Create(config).Error
}

// GetConfigs 获取用户推送配置
func (s *PushService) GetConfigs(userID uint) ([]models.PushConfig, error) {
	var configs []models.PushConfig
	err := s.db.Where("user_id = ?", userID).Find(&configs).Error
	return configs, err
}

// GetConfig 获取指定配置
func (s *PushService) GetConfig(userID uint, configID int) (*models.PushConfig, error) {
	var config models.PushConfig
	err := s.db.Where("id = ? AND user_id = ?", configID, userID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfig 更新用户推送配置
func (s *PushService) UpdateConfig(userID uint, configID int, updates *models.PushConfig) error {
	// 验证配置所有权
	var existing models.PushConfig
	if err := s.db.Where("id = ? AND user_id = ?", configID, userID).First(&existing).Error; err != nil {
		return fmt.Errorf("config not found or access denied")
	}

	// 验证频率
	if updates.Frequency != "" {
		validFrequencies := map[string]bool{"daily": true, "weekly": true, "monthly": true}
		if !validFrequencies[updates.Frequency] {
			return fmt.Errorf("invalid frequency: %s", updates.Frequency)
		}
	}

	// 验证推送时间格式
	if updates.PushTime != "" {
		if _, err := time.Parse("15:04", updates.PushTime); err != nil {
			return fmt.Errorf("invalid push time format: %s (expected HH:MM)", updates.PushTime)
		}
	}

	// 验证最小未读数
	if updates.MinUnreadCount < 0 {
		return fmt.Errorf("min_unread_count must be >= 0")
	}

	updates.UpdatedAt = time.Now()

	return s.db.Model(&existing).Updates(updates).Error
}

// DeleteConfig 删除用户推送配置
func (s *PushService) DeleteConfig(userID uint, configID int) error {
	// 验证配置所有权
	var existing models.PushConfig
	if err := s.db.Where("id = ? AND user_id = ?", configID, userID).First(&existing).Error; err != nil {
		return fmt.Errorf("config not found or access denied")
	}

	return s.db.Delete(&existing).Error
}

// TestConfig 测试推送配置
func (s *PushService) TestConfig(userID uint, configID int) error {
	config, err := s.GetConfig(userID, configID)
	if err != nil {
		return err
	}

	// 创建临时 FeishuClient
	testClient := NewFeishuClient(config.WebhookURL)
	message := "🔔 推送配置测试\n\n这是一条测试消息，如果您收到此消息，说明推送配置正常。"

	return testClient.SendTextMessage(message)
}

// LogPush 记录推送日志
func (s *PushService) LogPush(log *models.PushLog) error {
	log.SentAt = time.Now()
	return s.db.Create(log).Error
}

// GetPushLogs 获取推送日志（分页、过滤）
func (s *PushService) GetPushLogs(userID uint, filter map[string]interface{}, page, pageSize int) ([]models.PushLog, int64, error) {
	var logs []models.PushLog
	var total int64

	query := s.db.Model(&models.PushLog{}).Where("user_id = ?", userID)

	// 应用过滤
	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if startDate, ok := filter["start_date"]; ok {
		query = query.Where("sent_at >= ?", startDate)
	}
	if endDate, ok := filter["end_date"]; ok {
		query = query.Where("sent_at <= ?", endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sent_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetStats 获取推送统计
type PushStats struct {
	TotalPushes   int64   `json:"total_pushes"`
	SuccessCount  int64   `json:"success_count"`
	FailedCount   int64   `json:"failed_count"`
	SuccessRate   float64 `json:"success_rate"`
	AvgArticleCount float64 `json:"avg_article_count"`
}

func (s *PushService) GetStats(configID int, from, to time.Time) (*PushStats, error) {
	var stats PushStats

	// 总推送数
	s.db.Model(&models.PushLog{}).
		Where("push_config_id = ? AND sent_at BETWEEN ? AND ?", configID, from, to).
		Count(&stats.TotalPushes)

	// 成功次数
	s.db.Model(&models.PushLog{}).
		Where("push_config_id = ? AND status = 'success' AND sent_at BETWEEN ? AND ?", configID, from, to).
		Count(&stats.SuccessCount)

	// 失败次数
	s.db.Model(&models.PushLog{}).
		Where("push_config_id = ? AND status = 'failed' AND sent_at BETWEEN ? AND ?", configID, from, to).
		Count(&stats.FailedCount)

	// 计算成功率
	if stats.TotalPushes > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalPushes) * 100
	}

	// 平均文章数
	var avgArticleCount sql.NullFloat64
	s.db.Model(&models.PushLog{}).
		Where("push_config_id = ? AND sent_at BETWEEN ? AND ?", configID, from, to).
		Select("AVG(article_count)").
		Scan(&avgArticleCount)

	if avgArticleCount.Valid {
		stats.AvgArticleCount = avgArticleCount.Float64
	}

	return &stats, nil
}

// ProcessDailyPushes 处理每日推送
func (s *PushService) ProcessDailyPushes() error {
	// 查询所有每日推送配置
	var configs []models.PushConfig
	now := time.Now()
	currentTime := now.Format("15:04")

	if err := s.db.Where("frequency = ? AND push_time = ?", "daily", currentTime).Find(&configs).Error; err != nil {
		return fmt.Errorf("query daily configs error: %w", err)
	}

	// 处理每个配置
	for _, config := range configs {
		if err := s.processConfig(&config); err != nil {
			// 记录失败日志，但继续处理其他配置
			s.LogPush(&models.PushLog{
				UserID:       config.UserID,
				PushConfigID: config.ID,
				Status:       "failed",
				ArticleCount: 0,
				ErrorMessage: err.Error(),
			})
		}
	}

	return nil
}

// ProcessWeeklyPushes 处理每周推送
func (s *PushService) ProcessWeeklyPushes() error {
	// 查询所有每周推送配置
	var configs []models.PushConfig
	if err := s.db.Where("frequency = ?", "weekly").Find(&configs).Error; err != nil {
		return fmt.Errorf("query weekly configs error: %w", err)
	}

	// 过滤出今天应该推送的配置
	now := time.Now()
	currentTime := now.Format("15:04")
	weekday := int(now.Weekday()) // Sunday = 0, Monday = 1, etc.

	// 将 Go 的星期（0=Sunday）转换为常规的（1=Monday, 7=Sunday）
	weekNum := weekday
	if weekNum == 0 {
		weekNum = 7
	}

	// 处理每个配置
	for _, config := range configs {
		// 检查推送时间是否匹配
		if config.PushTime != currentTime {
			continue
		}

		// TODO: 这里需要添加星期几的过滤逻辑
		// 当前 PushConfig 模型没有 weekday 字段，所以暂时跳过
		// 未来可以添加 weekday 字段来实现精确的每周推送
		if err := s.processConfig(&config); err != nil {
			s.LogPush(&models.PushLog{
				UserID:       config.UserID,
				PushConfigID: config.ID,
				Status:       "failed",
				ArticleCount: 0,
				ErrorMessage: err.Error(),
			})
		}
	}

	return nil
}

// processConfig 处理单个配置的推送
func (s *PushService) processConfig(config *models.PushConfig) error {
	// 查询未读文章
	var articles []models.Article
	query := s.db.Where("user_id = ? AND is_read = false", config.UserID)

	// 应用订阅源过滤
	if len(config.FeedIDs) > 0 {
		query = query.Where("feed_id IN ?", config.FeedIDs)
	}

	// 应用分类过滤
	if len(config.CategoryIDs) > 0 {
		// 需要关联 feeds 表
		query = query.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.category IN ?", s.getCategoriesFromIDs(config.CategoryIDs))
	}

	// 检查最小未读数
	var totalCount int64
	query.Count(&totalCount)
	if totalCount < int64(config.MinUnreadCount) {
		return fmt.Errorf("not enough unread articles: %d (min: %d)", totalCount, config.MinUnreadCount)
	}

	// 限制文章数量（避免消息过大）
	query = query.Order("pub_date DESC").Limit(20)
	if err := query.Find(&articles).Error; err != nil {
		return fmt.Errorf("query articles error: %w", err)
	}

	// 发送推送
	message := s.formatPushMessage(articles)
	client := NewFeishuClient(config.WebhookURL)
	if err := client.SendTextMessage(message); err != nil {
		return err
	}

	// 记录成功日志
	s.LogPush(&models.PushLog{
		UserID:       config.UserID,
		PushConfigID: config.ID,
		Status:       "success",
		ArticleCount: len(articles),
		Message:      "Push sent successfully",
	})

	// 更新最后推送时间
	s.db.Model(config).Update("last_push_at", time.Now())

	return nil
}

// formatPushMessage 格式化推送消息
func (s *PushService) formatPushMessage(articles []models.Article) string {
	var builder strings.Builder
	builder.WriteString("📰 每日文章汇总\n\n")

	for i, article := range articles {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, article.Title))

		if article.Feed != nil {
			builder.WriteString(fmt.Sprintf("   来源：%s\n", article.Feed.Title))
		}

		if article.PubDate != nil {
			builder.WriteString(fmt.Sprintf("   时间：%s\n", article.PubDate.Format("2006-01-02 15:04")))
		}

		if article.Description != "" && len(article.Description) > 0 {
			maxDescLen := 50
			desc := article.Description
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen] + "..."
			}
			builder.WriteString(fmt.Sprintf("   描述：%s\n", desc))
		}

		builder.WriteString(fmt.Sprintf("   链接：%s\n", article.Link))
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("共 %d 篇文章", len(articles)))
	return builder.String()
}

// getCategoriesFromIDs 从分类 ID 获取分类名称（待实现）
func (s *PushService) getCategoriesFromIDs(ids []int64) []string {
	// 这里需要根据分类 ID 查询分类名称
	// 为了简化，这里直接返回 nil（不过滤）
	return nil
}

// SendDailySummary 发送每日汇总（兼容旧的全局推送功能）
func (s *PushService) SendDailySummary() error {
	// 获取所有用户的推送配置
	var configs []models.PushConfig
	now := time.Now()
	currentTime := now.Format("15:04")

	if err := s.db.Where("frequency = ? AND push_time = ?", "daily", currentTime).Find(&configs).Error; err != nil {
		return fmt.Errorf("query daily configs error: %w", err)
	}

	// 处理每个配置
	for _, config := range configs {
		if err := s.processConfig(&config); err != nil {
			s.LogPush(&models.PushLog{
				UserID:       config.UserID,
				PushConfigID: config.ID,
				Status:       "failed",
				ArticleCount: 0,
				ErrorMessage: err.Error(),
			})
		}
	}

	return nil
}
