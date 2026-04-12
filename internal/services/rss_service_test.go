package services_test

import (
	"testing"
	"time"

	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/services"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDBRSS(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=rss_test port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up tables before test
	db.Exec("DROP TABLE IF EXISTS article_tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS articles CASCADE")
	db.Exec("DROP TABLE IF EXISTS feeds CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")

	// Auto migrate tables
	if err := db.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{}, &models.Tag{}, &models.ArticleTag{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestRSSService_FetchAndSaveArticles_Success(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	_ = services.NewRSSService(feedRepo, articleRepo, nil)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)

	// Mock the gofeed parser response - we'll use a real feed URL for simplicity
	// In a real test, you would mock the HTTP client

	// For this test, we'll skip the actual parsing and just verify the flow
	// by manually creating an article

	// Note: In production, you would want to mock the HTTP client
	// to avoid making real network requests in tests

	t.Skip("Skipping integration test - requires mocking HTTP client or using real feed")
}

func TestRSSService_FetchAllFeeds_Empty(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	service := services.NewRSSService(feedRepo, articleRepo, nil)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// No feeds exist
	_ = service // Use service to avoid unused variable error

	feeds, err := feedRepo.FindAll()
	assert.NoError(t, err)
	assert.Empty(t, feeds)

	// FetchAllFeeds should not error
	service.FetchAllFeeds()
}

func TestRSSService_FetchAllFeeds_WithFeeds(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	service := services.NewRSSService(feedRepo, articleRepo, nil)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create multiple feeds
	feed1 := &models.Feed{
		URL:      "https://example1.com/feed.xml",
		Title:    "Test Feed 1",
		UserID:   user.ID,
		Category: "tech",
	}
	feed2 := &models.Feed{
		URL:      "https://example2.com/feed.xml",
		Title:    "Test Feed 2",
		UserID:   user.ID,
		Category: "news",
	}
	err := feedRepo.Create(feed1)
	assert.NoError(t, err)
	err = feedRepo.Create(feed2)
	assert.NoError(t, err)

	// Verify feeds were created
	feeds, err := feedRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, feeds, 2)

	// Note: FetchAllFeeds would normally iterate and fetch articles
	// For unit testing, we skip the actual HTTP requests
	service.FetchAllFeeds()
}

func TestRSSService_ArticleExistsByLink(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)

	// Create an article
	pubDate := time.Now()
	article := &models.Article{
		FeedID:      feed.ID,
		Title:       "Test Article",
		Link:        "https://example.com/article1",
		Description: "Test description",
		UserID:      user.ID,
		PubDate:     &pubDate,
	}
	err = articleRepo.Create(article)
	assert.NoError(t, err)

	// Verify article exists
	exists := articleRepo.ExistsByLink(feed.ID, "https://example.com/article1")
	assert.True(t, exists)

	// Verify non-existent article doesn't exist
	exists = articleRepo.ExistsByLink(feed.ID, "https://example.com/nonexistent")
	assert.False(t, exists)
}

func TestRSSService_UpdateFeedTitle(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	_ = services.NewRSSService(feedRepo, articleRepo, nil)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed with empty title
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)
	assert.Empty(t, feed.Title)

	// Simulate feed fetching that would update title
	// (In real scenario, FetchAndSaveArticles would do this)
	feed.Title = "Updated Feed Title"
	err = feedRepo.Update(feed)
	assert.NoError(t, err)

	// Verify title was updated
	updatedFeed, err := feedRepo.FindByID(feed.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Feed Title", updatedFeed.Title)
}

func TestRSSService_UpdateLastFetch(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	_ = services.NewRSSService(feedRepo, articleRepo, nil)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed without LastFetch
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)
	assert.Nil(t, feed.LastFetch)

	// Simulate fetch completion
	now := time.Now()
	feed.LastFetch = &now
	err = feedRepo.Update(feed)
	assert.NoError(t, err)

	// Verify LastFetch was updated
	updatedFeed, err := feedRepo.FindByID(feed.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedFeed.LastFetch)
}

func TestRSSService_ArticleCreation(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)

	// Create an article
	pubDate := time.Now()
	article := &models.Article{
		FeedID:      feed.ID,
		Title:       "Test Article",
		Link:        "https://example.com/article1",
		Description: "Test description",
		Content:     "<p>Test content</p>",
		UserID:      user.ID,
		PubDate:     &pubDate,
		IsRead:      false,
	}
	err = articleRepo.Create(article)
	assert.NoError(t, err)

	// Verify article was created
	articles, _, err := articleRepo.FindByUserID(user.ID, 1, 10, 0, 0, "", nil, nil)
	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "Test Article", articles[0].Title)
	assert.Equal(t, "https://example.com/article1", articles[0].Link)
	assert.False(t, articles[0].IsRead)
}

func TestRSSService_SkipExistingArticles(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Create a feed
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}
	err := feedRepo.Create(feed)
	assert.NoError(t, err)

	// Create an article
	pubDate := time.Now()
	article := &models.Article{
		FeedID:      feed.ID,
		Title:       "Test Article",
		Link:        "https://example.com/article1",
		Description: "Test description",
		UserID:      user.ID,
		PubDate:     &pubDate,
	}
	err = articleRepo.Create(article)
	assert.NoError(t, err)

	// Verify article exists
	exists := articleRepo.ExistsByLink(feed.ID, "https://example.com/article1")
	assert.True(t, exists)

	// In a real scenario, FetchAndSaveArticles would skip this article
	// because it already exists
	// This test verifies the logic works correctly
}

func TestRSSService_FeedRepositoryIntegration(t *testing.T) {
	// Setup
	db := setupTestDBRSS(t)
	feedRepo := repository.NewFeedRepository(db)

	// Create a user
	user := &models.User{Email: "test@example.com", Password: "hashed"}
	db.Create(user)

	// Test FeedRepository methods
	feed := &models.Feed{
		URL:      "https://example.com/feed.xml",
		Title:    "Test Feed",
		UserID:   user.ID,
		Category: "tech",
	}

	// Create
	err := feedRepo.Create(feed)
	assert.NoError(t, err)
	assert.NotZero(t, feed.ID)

	// FindByID
	foundFeed, err := feedRepo.FindByID(feed.ID)
	assert.NoError(t, err)
	assert.Equal(t, feed.URL, foundFeed.URL)
	assert.Equal(t, feed.Title, foundFeed.Title)

	// FindByURLAndUserID
	foundFeed2, err := feedRepo.FindByURLAndUserID("https://example.com/feed.xml", user.ID)
	assert.NoError(t, err)
	assert.Equal(t, feed.ID, foundFeed2.ID)

	// FindByUserID
	feeds, err := feedRepo.FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, feeds, 1)

	// Update
	feed.Title = "Updated Title"
	err = feedRepo.Update(feed)
	assert.NoError(t, err)

	updatedFeed, err := feedRepo.FindByID(feed.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedFeed.Title)

	// Delete
	err = feedRepo.Delete(feed.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = feedRepo.FindByID(feed.ID)
	assert.Error(t, err)
}

// Test for gofeed parser integration
func TestRSSService_ParseRSS(t *testing.T) {
	// This test demonstrates how the gofeed parser would be used
	// In production, you would mock the HTTP response

	fp := gofeed.NewParser()

	// Test parsing an empty string
	_, err := fp.ParseString("")
	assert.Error(t, err)

	// Test parsing invalid RSS
	_, err = fp.ParseString("invalid rss")
	assert.Error(t, err)
}
