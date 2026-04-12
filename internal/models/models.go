package models

import (
	"time"

	"gorm.io/gorm"
)

// PushConfig 用户推送配置
type PushConfig struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	UserID         uint       `gorm:"not null;uniqueIndex:idx_user_config" json:"user_id"`
	WebhookURL     string     `gorm:"not null" json:"webhook_url"`
	Frequency      string     `gorm:"not null;default:'daily'" json:"frequency"`
	PushTime       string     `gorm:"not null;default:'09:00'" json:"push_time"`
	MinUnreadCount int        `gorm:"default:1" json:"min_unread_count"`
	FeedIDs        string     `gorm:"type:jsonb" json:"feed_ids,omitempty"`
	CategoryIDs    string     `gorm:"type:jsonb" json:"category_ids,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// PushLog 推送日志
type PushLog struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	PushConfigID uint      `gorm:"not null;index" json:"push_config_id"`
	Status       string    `gorm:"not null" json:"status"`
	ArticleCount int       `gorm:"not null" json:"article_count"`
	Message      string    `gorm:"type:text" json:"message,omitempty"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       time.Time `gorm:"autoCreateTime" json:"sent_at"`
}
