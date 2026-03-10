package repository

import (
	"rss-reader/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type FeedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) Create(feed *models.Feed) error {
	return r.db.Create(feed).Error
}

func (r *FeedRepository) FindByUserID(userID uint) ([]models.Feed, error) {
	var feeds []models.Feed
	if err := r.db.Where("user_id = ?", userID).Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *FeedRepository) FindByID(id uint) (*models.Feed, error) {
	var feed models.Feed
	if err := r.db.First(&feed, id).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *FeedRepository) Update(feed *models.Feed) error {
	return r.db.Save(feed).Error
}

func (r *FeedRepository) Delete(id uint) error {
	// Delete related articles first
	if err := r.db.Where("feed_id = ?", id).Delete(&models.Article{}).Error; err != nil {
		return err
	}
	// Then delete the feed
	return r.db.Delete(&models.Feed{}, id).Error
}

func (r *FeedRepository) FindAll() ([]models.Feed, error) {
	var feeds []models.Feed
	if err := r.db.Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *FeedRepository) FindByURLAndUserID(url string, userID uint) (*models.Feed, error) {
	var feed models.Feed
	if err := r.db.Where("url = ? AND user_id = ?", url, userID).First(&feed).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

// FindByUserID 支持按 feed_id, tag_id, category, is_read 筛选
func (r *ArticleRepository) FindByUserID(userID uint, page, pageSize int, feedID uint, tagID uint, category string, isRead *bool, hasSummary *bool) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).Where("articles.user_id = ?", userID)

	// Filter by feed_id
	if feedID > 0 {
		query = query.Where("articles.feed_id = ?", feedID)
	}

	// Filter by tag_id
	if tagID > 0 {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tagID)
	}

	// Filter by category
	if category != "" {
		query = query.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.category = ?", category)
	}

	// Filter by is_read
	if isRead != nil {
		query = query.Where("articles.is_read = ?", *isRead)
	}

	// Filter by has_summary
	if hasSummary != nil {
		if *hasSummary {
			query = query.Where("articles.summary != ?", "")
		} else {
			query = query.Where("articles.summary = ?", "")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("articles.pub_date DESC").Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	// Fetch feed info
	for i := range articles {
		var feed models.Feed
		if err := r.db.First(&feed, articles[i].FeedID).Error; err == nil {
			articles[i].Feed = &feed
		}
		// Fetch tags
		r.db.Model(&articles[i]).Association("Tags").Find(&articles[i].Tags)
	}

	return articles, total, nil
}

func (r *ArticleRepository) SearchByTitle(userID uint, query string) ([]models.Article, error) {
	var articles []models.Article
	if err := r.db.Where("user_id = ? AND title ILIKE ?", userID, "%"+query+"%").
		Order("pub_date DESC").Find(&articles).Error; err != nil {
		return nil, err
	}

	for i := range articles {
		var feed models.Feed
		if err := r.db.First(&feed, articles[i].FeedID).Error; err == nil {
			articles[i].Feed = &feed
		}
	}

	return articles, nil
}

func (r *ArticleRepository) ExistsByLink(feedID uint, link string) bool {
	var count int64
	r.db.Model(&models.Article{}).Where("feed_id = ? AND link = ?", feedID, link).Count(&count)
	return count > 0
}

func (r *ArticleRepository) AddTag(articleID, tagID uint) error {
	return r.db.Exec("INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING", articleID, tagID).Error
}

func (r *ArticleRepository) RemoveTag(articleID, tagID uint) error {
	return r.db.Exec("DELETE FROM article_tags WHERE article_id = ? AND tag_id = ?", articleID, tagID).Error
}

// MarkAsRead 标记单篇文章为已读
func (r *ArticleRepository) MarkAsRead(userID, articleID uint) error {
	return r.db.Model(&models.Article{}).
		Where("id = ? AND user_id = ?", articleID, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead 批量标记已读，支持按 feed_id 和 category 筛选
func (r *ArticleRepository) MarkAllAsRead(userID uint, feedID uint, category string) (int64, error) {
	query := r.db.Model(&models.Article{}).Where("user_id = ? AND is_read = ?", userID, false)

	if feedID > 0 {
		query = query.Where("feed_id = ?", feedID)
	}

	if category != "" {
		query = query.Joins("JOIN feeds ON feeds.id = articles.feed_id").
			Where("feeds.category = ?", category)
	}

	result := query.Update("is_read", true)
	return result.RowsAffected, result.Error
}

// GetUnreadCount 获取未读数量统计
func (r *ArticleRepository) GetUnreadCount(userID uint) (int64, map[uint]int64, map[string]int64, error) {
	var total int64
	r.db.Model(&models.Article{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&total)

	// By feed
	byFeed := make(map[uint]int64)
	var feedCounts []struct {
		FeedID uint
		Count  int64
	}
	r.db.Model(&models.Article{}).
		Select("feed_id, count(*) as count").
		Where("user_id = ? AND is_read = ?", userID, false).
		Group("feed_id").
		Scan(&feedCounts)
	for _, fc := range feedCounts {
		byFeed[fc.FeedID] = fc.Count
	}

	// By category
	byCategory := make(map[string]int64)
	var catCounts []struct {
		Category string
		Count    int64
	}
	r.db.Model(&models.Article{}).
		Select("feeds.category, count(*) as count").
		Joins("JOIN feeds ON feeds.id = articles.feed_id").
		Where("articles.user_id = ? AND articles.is_read = ?", userID, false).
		Group("feeds.category").
		Scan(&catCounts)
	for _, cc := range catCounts {
		byCategory[cc.Category] = cc.Count
	}

	return total, byFeed, byCategory, nil
}

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepository) FindByUserID(userID uint) ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.db.Where("user_id = ?", userID).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *TagRepository) FindByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) FindByName(userID uint, name string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tag{}, id).Error
}

// UpdateArticleSummary 更新文章的 AI 摘要
func (r *ArticleRepository) UpdateArticleSummary(articleID uint, summary, keyPoints string) error {
	return r.db.Model(&models.Article{}).
		Where("id = ?", articleID).
		Updates(map[string]interface{}{
			"summary":    summary,
			"key_points": keyPoints,
		}).Error
}

// GetArticleByID 根据 ID 获取文章
func (r *ArticleRepository) GetArticleByID(articleID uint) (*models.Article, error) {
	var article models.Article
	if err := r.db.First(&article, articleID).Error; err != nil {
		return nil, err
	}
	return &article, nil
}
