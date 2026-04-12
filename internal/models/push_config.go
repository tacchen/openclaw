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
	FeedIDs        Int64Array `gorm:"type:jsonb" json:"feed_ids,omitempty"`
	CategoryIDs    Int64Array `gorm:"type:jsonb" json:"category_ids,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
