package repository_test

import (
	"rss-reader/internal/repository"
	"testing"
)

func TestArticleRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	article := createTestArticleNoDB(feed.ID, user.ID, "Test Article", "https://example.com/article")

	err := repo.Create(article)
	if err != nil {
		t.Fatalf("Failed to create article: %v", err)
	}

	if article.ID == 0 {
		t.Error("Article ID should not be 0")
	}

	if article.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", article.Title)
	}
}

func TestArticleRepository_FindByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建文章
	createTestArticle(db, feed.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed.ID, user.ID, "Article 2", "https://example.com/article2")
	createTestArticle(db, feed.ID, user.ID, "Article 3", "https://example.com/article3")

	// 查找文章
	articles, total, err := repo.FindByUserID(user.ID, 1, 10, 0, 0, "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to find articles: %v", err)
	}

	if len(articles) != 3 {
		t.Errorf("Expected 3 articles, got %d", len(articles))
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}

func TestArticleRepository_FindByUserID_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建 5 篇文章
	for i := 1; i <= 5; i++ {
		createTestArticle(db, feed.ID, user.ID, "Article", "https://example.com/article")
	}

	// 第一页
	articles1, total, err := repo.FindByUserID(user.ID, 1, 2, 0, 0, "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to find articles (page 1): %v", err)
	}
	if len(articles1) != 2 {
		t.Errorf("Expected 2 articles on page 1, got %d", len(articles1))
	}

	// 第二页
	articles2, _, err := repo.FindByUserID(user.ID, 2, 2, 0, 0, "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to find articles (page 2): %v", err)
	}
	if len(articles2) != 2 {
		t.Errorf("Expected 2 articles on page 2, got %d", len(articles2))
	}

	// 验证总数
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
}

func TestArticleRepository_FindByUserID_FilterByFeedID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed1 := createTestFeed(db, user.ID, "Feed 1", "https://example.com/feed1", "tech")
	feed2 := createTestFeed(db, user.ID, "Feed 2", "https://example.com/feed2", "news")

	// 为 feed1 创建文章
	createTestArticle(db, feed1.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed1.ID, user.ID, "Article 2", "https://example.com/article2")

	// 为 feed2 创建文章
	createTestArticle(db, feed2.ID, user.ID, "Article 3", "https://example.com/article3")

	// 筛选 feed1 的文章
	articles, _, err := repo.FindByUserID(user.ID, 1, 10, feed1.ID, 0, "", nil, nil)
	if err != nil {
		t.Fatalf("Failed to find articles by feed ID: %v", err)
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles for feed1, got %d", len(articles))
	}
}

func TestArticleRepository_FindByUserID_FilterByIsRead(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建已读文章
	article1 := createTestArticle(db, feed.ID, user.ID, "Read Article", "https://example.com/article1")
	db.Model(&article1).Update("is_read", true)

	// 创建未读文章
	createTestArticle(db, feed.ID, user.ID, "Unread Article", "https://example.com/article2")

	// 筛选已读文章
	isRead := true
	articles, _, err := repo.FindByUserID(user.ID, 1, 10, 0, 0, "", &isRead, nil)
	if err != nil {
		t.Fatalf("Failed to find read articles: %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("Expected 1 read article, got %d", len(articles))
	}

	// 筛选未读文章
	isRead = false
	articles, _, err = repo.FindByUserID(user.ID, 1, 10, 0, 0, "", &isRead, nil)
	if err != nil {
		t.Fatalf("Failed to find unread articles: %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("Expected 1 unread article, got %d", len(articles))
	}
}

func TestArticleRepository_SearchByTitle(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建文章
	createTestArticle(db, feed.ID, user.ID, "AI Advances in 2026", "https://example.com/article1")
	createTestArticle(db, feed.ID, user.ID, "New Technology Trends", "https://example.com/article2")
	createTestArticle(db, feed.ID, user.ID, "Machine Learning Basics", "https://example.com/article3")

	// 搜索 "AI"
	articles, err := repo.SearchByTitle(user.ID, "AI")
	if err != nil {
		t.Fatalf("Failed to search articles: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article matching 'AI', got %d", len(articles))
	}

	if articles[0].Title != "AI Advances in 2026" {
		t.Errorf("Expected title 'AI Advances in 2026', got '%s'", articles[0].Title)
	}

	// 搜索空关键词（应该返回所有文章）
	articles, err = repo.SearchByTitle(user.ID, "")
	if err != nil {
		t.Fatalf("Failed to search articles with empty query: %v", err)
	}
	if len(articles) != 3 {
		t.Errorf("Expected 3 articles with empty query, got %d", len(articles))
	}
}

func TestArticleRepository_ExistsByLink(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建文章
	createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")

	// 验证存在
	if !repo.ExistsByLink(feed.ID, "https://example.com/article") {
		t.Error("Expected article to exist")
	}

	// 验证不存在
	if repo.ExistsByLink(feed.ID, "https://example.com/nonexistent") {
		t.Error("Expected article to not exist")
	}
}

func TestArticleRepository_MarkAsRead(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")
	article := createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")

	// 标记为已读
	err := repo.MarkAsRead(user.ID, article.ID)
	if err != nil {
		t.Fatalf("Failed to mark article as read: %v", err)
	}

	// 验证
	articles, _, _ := repo.FindByUserID(user.ID, 1, 10, 0, 0, "", nil, nil)
	if !articles[0].IsRead {
		t.Error("Expected article to be marked as read")
	}
}

func TestArticleRepository_MarkAllAsRead(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")

	// 创建 3 篇未读文章
	createTestArticle(db, feed.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed.ID, user.ID, "Article 2", "https://example.com/article2")
	createTestArticle(db, feed.ID, user.ID, "Article 3", "https://example.com/article3")

	// 批量标记为已读
	affected, err := repo.MarkAllAsRead(user.ID, 0, "")
	if err != nil {
		t.Fatalf("Failed to mark all as read: %v", err)
	}

	if affected != 3 {
		t.Errorf("Expected 3 articles to be marked as read, got %d", affected)
	}

	// 验证
	articles, _, _ := repo.FindByUserID(user.ID, 1, 10, 0, 0, "", nil, nil)
	for _, article := range articles {
		if !article.IsRead {
			t.Error("Expected all articles to be marked as read")
		}
	}
}

func TestArticleRepository_MarkAllAsRead_FilterByFeedID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed1 := createTestFeed(db, user.ID, "Feed 1", "https://example.com/feed1", "tech")
	feed2 := createTestFeed(db, user.ID, "Feed 2", "https://example.com/feed2", "news")

	// 为两个 feed 创建文章
	createTestArticle(db, feed1.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed1.ID, user.ID, "Article 2", "https://example.com/article2")
	createTestArticle(db, feed2.ID, user.ID, "Article 3", "https://example.com/article3")

	// 只标记 feed1 的文章
	affected, err := repo.MarkAllAsRead(user.ID, feed1.ID, "")
	if err != nil {
		t.Fatalf("Failed to mark all as read for feed1: %v", err)
	}

	if affected != 2 {
		t.Errorf("Expected 2 articles to be marked as read, got %d", affected)
	}
}

func TestArticleRepository_GetUnreadCount(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed1 := createTestFeed(db, user.ID, "Feed 1", "https://example.com/feed1", "tech")
	feed2 := createTestFeed(db, user.ID, "Feed 2", "https://example.com/feed2", "news")

	// 创建文章（未读）
	createTestArticle(db, feed1.ID, user.ID, "Article 1", "https://example.com/article1")
	createTestArticle(db, feed1.ID, user.ID, "Article 2", "https://example.com/article2")
	createTestArticle(db, feed2.ID, user.ID, "Article 3", "https://example.com/article3")

	// 标记一篇为已读
	article4 := createTestArticle(db, feed1.ID, user.ID, "Article 4", "https://example.com/article4")
	db.Model(&article4).Update("is_read", true)

	// 获取未读数量
	total, byFeed, byCategory, err := repo.GetUnreadCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get unread count: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 unread articles, got %d", total)
	}

	if byFeed[feed1.ID] != 2 {
		t.Errorf("Expected 2 unread articles for feed1, got %d", byFeed[feed1.ID])
	}

	if byFeed[feed2.ID] != 1 {
		t.Errorf("Expected 1 unread article for feed2, got %d", byFeed[feed2.ID])
	}

	if byCategory["tech"] != 2 {
		t.Errorf("Expected 2 unread articles in 'tech' category, got %d", byCategory["tech"])
	}

	if byCategory["news"] != 1 {
		t.Errorf("Expected 1 unread article in 'news' category, got %d", byCategory["news"])
	}
}

func TestArticleRepository_AddTag(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")
	article := createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")
	tag := createTestTag(db, user.ID, "AI")

	// 添加标签
	err := repo.AddTag(article.ID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to add tag to article: %v", err)
	}

	// 验证标签已添加
	var count int64
	db.Table("article_tags").Where("article_id = ? AND tag_id = ?", article.ID, tag.ID).Count(&count)
	if count != 1 {
		t.Error("Expected tag to be added to article")
	}
}

func TestArticleRepository_RemoveTag(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")
	article := createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")
	tag := createTestTag(db, user.ID, "AI")

	// 添加标签
	repo.AddTag(article.ID, tag.ID)

	// 移除标签
	err := repo.RemoveTag(article.ID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to remove tag from article: %v", err)
	}

	// 验证标签已移除
	var count int64
	db.Table("article_tags").Where("article_id = ? AND tag_id = ?", article.ID, tag.ID).Count(&count)
	if count != 0 {
		t.Error("Expected tag to be removed from article")
	}
}

func TestArticleRepository_UpdateArticleSummary(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")
	article := createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")

	// 更新摘要
	summary := "This is a generated summary"
	keyPoints := "Point 1\nPoint 2\nPoint 3"

	err := repo.UpdateArticleSummary(article.ID, summary, keyPoints)
	if err != nil {
		t.Fatalf("Failed to update article summary: %v", err)
	}

	// 验证
	updatedArticle, err := repo.GetArticleByID(article.ID)
	if err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}

	if updatedArticle.Summary != summary {
		t.Errorf("Expected summary '%s', got '%s'", summary, updatedArticle.Summary)
	}

	if updatedArticle.KeyPoints != keyPoints {
		t.Errorf("Expected key points '%s', got '%s'", keyPoints, updatedArticle.KeyPoints)
	}
}
