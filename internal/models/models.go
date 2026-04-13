package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	Feeds    []Feed `gorm:"foreignKey:UserID" json:"feeds,omitempty"`
}

type Feed struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	URL       string     `gorm:"not null;uniqueIndex:idx_user_feed" json:"url"`
	Title     string     `json:"title"`
	Category  string     `json:"category"`
	IconURL   string     `json:"icon_url,omitempty"`
	UserID    uint       `gorm:"uniqueIndex:idx_user_feed" json:"user_id"`
	LastFetch *time.Time `json:"last_fetch,omitempty"`
	Articles  []Article `gorm:"foreignKey:FeedID" json:"articles,omitempty"`
}

type Article struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	FeedID      uint       `json:"feed_id"`
	Title       string     `gorm:"index;not null" json:"title"`
	Link        string     `json:"link"`
	Description string     `json:"description"`
	Content     string     `json:"content,omitempty"`
	PubDate     *time.Time `json:"pub_date,omitempty"`
	IsRead      bool       `gorm:"default:false" json:"is_read"`
	UserID      uint       `json:"user_id"`
	Tags        []Tag      `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	Summary    string     `json:"summary,omitempty"`
	KeyPoints  string     `json:"key_points,omitempty"`
	Feed        *Feed      `gorm:"foreignKey:FeedID" json:"feed,omitempty"`
}

type Tag struct {
	ID       uint      `gorm:"primarykey" json:"id"`
	Name     string    `json:"name"`
	UserID   uint      `json:"user_id"`
	Articles []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
}

type ArticleTag struct {
	ArticleID uint `json:"article_id"`
	TagID     uint `json:"tag_id"`
}

// PushConfig 用户推送配置
type PushConfig struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	UserID         uint       `gorm:"not null;uniqueIndex:idx_user_config" json:"user_id"`
	WebhookURL     string     `gorm:"not null" json:"webhook_url"`
	Frequency      string     `gorm:"not null;default:'daily'" json:"frequency"`
	PushTime       string     `gorm:"not null;default:'09:00'" json:"push_time"`
	MinUnreadCount int        `gorm:"default:1" json:"min_unread_count"`
	MaxArticleCount int        `gorm:"default:10" json:"max_article_count"`
	FeedIDs        []int64    `gorm:"type:jsonb" json:"feed_ids,omitempty"`
	CategoryIDs    []int64    `gorm:"type:jsonb" json:"category_ids,omitempty"`
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
	SentAt       time.Time `gorm:"column:sent_at;default:CURRENT_TIMESTAMP" json:"sent_at"`
}
