package repository_test

import (
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"testing"
)

func TestTagRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")

	tag := &models.Tag{
		Name:   "AI",
		UserID: user.ID,
	}

	err := repo.Create(tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.ID == 0 {
		t.Error("Tag ID should not be 0")
	}

	if tag.Name != "AI" {
		t.Errorf("Expected name 'AI', got '%s'", tag.Name)
	}
}

func TestTagRepository_Create_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")

	// 创建第一个标签
	tag1 := &models.Tag{
		Name:   "AI",
		UserID: user.ID,
	}
	err := repo.Create(tag1)
	if err != nil {
		t.Fatalf("Failed to create first tag: %v", err)
	}

	// 尝试创建相同名称的标签（应该失败）
	tag2 := &models.Tag{
		Name:   "AI",
		UserID: user.ID,
	}
	err = repo.Create(tag2)
	if err == nil {
		t.Error("Expected error when creating tag with duplicate name")
	}
}

func TestTagRepository_FindByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")

	// 创建多个标签
	tags := []*models.Tag{
		{Name: "AI", UserID: user.ID},
		{Name: "Startup", UserID: user.ID},
		{Name: "News", UserID: user.ID},
	}

	for _, tag := range tags {
		repo.Create(tag)
	}

	// 查找用户的所有标签
	foundTags, err := repo.FindByUserID(user.ID)
	if err != nil {
		t.Fatalf("Failed to find tags by user ID: %v", err)
	}

	if len(foundTags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(foundTags))
	}
}

func TestTagRepository_FindByUserID_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")

	// 查找用户的标签（应该为空）
	tags, err := repo.FindByUserID(user.ID)
	if err != nil {
		t.Fatalf("Failed to find tags by user ID: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}

func TestTagRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")
	createdTag := createTestTag(db, user.ID, "AI")

	// 查找标签
	foundTag, err := repo.FindByID(createdTag.ID)
	if err != nil {
		t.Fatalf("Failed to find tag by ID: %v", err)
	}

	if foundTag.ID != createdTag.ID {
		t.Errorf("Expected tag ID %d, got %d", createdTag.ID, foundTag.ID)
	}

	if foundTag.Name != "AI" {
		t.Errorf("Expected name 'AI', got '%s'", foundTag.Name)
	}
}

func TestTagRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	// 查找不存在的标签
	_, err := repo.FindByID(99999)
	if err == nil {
		t.Error("Expected error when finding non-existent tag")
	}
}

func TestTagRepository_FindByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")
	createdTag := createTestTag(db, user.ID, "AI")

	// 按名称查找标签
	foundTag, err := repo.FindByName(user.ID, "AI")
	if err != nil {
		t.Fatalf("Failed to find tag by name: %v", err)
	}

	if foundTag.ID != createdTag.ID {
		t.Errorf("Expected tag ID %d, got %d", createdTag.ID, foundTag.ID)
	}

	if foundTag.Name != "AI" {
		t.Errorf("Expected name 'AI', got '%s'", foundTag.Name)
	}
}

func TestTagRepository_FindByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")

	// 查找不存在的标签
	_, err := repo.FindByName(user.ID, "Nonexistent")
	if err == nil {
		t.Error("Expected error when finding non-existent tag")
	}
}

func TestTagRepository_FindByName_DifferentUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user1 := createTestUser(db, "user1@example.com")
	user2 := createTestUser(db, "user2@example.com")

	// user1 创建标签
	createTestTag(db, user1.ID, "AI")

	// user2 尝试查找同名标签（应该找不到）
	_, err := repo.FindByName(user2.ID, "AI")
	if err == nil {
		t.Error("Expected error when finding tag for different user")
	}
}

func TestTagRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user := createTestUser(db, "test@example.com")
	tag := createTestTag(db, user.ID, "AI")

	// 删除标签
	err := repo.Delete(tag.ID)
	if err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}

	// 验证删除
	_, err = repo.FindByID(tag.ID)
	if err == nil {
		t.Error("Expected error when finding deleted tag")
	}
}

func TestTagRepository_Delete_CascadeArticleTags(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := repository.NewTagRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	user := createTestUser(db, "test@example.com")
	feed := createTestFeed(db, user.ID, "Test Feed", "https://example.com/feed", "tech")
	article := createTestArticle(db, feed.ID, user.ID, "Test Article", "https://example.com/article")
	tag := createTestTag(db, user.ID, "AI")

	// 添加标签到文章
	articleRepo.AddTag(article.ID, tag.ID)

	// 验证标签已添加
	var countBefore int64
	db.Table("article_tags").Where("tag_id = ?", tag.ID).Count(&countBefore)
	if countBefore != 1 {
		t.Error("Expected 1 article-tag association before deletion")
	}

	// 删除标签
	err := tagRepo.Delete(tag.ID)
	if err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}

	// 验证文章-标签关联也被删除
	var countAfter int64
	db.Table("article_tags").Where("tag_id = ?", tag.ID).Count(&countAfter)
	if countAfter != 0 {
		t.Error("Expected article-tag associations to be deleted")
	}
}

func TestTagRepository_MultipleUsers_SameName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTagRepository(db)

	user1 := createTestUser(db, "user1@example.com")
	user2 := createTestUser(db, "user2@example.com")

	// 两个用户创建相同名称的标签（应该允许）
	tag1 := &models.Tag{Name: "AI", UserID: user1.ID}
	tag2 := &models.Tag{Name: "AI", UserID: user2.ID}

	err1 := repo.Create(tag1)
	err2 := repo.Create(tag2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create tags for different users: %v, %v", err1, err2)
	}

	// 验证两个标签都存在
	tags1, _ := repo.FindByUserID(user1.ID)
	tags2, _ := repo.FindByUserID(user2.ID)

	if len(tags1) != 1 || len(tags2) != 1 {
		t.Error("Expected 1 tag per user")
	}
}
