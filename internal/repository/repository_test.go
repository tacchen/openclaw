package repository_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"rss-reader/internal/models"
)

// setupTestDB 创建一个 PostgreSQL 测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	// 从环境变量获取数据库连接串
	// 如果未设置，使用默认的 PostgreSQL Docker 容器
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		// 默认使用 Docker Compose 的 PostgreSQL
		dsn = "postgres://postgres:postgres@localhost:5432/rss_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 清理现有数据 - 删除所有表并重新创建
	tables := []string{"article_tags", "tags", "articles", "feeds", "users"}
	for _, table := range tables {
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
	}

	// 自动迁移所有模型
	if err := db.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{}, &models.Tag{}, &models.ArticleTag{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// createTestUser 创建测试用户
func createTestUser(db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:    email,
		Password: "hashed_password",
	}
	db.Create(user)
	return user
}

// createTestFeed 创建测试订阅源
func createTestFeed(db *gorm.DB, userID uint, title, url, category string) *models.Feed {
	feed := &models.Feed{
		Title:    title,
		URL:      url,
		Category: category,
		UserID:   userID,
	}
	db.Create(feed)
	return feed
}

// createTestArticle 创建测试文章（不插入数据库）
func createTestArticleNoDB(feedID, userID uint, title, link string) *models.Article {
	pubTime := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	return &models.Article{
		Title:     title,
		Link:      link,
		FeedID:    feedID,
		UserID:    userID,
		IsRead:    false,
		PubDate:   &pubTime,
		Summary:   "",
		KeyPoints: "",
	}
}

// createTestArticle 创建测试文章（插入数据库）
func createTestArticle(db *gorm.DB, feedID, userID uint, title, link string) *models.Article {
	article := createTestArticleNoDB(feedID, userID, title, link)
	db.Create(article)
	return article
}

// createTestTag 创建测试标签
func createTestTag(db *gorm.DB, userID uint, name string) *models.Tag {
	tag := &models.Tag{
		Name:   name,
		UserID: userID,
	}
	db.Create(tag)
	return tag
}
