package repository_test

import (
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"testing"
)

func TestFeedRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")

	feed := &models.Feed{
		Title:    "Test Feed",
		URL:      "https://example.com/feed",
		Category: "tech",
		UserID:   user.ID,
	}

	err := repo.Create(feed)
	if err != nil {
		t.Fatalf("Failed to create feed: %v", err)
	}

	if feed.ID == 0 {
		t.Error("Feed ID should not be 0")
	}

	if feed.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", feed.Title)
	}
}

func TestFeedRepository_FindByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")

	// 创建多个订阅源
	feeds := []*models.Feed{
		{Title: "Feed 1", URL: "https://example.com/feed1", Category: "tech", UserID: user.ID},
		{Title: "Feed 2", URL: "https://example.com/feed2", Category: "news", UserID: user.ID},
		{Title: "Feed 3", URL: "https://example.com/feed3", Category: "tech", UserID: user.ID},
	}

	for _, feed := range feeds {
		db.Create(feed)
	}

	// 查找用户的所有订阅源
	foundFeeds, err := repo.FindByUserID(user.ID)
	if err != nil {
		t.Fatalf("Failed to find feeds by user ID: %v", err)
	}

	if len(foundFeeds) != 3 {
		t.Errorf("Expected 3 feeds, got %d", len(foundFeeds))
	}
}

func TestFeedRepository_FindByUserID_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")

	// 查找用户的订阅源（应该为空）
	feeds, err := repo.FindByUserID(user.ID)
	if err != nil {
		t.Fatalf("Failed to find feeds by user ID: %v", err)
	}

	if len(feeds) != 0 {
		t.Errorf("Expected 0 feeds, got %d", len(feeds))
	}
}

func TestFeedRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")
	createdFeed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 查找订阅源
	foundFeed, err := repo.FindByID(createdFeed.ID)
	if err != nil {
		t.Fatalf("Failed to find feed by ID: %v", err)
	}

	if foundFeed.ID != createdFeed.ID {
		t.Errorf("Expected feed ID %d, got %d", createdFeed.ID, foundFeed.ID)
	}

	if foundFeed.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", foundFeed.Title)
	}
}

func TestFeedRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	// 查找不存在的订阅源
	_, err := repo.FindByID(99999)
	if err == nil {
		t.Error("Expected error when finding non-existent feed")
	}
}

func TestFeedRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 更新订阅源
	feed.Title = "Updated Feed"
	feed.Category = "updated"

	err := repo.Update(feed)
	if err != nil {
		t.Fatalf("Failed to update feed: %v", err)
	}

	// 验证更新
	updatedFeed, err := repo.FindByID(feed.ID)
	if err != nil {
		t.Fatalf("Failed to find updated feed: %v", err)
	}

	if updatedFeed.Title != "Updated Feed" {
		t.Errorf("Expected title 'Updated Feed', got '%s'", updatedFeed.Title)
	}

	if updatedFeed.Category != "updated" {
		t.Errorf("Expected category 'updated', got '%s'", updatedFeed.Category)
	}
}

func TestFeedRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 删除订阅源
	err := repo.Delete(feed.ID)
	if err != nil {
		t.Fatalf("Failed to delete feed: %v", err)
	}

	// 验证删除
	_, err = repo.FindByID(feed.ID)
	if err == nil {
		t.Error("Expected error when finding deleted feed")
	}
}

func TestFeedRepository_Delete_CascadeArticles(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建文章
	createTestArticle(db, feed.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed.ID, user.ID, "Article 2", "https://example.com/article2")

	// 删除订阅源（应该级联删除文章）
	err := repo.Delete(feed.ID)
	if err != nil {
		t.Fatalf("Failed to delete feed: %v", err)
	}

	// 验证文章也被删除
	articles, _, _ := articleRepo.FindByUserID(user.ID, 1, 100, 0, 0, "", nil, nil)
	if len(articles) != 0 {
		t.Errorf("Expected 0 articles after feed deletion, got %d", len(articles))
	}
}

func TestFeedRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user1 := createTestUser(db, "user1@example.com")
	user2 := createTestUser(db, "user2@example.com")

	// 创建订阅源
	createTestFeed(db, user1.ID, "Feed 1", "https://example.com/feed1", "tech")
	createTestFeed(db, user1.ID, "Feed 2", "https://example.com/feed2", "news")
	createTestFeed(db, user2.ID, "Feed 3", "https://example.com/feed3", "tech")

	// 查找所有订阅源
	feeds, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Failed to find all feeds: %v", err)
	}

	if len(feeds) != 3 {
		t.Errorf("Expected 3 feeds, got %d", len(feeds))
	}
}

func TestFeedRepository_FindByURLAndUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user1 := createTestUser(db, "user1@example.com")
	user2 := createTestUser(db, "user2@example.com")

	url := "https://example.com/feed"

	// user1 创建订阅源
	feed1 := createTestFeed(db, user1.ID, "Feed 1", url, "tech")

	// user2 创建相同 URL 的订阅源（应该允许，因为用户不同）
	feed2 := createTestFeed(db, user2.ID, "Feed 2", url, "tech")

	// 验证 user1 可以找到
	foundFeed1, err := repo.FindByURLAndUserID(url, user1.ID)
	if err != nil {
		t.Fatalf("Failed to find feed by URL and user ID: %v", err)
	}
	if foundFeed1.ID != feed1.ID {
		t.Errorf("Expected feed ID %d, got %d", feed1.ID, foundFeed1.ID)
	}

	// 验证 user2 可以找到
	foundFeed2, err := repo.FindByURLAndUserID(url, user2.ID)
	if err != nil {
		t.Fatalf("Failed to find feed by URL and user ID: %v", err)
	}
	if foundFeed2.ID != feed2.ID {
		t.Errorf("Expected feed ID %d, got %d", feed2.ID, foundFeed2.ID)
	}
}

func TestFeedRepository_FindByURLAndUserID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")

	// 查找不存在的订阅源
	_, err := repo.FindByURLAndUserID("https://nonexistent.com/feed", user.ID)
	if err == nil {
		t.Error("Expected error when finding non-existent feed")
	}
}

func TestFeedRepository_DuplicateURLForSameUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFeedRepository(db)

	user := createTestUser(db, "test@example.com")
	url := "https://example.com/feed"

	// 创建第一个订阅源
	feed1 := &models.Feed{
		Title:    "Feed 1",
		URL:      url,
		Category: "tech",
		UserID:   user.ID,
	}
	err := repo.Create(feed1)
	if err != nil {
		t.Fatalf("Failed to create first feed: %v", err)
	}

	// 尝试创建相同 URL 的订阅源（应该失败）
	feed2 := &models.Feed{
		Title:    "Feed 2",
		URL:      url,
		Category: "news",
		UserID:   user.ID,
	}
	err = repo.Create(feed2)
	if err == nil {
		t.Error("Expected error when creating feed with duplicate URL for same user")
	}
}
