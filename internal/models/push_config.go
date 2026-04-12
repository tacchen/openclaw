package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Int64Array 自定义类型，用于 PostgreSQL 整数数组
type Int64Array []int64

// Value 实现 driver.Valuer 接口
func (a Int64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, a)
}

// PushConfig 用户推送配置
type PushConfig struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	UserID         uint       `gorm:"not null;index;uniqueIndex:idx_user_config" json:"user_id"`
	WebhookURL     string     `gorm:"not null" json:"webhook_url"`
	Frequency      string     `gorm:"not null;default:'daily'" json:"frequency"` // daily, weekly, monthly
	PushTime       string     `gorm:"not null;default:'09:00'" json:"push_time"` // HH:MM format
	MinUnreadCount int        `gorm:"default:1" json:"min_unread_count"`
	FeedIDs        Int64Array `gorm:"type:jsonb" json:"feed_ids,omitempty"`
	CategoryIDs    Int64Array `gorm:"type:jsonb" json:"category_ids,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
