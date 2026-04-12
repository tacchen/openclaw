package models

import "time"

// PushLog 推送日志
type PushLog struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	PushConfigID uint       `gorm:"not null;index" json:"push_config_id"`
	Status       string     `gorm:"not null" json:"status"` // success, failed
	ArticleCount int        `gorm:"not null" json:"article_count"`
	Message      string     `gorm:"type:text" json:"message,omitempty"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       time.Time  `gorm:"index" json:"sent_at"`
}
