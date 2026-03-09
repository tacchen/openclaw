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
	ID        uint       `gorm:"primarykey" json:"id"`
	Name      string     `json:"name"`
	UserID    uint       `json:"user_id"`
	Articles  []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
}

type ArticleTag struct {
	ArticleID uint `json:"article_id"`
	TagID     uint `json:"tag_id"`
}
